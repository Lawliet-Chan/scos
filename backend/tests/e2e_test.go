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
)

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
		common.HexToAddress(tokenAddr),
		parsedABI,
		client, client, client,
	)

	pk, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
	assert.NoError(t, err)

	// 调用 mint
	auth, _ := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(50341))
	tx, err := contract.Transact(auth, "mint", common.HexToAddress(to), big.NewInt(amount))
	assert.NoError(t, err)

	log.Printf("Mint交易已发送: %s", tx.Hash().Hex())
}

func cleanDB(t *testing.T) {
	assert.NoError(t, os.RemoveAll("*/**/scos.db"))
}

// ============ 场景测试 ============
func TestScenario_FullFlow(t *testing.T) {
	cleanDB(t)
	// 1. 给 user mint 100 APPLE token
	mintAppleToken(t, userAddr, 100)

	// 2. 质押 APPLE 20 个
	stakeBody := map[string]interface{}{
		"user_address":  userAddr,
		"token_address": tokenAddr,
		"chain":         chainName,
		"amount":        "20",
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
		"token_address": tokenAddr,
		"chain":         chainName,
		"amount":        "ALL", // ⚠️ 这里需要你API支持全额买入, 不然换成具体数额
	}
	data, code = doRequest(t, "POST", fmt.Sprintf("%s/api/buy", baseURL), buyBody)
	if code != http.StatusOK {
		t.Fatalf("买入失败: %s", string(data))
	}
	t.Logf("买入成功: %s", string(data))

	// 卖出 APPLE
	sellBody := map[string]interface{}{
		"user_address":  userAddr,
		"token_address": tokenAddr,
		"chain":         chainName,
		"amount":        "ALL",
	}
	data, code = doRequest(t, "POST", fmt.Sprintf("%s/api/sell", baseURL), sellBody)
	if code != http.StatusOK {
		t.Fatalf("卖出失败: %s", string(data))
	}
	t.Logf("卖出成功: %s", string(data))

	// 4. 赎回 APPLE
	redeemBody := map[string]interface{}{
		"user_address":  userAddr,
		"token_address": tokenAddr,
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
