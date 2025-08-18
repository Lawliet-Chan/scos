package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"scos/models"
)

type StockHandler struct {
	db *gorm.DB
}

func NewStockHandler(db *gorm.DB) *StockHandler {
	return &StockHandler{db: db}
}

func (h *StockHandler) GetStockPrice(c *gin.Context) {
	symbol := c.Param("symbol")

	var price models.TokenPrice
	if err := h.db.Where("symbol = ?", symbol).First(&price).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":     price.Symbol,
		"price":      price.Price,
		"updated_at": price.UpdatedAt,
	})
}

func (h *StockHandler) GetAllStockPrices(c *gin.Context) {
	var prices []models.TokenPrice
	if err := h.db.Find(&prices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stocks": prices,
	})
}

func (h *StockHandler) UpdateStockPrice(c *gin.Context) {
	symbol := c.Param("symbol")

	var req struct {
		Price float64 `json:"price" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var price models.TokenPrice
	err := h.db.Where("symbol = ?", symbol).First(&price).Error
	if err == gorm.ErrRecordNotFound {
		price = models.TokenPrice{
			Symbol:    symbol,
			Price:     req.Price,
			UpdatedAt: time.Now(),
		}
		h.db.Create(&price)
	} else {
		price.Price = req.Price
		price.UpdatedAt = time.Now()
		h.db.Save(&price)
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":     price.Symbol,
		"price":      price.Price,
		"updated_at": price.UpdatedAt,
	})
}

func (h *StockHandler) GetUserSCOSBalance(c *gin.Context) {
	userAddress := c.Param("address")

	var stakes []models.StakeRecord
	if err := h.db.Where("user_address = ? AND status = ?", userAddress, "active").Find(&stakes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stakes"})
		return
	}

	totalSCOS := 0.0
	for _, stake := range stakes {
		scos, _ := strconv.ParseFloat(stake.SCOSBorrowed, 64)
		totalSCOS += scos
	}

	c.JSON(http.StatusOK, gin.H{
		"address":       userAddress,
		"scos_balance":  totalSCOS,
		"active_stakes": len(stakes),
	})
}
