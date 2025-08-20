package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	ethRPC     = "https://reddio-dev.reddio.com/"
	privateKey = "01c7939dc6827ee10bb7d26f420618c4af88c0029aa70be202f1ca7f29fe5bb4"

	StockSellerAddr = "0xf30e1edec3c1633d3b4b67b9c37a597a95d808ff"
)

var ReddioChainID = big.NewInt(50341)

// ============ 通用请求工具 ============
func doRequest(t *testing.T, method, url string, body interface{}) ([]byte, int) {
	t.Helper()

	var req *http.Request
	var err error
	if body != nil {
		b, _ := json.Marshal(body)
		req, err = http.NewRequest(method, url, bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	return data, resp.StatusCode
}

// ============ ERC20 Mint ============
func mintAppleToken(t *testing.T, to string, amount int64) {
	client, err := ethclient.Dial(ethRPC)
	if err != nil {
		t.Fatalf("连接以太坊失败: %v", err)
	}

	erc20ABI := `[{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"name":"mint","outputs":[],"type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	assert.NoError(t, err)

	contract := bind.NewBoundContract(
		common.HexToAddress(APPLEtokenAddr),
		parsedABI,
		client, client, client,
	)

	pk, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
	assert.NoError(t, err)

	// 调用 mint
	auth, _ := bind.NewKeyedTransactorWithChainID(pk, ReddioChainID)
	tx, err := contract.Transact(auth, "mint", common.HexToAddress(to), big.NewInt(amount))
	assert.NoError(t, err)

	log.Printf("Mint交易已发送: %s 金额: %d", tx.Hash().Hex(), amount)
}

func transferApple(t *testing.T, to string, amount int64) {
	client, err := ethclient.Dial(ethRPC)
	if err != nil {
		t.Fatalf("连接以太坊失败: %v", err)
	}

	erc20ABI := `[{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"name":"transfer","outputs":[],"type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	assert.NoError(t, err)

	contract := bind.NewBoundContract(
		common.HexToAddress(SCOStokenAddr),
		parsedABI,
		client, client, client,
	)
	pk, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
	assert.NoError(t, err)

	// 调用 transfer
	auth, _ := bind.NewKeyedTransactorWithChainID(pk, ReddioChainID)
	tx, err := contract.Transact(auth, "transfer", common.HexToAddress(to), big.NewInt(amount))
	assert.NoError(t, err)

	log.Printf("Transfer交易已发送: %s", tx.Hash().Hex())
}

func cleanDB(t *testing.T) {
	assert.NoError(t, os.RemoveAll("*/**/scos.db"))
}

// ============ 场景测试 ============
func TestScenario_FullFlow(t *testing.T) {

	t.Log("========== 功能测试 ===========")
	cleanDB(t)

	stockAmount := "20"
	var (
		appleAmount     int64 = 10_000000
		sellAppleAmount       = "5_000000"
	)

	// 1. 给 user mint 10000000 APPLE token
	mintAppleToken(t, userAddr, appleAmount)
	mintAppleToken(t, StockSellerAddr, appleAmount)

	// 2. 质押 APPLE 20 个
	stakeBody := map[string]interface{}{
		"user_address":  userAddr,
		"token_address": SCOStokenAddr,
		"chain":         chainName,
		"stock_symbol":  symbol,
		"amount":        stockAmount,
	}
	data, code := doRequest(t, "POST", fmt.Sprintf("%s/api/stake", baseURL), stakeBody)
	if code != http.StatusOK {
		t.Fatalf("质押失败: %s", string(data))
	}
	t.Logf("质押成功: %s", string(data))

	// 3. 查询 SCOS 余额
	data, code = doRequest(t, "GET", fmt.Sprintf("%s/api/user/%s/scos", baseURL, userAddr), nil)
	if code != http.StatusOK {
		t.Fatalf("查询SCOS失败: %s", string(data))
	}
	t.Logf("SCOS余额: %s", string(data))

	// 用全数 SCOS 买 APPLE
	buyBody := map[string]interface{}{
		"user_address":  userAddr,
		"token_address": SCOStokenAddr,
		"chain":         chainName,
		"amount":        sellAppleAmount,
	}
	data, code = doRequest(t, "POST", fmt.Sprintf("%s/api/buy", baseURL), buyBody)
	if code != http.StatusOK {
		t.Fatalf("买入失败: %s", string(data))
	}
	t.Logf("买入成功: %s", string(data))

	// 卖出 APPLE
	sellBody := map[string]interface{}{
		"user_address":  userAddr,
		"token_address": SCOStokenAddr,
		"chain":         chainName,
		"amount":        sellAppleAmount,
	}
	data, code = doRequest(t, "POST", fmt.Sprintf("%s/api/sell", baseURL), sellBody)
	if code != http.StatusOK {
		t.Fatalf("卖出失败: %s", string(data))
	}
	t.Logf("卖出成功: %s", string(data))

	// 4. 赎回 APPLE
	redeemBody := map[string]interface{}{
		"user_address":  userAddr,
		"token_address": SCOStokenAddr,
		"chain":         chainName,
	}
	data, code = doRequest(t, "POST", fmt.Sprintf("%s/api/redeem", baseURL), redeemBody)
	if code != http.StatusOK {
		t.Fatalf("赎回失败: %s", string(data))
	}
	t.Logf("赎回成功: %s", string(data))

	// 5. 修改 APPLE 价格，下跌30%
	updateBody := map[string]interface{}{"price": 70.0} // 假设初始100
	data, code = doRequest(t, "POST", fmt.Sprintf("%s/api/stock/%s/price", baseURL, symbol), updateBody)
	if code != http.StatusOK {
		t.Fatalf("修改价格失败: %s", string(data))
	}
	t.Logf("价格更新成功: %s", string(data))

	// 6. 查看质押是否被清算
	data, code = doRequest(t, "GET", fmt.Sprintf("%s/api/stake/%s", baseURL, userAddr), nil)
	if code != http.StatusOK {
		t.Fatalf("查询质押失败: %s", string(data))
	}
	t.Logf("质押状态: %s", string(data))
}
