package blockchain

import (
	"context"
	"crypto/ecdsa"
	"log"
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
	clients    map[string]*cli
	privateKey *ecdsa.PrivateKey
}

type cli struct {
	client       *ethclient.Client
	VaultAddress common.Address
}

func NewBlockchainClient(chains map[string]config.ChainInfo, privateKeyHex string) (*BlockchainClient, error) {
	clients := make(map[string]*cli)

	for chainName, chain := range chains {
		client, err := ethclient.Dial(chain.RPC)
		if err != nil {
			return nil, err
		}
		clients[chainName] = &cli{
			client:       client,
			VaultAddress: common.HexToAddress(chain.VaultAddress),
		}
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, err
	}

	return &BlockchainClient{
		clients:    clients,
		privateKey: privateKey,
	}, nil
}

func (bc *BlockchainClient) GetClient(chain string) *ethclient.Client {
	info := bc.clients[chain]
	if info == nil {
		log.Panicf("chain %s not found", chain)
	}
	return info.client
}

func (bc *BlockchainClient) GetVaultAddr(chain string) common.Address {
	return bc.clients[chain].VaultAddress
}

func (bc *BlockchainClient) auth(chain string) (*bind.TransactOpts, error) {
	chainID, err := bc.GetClient(chain).ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	return bind.NewKeyedTransactorWithChainID(bc.privateKey, chainID)

}

func (bc *BlockchainClient) StakeStock(chain, tokenAddr string, amount *big.Int, scosAmount *big.Int) (string, error) {
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

	auth, err := bc.auth(chain)
	if err != nil {
		return "", err
	}

	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return "", err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	msg := ethereum.CallMsg{
		From: auth.From,
		To:   &common.Address{},
		Data: data,
	}
	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		gasLimit = 300000 // fallback
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	boundContract := bind.NewBoundContract(bc.GetVaultAddr(chain), parsedABI, client, client, client)
	tx, err := boundContract.Transact(auth, "stakeStock", common.HexToAddress(tokenAddr), amount, scosAmount)
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

func (bc *BlockchainClient) UnstakeStock(chain, tokenAddr string) (string, error) {
	client := bc.GetClient(chain)

	contractABI := `[{"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"unstakeStock","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return "", err
	}

	data, err := parsedABI.Pack("unstakeStock", common.HexToAddress(tokenAddr))
	if err != nil {
		return "", err
	}

	auth, err := bc.auth(chain)
	if err != nil {
		return "", err
	}

	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return "", err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	msg := ethereum.CallMsg{
		From: auth.From,
		To:   &common.Address{},
		Data: data,
	}
	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		gasLimit = 300000 // fallback
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	boundContract := bind.NewBoundContract(bc.GetVaultAddr(chain), parsedABI, client, client, client)
	tx, err := boundContract.Transact(auth, "unstakeStock", common.HexToAddress(tokenAddr))
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

func (bc *BlockchainClient) Liquidate(chain, userAddr, tokenAddr string) (string, error) {
	client := bc.GetClient(chain)

	contractABI := `{
  "type": "function",
  "name": "liquidate",
  "inputs": [
    {
      "name": "user",
      "type": "address"
    },
    {
      "name": "token",
      "type": "address"
    }
  ],
  "outputs": [],
  "stateMutability": "nonpayable"
}`

	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return "", err
	}

	data, err := parsedABI.Pack("liquidate", common.HexToAddress(userAddr), common.HexToAddress(tokenAddr))
	if err != nil {
		return "", err
	}

	auth, err := bc.auth(chain)
	if err != nil {
		return "", err
	}

	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return "", err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	msg := ethereum.CallMsg{
		From: auth.From,
		To:   &common.Address{},
		Data: data,
	}
	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		gasLimit = 300000 // fallback
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	boundContract := bind.NewBoundContract(bc.GetVaultAddr(chain), parsedABI, client, client, client)
	tx, err := boundContract.Transact(auth, "liquidate", common.HexToAddress(userAddr), common.HexToAddress(tokenAddr))
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}
