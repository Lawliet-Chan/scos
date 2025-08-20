package handlers

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"scos/blockchain"
	"scos/models"
)

type StakingHandler struct {
	db *gorm.DB
	bc *blockchain.BlockchainClient
}

func NewStakingHandler(db *gorm.DB, bc *blockchain.BlockchainClient) *StakingHandler {
	return &StakingHandler{db: db, bc: bc}
}

type StakeRequest struct {
	UserAddress     string `json:"user_address" binding:"required"`
	TokenAddress    string `json:"token_address" binding:"required"`
	Chain           string `json:"chain" binding:"required"`
	Amount          string `json:"amount" binding:"required"`
	StockSymbol     string `json:"stock_symbol" binding:"required"`
	ContractAddress string `json:"contract_address" binding:"required"`
}

func (h *StakingHandler) StakeStock(c *gin.Context) {
	var req StakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取指定股票价格
	var price models.TokenPrice
	if err := h.db.Where("symbol = ?", req.StockSymbol).First(&price).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock price not found"})
		return
	}

	// 计算可借贷SCOS数量 (数量/140% x 价格)
	amount, _ := strconv.ParseFloat(req.Amount, 64)
	scosAmount := (amount / 1.4) * price.Price

	// 调用区块链合约
	amountWei := new(big.Int)
	amountWei.SetString(req.Amount+"000000", 10) // 假设6位小数

	scosAmountWei := new(big.Int)
	scosAmountWei.SetString(fmt.Sprintf("%.0f", scosAmount*1000000), 10)

	txHash, err := h.bc.StakeStock(req.Chain, req.TokenAddress, amountWei, scosAmountWei)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stake on blockchain"})
		return
	}

	// 记录到数据库
	stake := models.StakeRecord{
		UserAddress:     req.UserAddress,
		TokenAddress:    req.TokenAddress,
		StockSymbol:     req.StockSymbol,
		Chain:           req.Chain,
		ContractAddress: req.ContractAddress,
		Amount:          req.Amount,
		SCOSBorrowed:    fmt.Sprintf("%.6f", scosAmount),
		Status:          "active",
		StakePrice:      fmt.Sprintf("%.6f", price.Price),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.db.Create(&stake).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save stake record"})
		return
	}

	// 记录交易
	tx := models.Transaction{
		UserAddress: req.UserAddress,
		Type:        "stake",
		TxHash:      txHash,
		Chain:       req.Chain,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	h.db.Create(&tx)

	c.JSON(http.StatusOK, gin.H{
		"stake_id":      stake.ID,
		"scos_borrowed": scosAmount,
		"tx_hash":       txHash,
		"status":        "success",
	})
}

type RedeemRequest struct {
	UserAddress     string `json:"user_address" binding:"required"`
	TokenAddress    string `json:"token_address" binding:"required"`
	Chain           string `json:"chain" binding:"required"`
	ContractAddress string `json:"contract_address" binding:"required"`
}

func (h *StakingHandler) RedeemStock(c *gin.Context) {
	var req RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找活跃质押
	var stake models.StakeRecord
	if err := h.db.Where("user_address = ? AND token_address = ? AND chain = ? AND status = ?",
		req.UserAddress, req.TokenAddress, req.Chain, "active").First(&stake).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active stake found"})
		return
	}

	_, err := h.bc.UnstakeStock(req.Chain, req.TokenAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unstake on blockchain"})
		return
	}

	// 更新状态
	stake.Status = "redeemed"
	stake.UpdatedAt = time.Now()
	h.db.Save(&stake)

	c.JSON(http.StatusOK, gin.H{
		"stake_id": stake.ID,
		"status":   "redeemed",
		"message":  "Stock redeemed successfully",
	})
}
