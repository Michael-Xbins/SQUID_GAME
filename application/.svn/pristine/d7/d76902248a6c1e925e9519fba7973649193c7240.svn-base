package wallet

import (
	"application/pkg/utils"
	"application/pkg/utils/log"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	withdrawApikey         = "LY6sS2fABFXqiXFIgXBCBFKVzb5krQaAHXDrM4VWRupLvtooEMn2aTP8FIsuy0La"
	withdrawPrivatekeyPath = "./wallet/withdraw_privateKey.txt"
)

func Withdrawal(address string, network string, amount string) (string, error) {
	pemData, err := ioutil.ReadFile(withdrawPrivatekeyPath)
	if err != nil {
		log.Errorf("Failed to read private key file: %v", err)
		return "", err
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		log.Errorf("Failed to parse PEM block containing the key")
		return "", errors.New("failed to parse PEM block containing the key")
	}

	privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Errorf("Failed to parse PKCS#8 private key: %v", err)
		return "", err
	}

	edKey, ok := privateKeyInterface.(ed25519.PrivateKey)
	if !ok {
		log.Errorf("Not an Ed25519 private key")
		return "", errors.New("not an Ed25519 private key")
	}

	if network == utils.NetworkTRC20 {
		network = "TRX"
	}
	params := url.Values{}
	params.Add("coin", "USDT")
	params.Add("network", network)
	params.Add("address", address)
	params.Add("amount", amount)
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
		log.Errorf("Failed to create request: %v", err)
		return "", err
	}
	req.Header.Add("X-MBX-APIKEY", withdrawApikey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(req)
	if err != nil {
		log.Errorf("Failed to send request: %v", err)
		return "", err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf("Failed to read response body: %v", err)
		return "", err
	}
	var resp struct {
		ID      string `json:"id"`
		Msg     string `json:"msg,omitempty"`
		Success bool   `json:"success,omitempty"`
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Errorf("Failed to parse response JSON: %v", err)
		return "", err
	}

	if resp.ID != "" {
		log.Infof("提现成功, Order ID: %s", resp.ID)
		return resp.ID, nil
	} else {
		log.Errorf("提现失败: %s", resp.Msg)
		return "", errors.New("withdrawal transaction failed")
	}
}

func WithdrawHistoryById(binId string) string {
	// Load the private key from PEM file
	pemData, err := ioutil.ReadFile(withdrawPrivatekeyPath)
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
	req.Header.Add("X-MBX-APIKEY", withdrawApikey)

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
