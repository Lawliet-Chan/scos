package blockchain

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"scos/config"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockchainClient struct {
	clients    map[string]*ethclient.Client
	privateKey *ecdsa.PrivateKey
	auth       *bind.TransactOpts
}

func NewBlockchainClient(chainInfos map[string]config.ChainInfo, privateKeyHex string) (*BlockchainClient, error) {
	clients := make(map[string]*ethclient.Client)

	for chain, info := range chainInfos {
		client, err := ethclient.Dial(info.RPC)
		if err != nil {
			return nil, err
		}
		clients[chain] = client
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1))
	if err != nil {
		return nil, err
	}

	return &BlockchainClient{
		clients:    clients,
		privateKey: privateKey,
		auth:       auth,
	}, nil
}

func (bc *BlockchainClient) GetClient(chain string) *ethclient.Client {
	return bc.clients[chain]
}

func (bc *BlockchainClient) StakeStock(chain, contractAddr, tokenAddr string, amount *big.Int, scosAmount *big.Int) (string, error) {
	client := bc.GetClient(chain)

	contractABI := `[{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"scosAmount","type":"uint256"}],"name":"stakeStock","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return "", err
	}

	data, err := parsedABI.Pack("stakeStock", common.HexToAddress(tokenAddr), amount, scosAmount)
	if err != nil {
		return "", err
	}

	nonce, err := client.PendingNonceAt(context.Background(), bc.auth.From)
	if err != nil {
		return "", err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	msg := ethereum.CallMsg{
		From: bc.auth.From,
		To:   &common.Address{},
		Data: data,
	}
	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		gasLimit = 300000 // fallback
	}

	bc.auth.Nonce = big.NewInt(int64(nonce))
	bc.auth.Value = big.NewInt(0)
	bc.auth.GasLimit = gasLimit
	bc.auth.GasPrice = gasPrice

	boundContract := bind.NewBoundContract(common.HexToAddress(contractAddr), parsedABI, client, client, client)
	tx, err := boundContract.Transact(bc.auth, "stakeStock", common.HexToAddress(tokenAddr), amount, scosAmount)
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}
