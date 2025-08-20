package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	baseURL   = "http://localhost:8080" // ⚠️ 改成你的API服务地址
	userAddr  = "0x36a15f8d63742eaabf9ebb32a8551db13d6a3167"
	tokenAddr = "0xeB5e9Af4b798ec27A0f24DA22C7A7b3b657D05d9"
	symbol    = "APPLE"
	chainName = "Reddio"
)

// 通用请求函数
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

// 1. 获取Stock价格
func TestGetStockPrice(t *testing.T) {
	url := fmt.Sprintf("%s/api/stock/%s/price", baseURL, symbol)
	data, code := doRequest(t, "GET", url, nil)

	if code != http.StatusOK {
		t.Errorf("期望状态码200, 得到%d, 响应: %s", code, string(data))
	}
	t.Logf("获取Stock价格成功: %s", string(data))
}

// 2. 更新Stock价格 (管理员)
func TestUpdateStockPrice(t *testing.T) {
	url := fmt.Sprintf("%s/api/stock/%s/price", baseURL, symbol)
	body := map[string]interface{}{"price": 120.0}
	data, code := doRequest(t, "POST", url, body)

	if code != http.StatusOK {
		t.Errorf("期望状态码200, 得到%d, 响应: %s", code, string(data))
	}
	t.Logf("更新Stock价格成功: %s", string(data))
}

// 3. 获取用户SCOS余额
func TestGetUserSCOS(t *testing.T) {
	url := fmt.Sprintf("%s/api/user/%s/scos", baseURL, userAddr)
	data, code := doRequest(t, "GET", url, nil)

	if code != http.StatusOK {
		t.Errorf("期望状态码200, 得到%d, 响应: %s", code, string(data))
	}
	t.Logf("获取用户SCOS余额成功: %s", string(data))
}

// 4. 质押Stock
func TestStakeStock(t *testing.T) {
	url := fmt.Sprintf("%s/api/stake", baseURL)
	body := map[string]interface{}{
		"user_address":  userAddr,
		"token_address": tokenAddr,
		"chain":         chainName,
		"amount":        "100",
		"stock_symbol":  symbol,
	}
	data, code := doRequest(t, "POST", url, body)

	if code != http.StatusOK {
		t.Errorf("期望状态码200, 得到%d, 响应: %s", code, string(data))
	}
	t.Logf("质押Stock成功: %s", string(data))
}

// 5. 赎回Stock
func TestRedeemStock(t *testing.T) {
	url := fmt.Sprintf("%s/api/redeem", baseURL)
	body := map[string]interface{}{
		"user_address":  userAddr,
		"token_address": tokenAddr,
		"chain":         chainName,
	}
	data, code := doRequest(t, "POST", url, body)

	if code != http.StatusOK {
		t.Errorf("期望状态码200, 得到%d, 响应: %s", code, string(data))
	}
	t.Logf("赎回Stock成功: %s", string(data))
}

// 6. 买入Stock
func TestBuyStock(t *testing.T) {
	url := fmt.Sprintf("%s/api/buy", baseURL)
	body := map[string]interface{}{
		"user_address":  userAddr,
		"token_address": tokenAddr,
		"chain":         chainName,
		"amount":        "100",
	}
	data, code := doRequest(t, "POST", url, body)

	if code != http.StatusOK {
		t.Errorf("期望状态码200, 得到%d, 响应: %s", code, string(data))
	}
	t.Logf("买入Stock成功: %s", string(data))
}

// 7. 卖出Stock
func TestSellStock(t *testing.T) {
	url := fmt.Sprintf("%s/api/sell", baseURL)
	body := map[string]interface{}{
		"user_address":  userAddr,
		"token_address": tokenAddr,
		"chain":         chainName,
		"amount":        "100",
	}
	data, code := doRequest(t, "POST", url, body)

	if code != http.StatusOK {
		t.Errorf("期望状态码200, 得到%d, 响应: %s", code, string(data))
	}
	t.Logf("卖出Stock成功: %s", string(data))
}
