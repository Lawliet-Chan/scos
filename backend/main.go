package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"scos/blockchain"
	"scos/config"
	"scos/handlers"
	"scos/models"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	db, err := gorm.Open(sqlite.Open("scos.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 自动迁移
	db.AutoMigrate(&models.User{}, &models.StakeRecord{}, &models.TokenPrice{}, &models.Transaction{})

	// 初始化区块链客户端
	bc, err := blockchain.NewBlockchainClient(cfg.Chains, cfg.PrivateKey)
	if err != nil {
		log.Fatal("Failed to create blockchain client:", err)
	}

	// 初始化处理器
	stockHandler := handlers.NewStockHandler(db)
	stakingHandler := handlers.NewStakingHandler(db, bc)
	tradingHandler := handlers.NewTradingHandler(db, bc)

	// 设置Gin
	r := gin.Default()

	// CORS配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API路由
	api := r.Group("/api")
	{
		// Stock价格相关
		api.GET("/stock/:symbol/price", stockHandler.GetStockPrice)
		api.GET("/stocks/prices", stockHandler.GetAllStockPrices)
		api.POST("/stock/:symbol/price", stockHandler.UpdateStockPrice)
		api.GET("/user/:address/scos", stockHandler.GetUserSCOSBalance)

		// 质押相关
		api.POST("/stake", stakingHandler.StakeStock)
		api.POST("/redeem", stakingHandler.RedeemStock)

		// 交易相关
		api.POST("/buy", tradingHandler.BuyStock)
		api.POST("/sell", tradingHandler.SellStock)
	}

	// 初始化一些测试数据
	initTestData(db)

	// 启动清算监控
	go startLiquidationMonitor(db, bc)

	log.Printf("Server starting on port %s", cfg.Port)
	r.Run(":" + cfg.Port)
}

func initTestData(db *gorm.DB) {
	// 初始化三个股票价格
	stocks := []models.TokenPrice{
		{Symbol: "APPLE", Price: 3000.0, UpdatedAt: time.Now()},
		{Symbol: "GOOGLE", Price: 3000.0, UpdatedAt: time.Now()},
		{Symbol: "MICROSOFT", Price: 3000.0, UpdatedAt: time.Now()},
	}

	for _, stock := range stocks {
		var existingStock models.TokenPrice
		if err := db.Where("symbol = ?", stock.Symbol).First(&existingStock).Error; err == gorm.ErrRecordNotFound {
			db.Create(&stock)
			fmt.Printf("Initialized %s stock price: $%.2f\n", stock.Symbol, stock.Price)
		}
	}
}

func startLiquidationMonitor(db *gorm.DB, bc *blockchain.BlockchainClient) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		checkLiquidations(db, bc)
	}
}

func checkLiquidations(db *gorm.DB, bc *blockchain.BlockchainClient) {
	var stakes []models.StakeRecord
	db.Where("status = ?", "active").Find(&stakes)

	// 获取所有当前股票价格
	var currentPrices []models.TokenPrice
	db.Find(&currentPrices)

	priceMap := make(map[string]float64)
	for _, price := range currentPrices {
		priceMap[price.Symbol] = price.Price
	}

	for _, stake := range stakes {
		// 解析质押时的价格和当前价格
		stakePrice := 0.0
		if _, err := fmt.Sscanf(stake.StakePrice, "%f", &stakePrice); err != nil {
			continue
		}

		// 根据stake记录中的TokenAddress确定股票类型
		// 这里需要一个映射关系，或者在StakeRecord中添加Symbol字段
		currentPrice := 3000.0 // 默认价格，实际应该根据token地址查询

		// 检查是否跌价超过25%
		priceDropRatio := (stakePrice - currentPrice) / stakePrice
		if priceDropRatio > 0.25 {
			// 执行清算
			log.Printf("Liquidating stake %d due to price drop: %.2f%%", stake.ID, priceDropRatio*100)

			txHash, err := bc.Liquidate(stake.Chain, stake.ContractAddress, stake.UserAddress, stake.TokenAddress)
			if err != nil {
				log.Printf("Failed to liquidate stake %d: %v", stake.ID, err)
				continue
			}

			stake.Status = "liquidated"
			stake.UpdatedAt = time.Now()
			db.Save(&stake)

			// 记录清算交易
			tx := models.Transaction{
				UserAddress: stake.UserAddress,
				Type:        "liquidate",
				TxHash:      txHash,
				Chain:       stake.Chain,
				Status:      "completed",
				CreatedAt:   time.Now(),
			}
			db.Create(&tx)
		}
	}
}
