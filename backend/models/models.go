package models

import (
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Address   string    `json:"address" gorm:"unique"`
	CreatedAt time.Time `json:"created_at"`
}

type StakeRecord struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	UserAddress     string    `json:"user_address"`
	TokenAddress    string    `json:"token_address"`
	StockSymbol     string    `json:"stock_symbol"`
	Chain           string    `json:"chain"`
	ContractAddress string    `json:"contract_address"`
	Amount          string    `json:"amount"`
	SCOSBorrowed    string    `json:"scos_borrowed"`
	Status          string    `json:"status"` // active, redeemed, liquidated
	StakePrice      string    `json:"stake_price"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type TokenPrice struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Symbol    string    `json:"symbol" gorm:"unique"`
	Price     float64   `json:"price"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserAddress string    `json:"user_address"`
	Type        string    `json:"type"` // stake, unstake, buy, sell, liquidate
	TxHash      string    `json:"tx_hash"`
	Chain       string    `json:"chain"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
