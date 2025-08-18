package pkg

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Server struct {
	config       *Config
	priceManager *PriceManager
	clients      map[string]*ethclient.Client
	privateKey   *ecdsa.PrivateKey
	userLoans    map[string]*UserLoan
	loansMutex   sync.RWMutex
}

func NewServer() *Server {
	config := &Config{
		Port: "8080",
		Networks: map[string]string{
			"reddio": "https://reddio-dev.reddio.com/",
		},
		ContractAddr: "0x1234567890123456789012345678901234567890",
		SCOSAddr:     "0x0987654321098765432109876543210987654321",
	}

	server := &Server{
		config:       config,
		priceManager: NewPriceManager(),
		clients:      make(map[string]*ethclient.Client),
		userLoans:    make(map[string]*UserLoan),
	}

	// 初始化一些默认价格
	server.priceManager.SetPrice("Apple", 4000.0)
	server.priceManager.SetPrice("Google", 4300.0)
	server.priceManager.SetPrice("Microsoft", 300.0)

	go func() {
		for {
			server.priceManager.DecreasePrice("Apple", 100.0)
			time.Sleep(3 * time.Second)
			//server.priceManager.DecreasePrice("Apple", 100.0)
		}
	}()

	return server
}

func (s *Server) GetPort() string {
	return s.config.Port
}

func (s *Server) InitClients() error {
	for network, rpc := range s.config.Networks {
		client, err := ethclient.Dial(rpc)
		if err != nil {
			log.Printf("Failed to connect to %s: %v", network, err)
			continue
		}
		s.clients[network] = client
	}

	// 设置私钥（实际使用时应该从安全的地方获取）
	privateKey, err := crypto.HexToECDSA("07b00892ace244f4f9f2f3b43ec294d87ed113c33790d330e4c28e2ff1133d3d")
	if err != nil {
		return fmt.Errorf("invalid private key: %v", err)
	}
	s.privateKey = privateKey

	return nil
}

func (s *Server) SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// API路由
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/price/{token}", s.getTokenPrice).Methods("GET")
	api.HandleFunc("/price", s.setTokenPrice).Methods("POST")
	api.HandleFunc("/stake", s.stakeToken).Methods("POST")
	api.HandleFunc("/buy", s.buyToken).Methods("POST")
	api.HandleFunc("/redeem", s.redeemToken).Methods("POST")
	api.HandleFunc("/balance/{network}/{address}", s.getSCOSBalance).Methods("GET")
	api.HandleFunc("/position/{network}/{user}/{token}", s.getUserPosition).Methods("GET")

	// 静态文件服务
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	return r
}

func (s *Server) getTokenPrice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	price := s.priceManager.GetPrice(token)

	response := ApiResponse{
		Success: true,
		Data: map[string]float64{
			"price": price,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) setTokenPrice(w http.ResponseWriter, r *http.Request) {
	var req SetPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	s.priceManager.SetPrice(req.Token, req.Price)

	response := ApiResponse{
		Success: true,
		Message: "Price updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) stakeToken(w http.ResponseWriter, r *http.Request) {
	var req StakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 获取token价格
	price := s.priceManager.GetPrice(req.Token)
	if price == 0 {
		http.Error(w, "Token price not found", http.StatusBadRequest)
		return
	}

	client, exists := s.clients[req.Network]
	if !exists {
		http.Error(w, "Network not supported", http.StatusBadRequest)
		return
	}

	// 解析质押金额
	amount := new(big.Int)
	amount.SetString(req.Amount, 10)

	// 调用合约
	err := s.callContractDeposit(client, req.Token, amount, price)
	if err != nil {
		response := ApiResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to stake: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 计算借贷的SCOS数量
	priceWei := new(big.Int)
	priceWei.SetInt64(int64(price * 1e18))
	borrowedAmount := new(big.Int)
	borrowedAmount.Mul(amount, priceWei)
	borrowedAmount.Div(borrowedAmount, big.NewInt(140)) // 除以1.4

	// 记录用户借贷
	loanKey := fmt.Sprintf("%s_%s_%s", req.UserAddress, req.Network, req.Token)
	s.loansMutex.Lock()
	s.userLoans[loanKey] = &UserLoan{
		User:             req.UserAddress,
		Token:            req.Token,
		CollateralAmount: amount,
		BorrowedAmount:   borrowedAmount,
		TokenPrice:       price,
		Network:          req.Network,
		Timestamp:        time.Now().Unix(),
	}
	s.loansMutex.Unlock()

	response := ApiResponse{
		Success: true,
		Message: "Staking successful",
		Data: map[string]interface{}{
			"borrowed_scos": borrowedAmount.String(),
			"collateral":    amount.String(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) buyToken(w http.ResponseWriter, r *http.Request) {
	var req BuyTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	price := s.priceManager.GetPrice(req.Token)
	if price == 0 {
		http.Error(w, "Token price not found", http.StatusBadRequest)
		return
	}

	client, exists := s.clients[req.Network]
	if !exists {
		http.Error(w, "Network not supported", http.StatusBadRequest)
		return
	}

	amount := new(big.Int)
	amount.SetString(req.Amount, 10)

	err := s.callContractBuyToken(client, req.Token, amount, price)
	if err != nil {
		response := ApiResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to buy token: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ApiResponse{
		Success: true,
		Message: "Token purchase successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) redeemToken(w http.ResponseWriter, r *http.Request) {
	var req RedeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	client, exists := s.clients[req.Network]
	if !exists {
		http.Error(w, "Network not supported", http.StatusBadRequest)
		return
	}

	err := s.callContractWithdraw(client, req.Token)
	if err != nil {
		response := ApiResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to redeem: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// 删除用户借贷记录
	loanKey := fmt.Sprintf("%s_%s_%s", req.UserAddress, req.Network, req.Token)
	s.loansMutex.Lock()
	delete(s.userLoans, loanKey)
	s.loansMutex.Unlock()

	response := ApiResponse{
		Success: true,
		Message: "Token redeemed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) getSCOSBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	network := vars["network"]
	address := vars["address"]

	client, exists := s.clients[network]
	if !exists {
		http.Error(w, "Network not supported", http.StatusBadRequest)
		return
	}

	balance, err := s.getSCOSBalanceFromContract(client, address)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get balance: %v", err), http.StatusInternalServerError)
		return
	}

	response := ApiResponse{
		Success: true,
		Data: map[string]string{
			"balance": balance.String(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) getUserPosition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	network := vars["network"]
	user := vars["user"]
	token := vars["token"]

	loanKey := fmt.Sprintf("%s_%s_%s", user, network, token)
	s.loansMutex.RLock()
	loan, exists := s.userLoans[loanKey]
	s.loansMutex.RUnlock()

	if !exists {
		response := ApiResponse{
			Success: false,
			Message: "No position found",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ApiResponse{
		Success: true,
		Data:    loan,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 合约调用函数
func (s *Server) callContractDeposit(client *ethclient.Client, tokenAddr string, amount *big.Int, price float64) error {
	contractAddress := common.HexToAddress(s.config.ContractAddr)

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.privateKey, big.NewInt(1)) // 主网ID
	if err != nil {
		return err
	}

	priceWei := big.NewInt(int64(price * 1e18))

	contract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

	_, err = contract.Transact(auth, "deposit", common.HexToAddress(tokenAddr), amount, priceWei)
	return err
}

func (s *Server) callContractWithdraw(client *ethclient.Client, tokenAddr string) error {
	contractAddress := common.HexToAddress(s.config.ContractAddr)

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.privateKey, big.NewInt(1))
	if err != nil {
		return err
	}

	contract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

	_, err = contract.Transact(auth, "withdraw", common.HexToAddress(tokenAddr))
	return err
}

func (s *Server) callContractBuyToken(client *ethclient.Client, tokenAddr string, amount *big.Int, price float64) error {
	contractAddress := common.HexToAddress(s.config.ContractAddr)

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.privateKey, big.NewInt(1))
	if err != nil {
		return err
	}

	priceWei := big.NewInt(int64(price * 1e18))

	contract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

	_, err = contract.Transact(auth, "buyToken", common.HexToAddress(tokenAddr), amount, priceWei)
	return err
}

func (s *Server) getSCOSBalanceFromContract(client *ethclient.Client, address string) (*big.Int, error) {
	contractAddress := common.HexToAddress(s.config.SCOSAddr)

	parsedABI, err := abi.JSON(strings.NewReader(scosABI))
	if err != nil {
		return nil, err
	}

	contract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

	var result []interface{}
	err = contract.Call(&bind.CallOpts{}, &result, "balanceOf", common.HexToAddress(address))
	if err != nil {
		return nil, err
	}

	return result[0].(*big.Int), nil
}

// 价格监控和清算
func (s *Server) StartPriceMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkForLiquidations()
		}
	}
}

func (s *Server) checkForLiquidations() {
	s.loansMutex.RLock()
	loans := make([]*UserLoan, 0, len(s.userLoans))
	for _, loan := range s.userLoans {
		loans = append(loans, loan)
	}
	s.loansMutex.RUnlock()

	for _, loan := range loans {
		currentPrice := s.priceManager.GetPrice(loan.Token)
		if currentPrice == 0 {
			continue
		}

		// 检查是否需要清算（价格下跌超过25%）
		priceDropThreshold := loan.TokenPrice * 0.75
		if currentPrice <= priceDropThreshold {
			log.Printf("Liquidating position for user %s, token %s", loan.User, loan.Token)
			client := s.clients[loan.Network]
			if client != nil {
				err := s.liquidatePosition(client, loan.User, loan.Token, currentPrice)
				if err != nil {
					log.Printf("Failed to liquidate position: %v", err)
				} else {
					// 删除已清算的头寸
					loanKey := fmt.Sprintf("%s_%s_%s", loan.User, loan.Network, loan.Token)
					s.loansMutex.Lock()
					delete(s.userLoans, loanKey)
					s.loansMutex.Unlock()
				}
			}
		}
	}
}

func (s *Server) liquidatePosition(client *ethclient.Client, user, token string, currentPrice float64) error {
	contractAddress := common.HexToAddress(s.config.ContractAddr)

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.privateKey, big.NewInt(1))
	if err != nil {
		return err
	}

	priceWei := big.NewInt(int64(currentPrice * 1e18))

	contract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

	_, err = contract.Transact(auth, "liquidate", common.HexToAddress(user), common.HexToAddress(token), priceWei)
	return err
}
