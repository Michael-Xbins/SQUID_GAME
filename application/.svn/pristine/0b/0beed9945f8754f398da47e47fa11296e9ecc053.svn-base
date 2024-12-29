package recharge

import (
	"application/api/presenter"
	"application/mongodb"
	"application/pkg/utils"
	"application/pkg/utils/log"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
	"github.com/skip2/go-qrcode"
	"io"
)

const (
	QRCodeSize = 256 // 二维码大小
)

// AssignWalletToUser 分配新生成的钱包地址给用户，并进行绑定
func AssignWalletToUser(ctx context.Context, network string, userInfo *presenter.UserInfo) (string, error) {
	if network == utils.NetworkTRC20 {
		// 生成新的钱包地址和私钥
		address, err := GenerateTronWallet(ctx, userInfo)
		if err != nil {
			return "", fmt.Errorf("failed to generate wallet: %v", err)
		}
		return address, nil
	} else if network == utils.NetworkTON {
		return "", nil
	}
	return "", fmt.Errorf("unsupported network: %s", network)
}

// GenerateTronWallet 生成新的 TRON 钱包地址和私钥，并将其存储到数据库
func GenerateTronWallet(ctx context.Context, userInfo *presenter.UserInfo) (string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)

	//fmt.Println("privateKey:", hexutil.Encode(privateKeyBytes)[2:])

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Error("error casting public key to ECDSA")
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	log.Infof("amount:%s, publicKey:%s", userInfo.Account, hexutil.Encode(publicKeyBytes)[2:])
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	address = "41" + address[2:]
	addb, _ := hex.DecodeString(address)
	hash1 := s256(s256(addb))
	secret := hash1[:4]
	for _, v := range secret {
		addb = append(addb, v)
	}
	//fmt.Println("address base58: ", base58.Encode(addb))
	readlAddress := base58.Encode(addb)
	//readlAddress = "TCr7k73wZR3cPkvHzC1JBZM4e1R9H5rU1j" // 临时
	// 存储钱包信息到数据库
	if err := mongodb.StoreUserWallet(ctx, readlAddress, hexutil.Encode(privateKeyBytes)[2:], userInfo); err != nil {
		return "", fmt.Errorf("failed to store wallet info in database: %v", err)
	}
	return readlAddress, nil
}
func s256(s []byte) []byte {
	h := sha256.New()
	h.Write(s)
	bs := h.Sum(nil)
	return bs
}

func GenerateQRCode(address, network string) (string, error) {
	var content string
	switch network {
	case utils.NetworkTRC20:
		content = address
	case utils.NetworkTON:
		content = address
	default:
		return "", fmt.Errorf("unsupported network: %s, only TRC20 and TON are supported", network)
	}
	qr, err := qrcode.Encode(content, qrcode.Medium, QRCodeSize)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %v", err)
	}
	return base64.StdEncoding.EncodeToString(qr), nil
}

func EncryptPrivateKey(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)
	return ciphertext, nil
}
func decryptPrivateKey(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext, nil
}
