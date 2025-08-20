package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"scos/blockchain"
	"scos/models"
)

type TradingHandler struct {
	db *gorm.DB
	bc *blockchain.BlockchainClient
}

func NewTradingHandler(db *gorm.DB, bc *blockchain.BlockchainClient) *TradingHandler {
	return &TradingHandler{db: db, bc: bc}
}

type BuyRequest struct {
	UserAddress  string `json:"user_address" binding:"required"`
	TokenAddress string `json:"token_address" binding:"required"`
	Chain        string `json:"chain" binding:"required"`
	Amount       string `json:"amount" binding:"required"`
}

func (h *TradingHandler) BuyStock(c *gin.Context) {
	var req BuyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 记录交易
	tx := models.Transaction{
		UserAddress: req.UserAddress,
		Type:        "buy",
		TxHash:      "pending",
		Chain:       req.Chain,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	h.db.Create(&tx)

	c.JSON(http.StatusOK, gin.H{
		"transaction_id": tx.ID,
		"status":         "success",
		"message":        "Buy order processed",
	})
}

type SellRequest struct {
	UserAddress  string `json:"user_address" binding:"required"`
	TokenAddress string `json:"token_address" binding:"required"`
	Chain        string `json:"chain" binding:"required"`
	Amount       string `json:"amount" binding:"required"`
}

func (h *TradingHandler) SellStock(c *gin.Context) {
	var req SellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 记录交易
	tx := models.Transaction{
		UserAddress: req.UserAddress,
		Type:        "sell",
		TxHash:      "pending",
		Chain:       req.Chain,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	h.db.Create(&tx)

	c.JSON(http.StatusOK, gin.H{
		"transaction_id": tx.ID,
		"status":         "success",
		"message":        "Sell order processed",
	})
}
