package wallet

import (
	"application/pkg/utils/log"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type WithdrawHistory struct {
	ID              string `json:"id"`              // 提现记录唯一标识符, 由Binance系统生成
	Amount          string `json:"amount"`          // 提现转出金额
	TransactionFee  string `json:"transactionFee"`  // 手续费
	Coin            string `json:"coin"`            // 提币的资产种类, 例如USDT、BTC、ETH等
	Status          int    `json:"status"`          // 状态码(6)
	Address         string `json:"address"`         // 提现目的地址, 即资金被发送到的地址
	TxID            string `json:"txId"`            // 交易hash, 可在区块链上追踪这笔交易
	ApplyTime       string `json:"applyTime"`       // 提现请求的时间
	Network         string `json:"network"`         // 提现使用的网络, 例如ETH、BSC（Binance Smart Chain）等
	TransferType    int    `json:"transferType"`    // 1: 站内转账, 0: 站外转账
	WithdrawOrderId int    `json:"withdrawOrderId"` // 自定义ID, 如果没有则不返回该字段
	Info            string `json:"info"`            // 提现失败原因
	ConfirmNo       int    `json:"confirmNo"`       // 提现确认数
	WalletType      int    `json:"walletType"`      // 1: 资金钱包 0:现货钱包
	TxKey           string `json:"txKey"`           // 交易的关键信息
	CompleteTime    string `json:"completeTime"`    // 提现完成时间(UTC), 即资金到达目的地址的时间
}

const (
	transferApiKey        = "ETulBV59JimB3eIfv5nvvl91rPI5Vxz2vtzRH7z4LwASDzMkRqV2m8oHBMLu7cIG"
	privateKeyPath        = "./wallet/Private_key.txt"
	transferAddress       = "0x74748E5E84dBddDdd66E3dd0f6aC90585AbEaE75" //转账地址(获取梯子hash)
	withdrawAPIURL        = "https://api.binance.com/sapi/v1/capital/withdraw/apply"
	withdrawHistoryAPIURL = "https://api.binance.com/sapi/v1/capital/withdraw/history" //取款历史
	apiURL                = "https://api.binance.com/api/v3/account"
)

const (
	ErrFailedToRetrieve   = "Failed to retrieve withdrawal history"
	ErrNoTransactionFound = "No transaction found for the given ID"
)

func Transfer(roundNum int64) (string, error) {
	pemData, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key file: %v", err)
		return "", err
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		log.Fatalf("Failed to parse PEM block containing the key")
		return "", errors.New("failed to parse PEM block containing the key")
	}

	privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse PKCS#8 private key: %v", err)
		return "", err
	}

	edKey, ok := privateKeyInterface.(ed25519.PrivateKey)
	if !ok {
		log.Fatalf("Not an Ed25519 private key")
		return "", errors.New("not an Ed25519 private key")
	}

	params := url.Values{}
	params.Add("coin", "USDT")
	params.Add("network", "BSC") // Specify the BNB Chain (BEP20) network
	params.Add("address", transferAddress)
	params.Add("amount", "10") // 10 USDT
	params.Add("timestamp", strconv.FormatInt(time.Now().Unix()*1000, 10))

	// Sign the request
	payload := params.Encode()
	signature := ed25519.Sign(edKey, []byte(payload))
	encodedSignature := base64.StdEncoding.EncodeToString(signature)
	params.Add("signature", encodedSignature)

	// Send the request
	client := &http.Client{}
	req, err := http.NewRequest("POST", withdrawAPIURL, strings.NewReader(params.Encode()))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
		return "", err
	}
	req.Header.Add("X-MBX-APIKEY", transferApiKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
		return "", err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
		return "", err
	}
	var resp struct {
		ID      string `json:"id"`
		Msg     string `json:"msg,omitempty"`     // 可选字段
		Success bool   `json:"success,omitempty"` // 可选字段，不一定总是返回
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatalf("Failed to parse response JSON: %v", err)
		return "", err
	}

	if resp.ID != "" {
		orderID := resp.ID
		log.Infof("转账交易成功, Order ID: %s", orderID)

		startTime := time.Now() // 开始计时
		txHash, err := RetryHistoryById(orderID, 60, 1000*time.Millisecond)
		if err != nil {
			log.Errorf("Error: %v", err)
			return "", err
		}
		elapsedTime := time.Since(startTime)
		log.Infof("梯子期数:%d, Transaction Hash:%s, 发起交易 至 获取到交易哈希值, 经历了%f秒", roundNum, txHash, elapsedTime.Seconds())
		return txHash, nil

	} else {
		log.Errorf("转账交易失败: %s", resp.Msg)
		return "", errors.New("转账交易失败")
	}
}

func RetryHistoryById(orderID string, retries int, delay time.Duration) (string, error) {
	var txHash string
	for i := 0; i < retries; i++ {
		if i > 0 {
			time.Sleep(delay)
		}
		txHash = HistoryById(orderID)
		if txHash != "" && txHash != ErrNoTransactionFound && txHash != ErrFailedToRetrieve {
			return txHash, nil
		}
	}
	if txHash == "" {
		return "", fmt.Errorf("no transaction hash was found after %d retries", retries)
	}
	return "", fmt.Errorf("error after %d retries: %s", retries, txHash)
}

func HistoryById(binId string) string {
	// Load the private key from PEM file
	pemData, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		log.Errorf("Failed to read private key file: %v", err)
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		log.Errorf("Failed to parse PEM block containing the key")
	}

	privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Errorf("Failed to parse PKCS#8 private key: %v", err)
	}

	edKey, ok := privateKeyInterface.(ed25519.PrivateKey)
	if !ok {
		log.Errorf("Not an Ed25519 private key")
	}

	// Set up the request parameters
	params := url.Values{}
	params.Add("timestamp", strconv.FormatInt(time.Now().Unix()*1000, 10))
	params.Add("idList", binId)

	// Sign the request
	payload := params.Encode()
	signature := ed25519.Sign(edKey, []byte(payload))
	encodedSignature := base64.StdEncoding.EncodeToString(signature)
	params.Add("signature", encodedSignature)

	// Send the request
	client := &http.Client{}
	req, err := http.NewRequest("GET", withdrawHistoryAPIURL, nil)
	if err != nil {
		log.Errorf("Failed to create request: %v", err)
	}
	req.URL.RawQuery = params.Encode()
	req.Header.Add("X-MBX-APIKEY", transferApiKey)

	response, err := client.Do(req)
	if err != nil {
		log.Errorf("Failed to send request: %v", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("Failed to read response body: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		log.Errorf("Failed to retrieve withdrawal history: HTTP %d", response.StatusCode)
		return ErrFailedToRetrieve
	}

	var history []WithdrawHistory
	if err := json.Unmarshal(body, &history); err != nil {
		log.Errorf("Error parsing JSON: %v", err)
	}

	if len(history) == 0 {
		return ErrNoTransactionFound
	}

	for _, record := range history {
		if record.ID == binId {
			return record.TxID
		}
	}

	return ErrNoTransactionFound
}

// History 获取所有取款历史信息
func History() string {
	pemData, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		log.Errorf("Failed to read private key file: %v", err)
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		log.Errorf("Failed to parse PEM block containing the key")
	}
	privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Errorf("Failed to parse PKCS#8 private key: %v", err)
	}

	edKey, ok := privateKeyInterface.(ed25519.PrivateKey)
	if !ok {
		log.Errorf("Not an Ed25519 private key")
	}
	params := url.Values{}
	params.Add("timestamp", strconv.FormatInt(time.Now().Unix()*1000, 10))

	// 过滤条件
	params.Add("coin", "USDT")
	// params.Add("status", "0") // 0 for pending, 1 for success, etc.
	// params.Add("startTime", "1609459200000") // Start time in milliseconds
	// params.Add("endTime", "1640995200000") // End time in milliseconds

	payload := params.Encode()
	signature := ed25519.Sign(edKey, []byte(payload))
	encodedSignature := base64.StdEncoding.EncodeToString(signature)
	params.Add("signature", encodedSignature)
	client := &http.Client{}
	req, err := http.NewRequest("GET", withdrawHistoryAPIURL, nil)
	if err != nil {
		log.Errorf("Failed to create request: %v", err)
	}
	req.URL.RawQuery = params.Encode()
	req.Header.Add("X-MBX-APIKEY", transferApiKey)
	response, err := client.Do(req)
	if err != nil {
		log.Errorf("Failed to send request: %v", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("Failed to read response body: %v", err)
	}
	return string(body)
}
