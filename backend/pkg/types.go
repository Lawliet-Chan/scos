package pkg

import "math/big"

// 配置结构
type Config struct {
	Port         string            `json:"port"`
	PrivateKey   string            `json:"private_key"`
	Networks     map[string]string `json:"networks"`
	ContractAddr string            `json:"contract_addr"`
	SCOSAddr     string            `json:"scos_addr"`
}

// 用户借贷记录
type UserLoan struct {
	User             string   `json:"user"`
	Token            string   `json:"token"`
	CollateralAmount *big.Int `json:"collateral_amount"`
	BorrowedAmount   *big.Int `json:"borrowed_amount"`
	TokenPrice       float64  `json:"token_price"`
	Network          string   `json:"network"`
	Timestamp        int64    `json:"timestamp"`
}

// API请求结构
type StakeRequest struct {
	Network     string `json:"network"`
	Token       string `json:"token"`
	Amount      string `json:"amount"`
	UserAddress string `json:"user_address"`
}

type BuyTokenRequest struct {
	Network     string `json:"network"`
	Token       string `json:"token"`
	Amount      string `json:"amount"`
	UserAddress string `json:"user_address"`
}

type RedeemRequest struct {
	Network     string `json:"network"`
	Token       string `json:"token"`
	UserAddress string `json:"user_address"`
}

type SetPriceRequest struct {
	Token string  `json:"token"`
	Price float64 `json:"price"`
}

// API响应结构
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 智能合约ABI（简化版）
const contractABI = `[
	{
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "amount", "type": "uint256"},
			{"name": "tokenPrice", "type": "uint256"}
		],
		"name": "deposit",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "token", "type": "address"}
		],
		"name": "withdraw",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "tokenAmount", "type": "uint256"},
			{"name": "tokenPrice", "type": "uint256"}
		],
		"name": "buyToken",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "user", "type": "address"},
			{"name": "token", "type": "address"},
			{"name": "currentPrice", "type": "uint256"}
		],
		"name": "liquidate",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "user", "type": "address"},
			{"name": "token", "type": "address"}
		],
		"name": "getUserPosition",
		"outputs": [
			{"name": "collateralAmount", "type": "uint256"},
			{"name": "borrowedAmount", "type": "uint256"},
			{"name": "lastPrice", "type": "uint256"},
			{"name": "isActive", "type": "bool"}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`

const scosABI = `[
	{
		"inputs": [
			{"name": "account", "type": "address"}
		],
		"name": "balanceOf",
		"outputs": [
			{"name": "", "type": "uint256"}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`
