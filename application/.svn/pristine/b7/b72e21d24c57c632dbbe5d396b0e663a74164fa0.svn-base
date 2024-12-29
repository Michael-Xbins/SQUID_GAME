package recharge

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	APIKey = "a4039719-b24e-4a72-a69e-083e2612b9a2"
)

// TransactionResponse 用于解析区块链 API 返回的交易数据
type Trc20Response struct {
	Data    []Trc20Result `json:"data"`
	Success bool          `json:"success"`
	Meta    MetaData      `json:"meta"`
}
type Trc20Result struct {
	TransactionID  string      `json:"transaction_id"` // 交易hash
	BlockTimestamp int64       `json:"block_timestamp"`
	Value          string      `json:"value"`      // 交易金额
	From           string      `json:"from"`       // 发送方地址
	To             string      `json:"to"`         // 接收方地址
	Type           string      `json:"type"`       // 交易类型
	TokenInfo      TokenDetail `json:"token_info"` // 代币信息
}
type TokenDetail struct {
	Symbol   string `json:"symbol"`   // 代币符号
	Address  string `json:"address"`  // 代币合约地址
	Decimals int    `json:"decimals"` // 代币小数位
	Name     string `json:"name"`     // 代币名称
}
type MetaData struct {
	At       int64 `json:"at"`
	PageSize int   `json:"page_size"`
}

// CheckTRC20Status 检查 TRC20 交易状态
func CheckTRC20Status(ctx context.Context, address string, network string, startTimeMs int64) ([]Trc20Result, error) {
	if network != "TRC20" {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
	//apiUrl := fmt.Sprintf("https://api.trongrid.io/v1/accounts/%s/transactions/trc20?min_timestamp=%d&only_to=true", address, startTimeMs)// 查询min_timestamp之后的交易
	apiUrl := fmt.Sprintf("https://api.trongrid.io/v1/accounts/%s/transactions/trc20?only_to=true", address)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	// 添加 API 密钥
	req.Header.Add("TRON-PRO-API-KEY", APIKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var apiResponse Trc20Response
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}
	if !apiResponse.Success {
		return nil, fmt.Errorf("API response was not successful")
	}
	return apiResponse.Data, nil
}

//const (
//	// Transfer 事件的签名哈希
//	TransferEventSignature = "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
//
//	// USDT 合约地址
//	USDTContractAddressMainnet = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
//)
//
//type TronClient struct {
//	conn            *grpc.ClientConn
//	walletClient    api.WalletClient
//	walletExtClient api.WalletExtensionClient
//}
//
//// CheckAddressTransaction 检查地址的USDT交易
//func CheckAddressTransaction(rechargeInfo presenter.RechargeAddressInfo) (*TransactionResult, error) {
//	url := fmt.Sprintf("https://api.trongrid.io/v1/accounts/%s/transactions/trc20", rechargeInfo.Address)
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		return nil, fmt.Errorf("create request error: %w", err)
//	}
//	req.Header.Add("TRON-PRO-API-KEY", APIKey)
//	client := &http.Client{Timeout: 10 * time.Second}
//	resp, err := client.Do(req)
//	if err != nil {
//		return nil, fmt.Errorf("request error: %w", err)
//	}
//	defer resp.Body.Close()
//	var result struct {
//		Data []struct {
//			TransactionID string `json:"transaction_id"`
//			TokenInfo     struct {
//				Symbol  string `json:"symbol"`
//				Address string `json:"address"`
//			} `json:"token_info"`
//			From           string `json:"from"`
//			To             string `json:"to"`
//			Type           string `json:"type"`
//			Value          string `json:"value"`
//			BlockTimestamp int64  `json:"block_timestamp"`
//		} `json:"data"`
//	}
//
//	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
//		return nil, fmt.Errorf("decode response error: %w", err)
//	}
//
//	//// 获取这个地址对应的充值记录
//	//rechargeRecord, err := mongodb.GetRechargeRecord(context.Background(), address)
//	//if err != nil {
//	//	return nil, fmt.Errorf("get recharge record error: %w", err)
//	//}
//
//	// 只检查充值记录创建时间之后的交易
//	var validTx *TransactionResult
//	for _, tx := range result.Data {
//		// 检查交易时间是否在充值记录创建之后
//		txTime := time.Unix(tx.BlockTimestamp/1000, 0)
//		if txTime.Before(rechargeInfo.CreatedAt) {
//			continue
//		}
//
//		// 检查是否是USDT合约地址
//		if tx.TokenInfo.Address != USDTContractAddressMainnet {
//			continue
//		}
//
//		// 检查是否转入到指定地址
//		if !strings.EqualFold(tx.To, address) {
//			continue
//		}
//
//		// 解析转账金额
//		value := new(big.Int)
//		value.SetString(tx.Value, 10)
//
//		// 转换为USDT金额（6位精度）
//		amount := new(big.Float).SetInt(value)
//		divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil))
//		amount = new(big.Float).Quo(amount, divisor)
//
//		// 检查金额是否匹配
//		expectedAmountFloat := new(big.Float).SetInt64(int64(expectedAmount))
//		if amount.Cmp(expectedAmountFloat) == 0 {
//			// 找到匹配的交易
//			validTx = &TransactionResult{
//				TxID:      tx.TransactionID,
//				Amount:    amount,
//				From:      tx.From,
//				Timestamp: txTime,
//			}
//			break
//		}
//	}
//
//	if validTx == nil {
//		return nil, nil // 没有找到匹配的交易
//	}
//
//	// 检查这笔交易是否已经被处理过
//	processed, err := mongodb.IsTransactionProcessed(context.Background(), validTx.TxID)
//	if err != nil {
//		return nil, fmt.Errorf("check transaction processed error: %w", err)
//	}
//	if processed {
//		return nil, fmt.Errorf("transaction already processed")
//	}
//
//	return validTx, nil
//}
