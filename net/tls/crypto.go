package tls

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// ===== AES 加密/解密 =====

// AESEncrypt 使用 AES-CBC 模式加密数据。
//
// 参数:
//   - plaintext: 明文数据
//   - key: 密钥（16/24/32 字节对应 AES-128/AES-192/AES-256）
//   - iv: 初始向量（16 字节）
//
// 返回值:
//   - []byte: 密文数据
//   - error: 加密错误
func AESEncrypt(plaintext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("https: aes new cipher failed: %w", err)
	}

	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("https: iv length must be %d bytes", aes.BlockSize)
	}

	padded := pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	return ciphertext, nil
}

// AESDecrypt 使用 AES-CBC 模式解密数据。
//
// 参数:
//   - ciphertext: 密文数据
//   - key: 密钥（需与加密时一致）
//   - iv: 初始向量（需与加密时一致）
//
// 返回值:
//   - []byte: 明文数据
//   - error: 解密错误
func AESDecrypt(ciphertext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("https: aes new cipher failed: %w", err)
	}

	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("https: iv length must be %d bytes", aes.BlockSize)
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("https: ciphertext too short")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	return pkcs7Unpad(plaintext, aes.BlockSize)
}

// AESGCMEncrypt 使用 AES-GCM 模式加密数据（推荐，自带认证）。
//
// 参数:
//   - plaintext: 明文数据
//   - key: 密钥（16/24/32 字节）
//   - additionalData: 附加认证数据（可选）
//
// 返回值:
//   - []byte: nonce + 密文 + tag（拼接格式）
//   - error: 加密错误
func AESGCMEncrypt(plaintext, key, additionalData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("https: aes new cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("https: new gcm failed: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("https: generate nonce failed: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, additionalData)
	return ciphertext, nil
}

// AESGCMDecrypt 使用 AES-GCM 模式解密数据。
//
// 参数:
//   - ciphertext: nonce + 密文 + tag（AESGCMEncrypt 的输出格式）
//   - key: 密钥（需与加密时一致）
//   - additionalData: 附加认证数据（需与加密时一致）
//
// 返回值:
//   - []byte: 明文数据
//   - error: 解密错误
func AESGCMDecrypt(ciphertext, key, additionalData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("https: aes new cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("https: new gcm failed: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("https: ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		return nil, fmt.Errorf("https: gcm open failed: %w", err)
	}

	return plaintext, nil
}

// pkcs7Pad PKCS7 填充。
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

// pkcs7Unpad PKCS7 去填充。
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("https: empty data")
	}
	if len(data)%blockSize != 0 {
		return nil, fmt.Errorf("https: invalid padding size")
	}

	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize {
		return nil, fmt.Errorf("https: invalid padding")
	}

	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("https: invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}

// ===== RSA 加密/解密/签名 =====

// RSAGenerateKey 生成 RSA 密钥对。
//
// 参数:
//   - bits: 密钥位数（建议 2048 或 4096）
//
// 返回值:
//   - *rsa.PrivateKey: 私钥
//   - error: 生成错误
func RSAGenerateKey(bits int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("https: generate rsa key failed: %w", err)
	}
	return privateKey, nil
}

// RSAEncrypt 使用 RSA 公钥加密数据（OAEP-SHA256）。
//
// 参数:
//   - publicKey: RSA 公钥
//   - plaintext: 明文数据
//
// 返回值:
//   - []byte: 密文数据
//   - error: 加密错误
func RSAEncrypt(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, plaintext, nil)
	if err != nil {
		return nil, fmt.Errorf("https: rsa encrypt failed: %w", err)
	}
	return ciphertext, nil
}

// RSADecrypt 使用 RSA 私钥解密数据（OAEP-SHA256）。
//
// 参数:
//   - privateKey: RSA 私钥
//   - ciphertext: 密文数据
//
// 返回值:
//   - []byte: 明文数据
//   - error: 解密错误
func RSADecrypt(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("https: rsa decrypt failed: %w", err)
	}
	return plaintext, nil
}

// RSASign 使用 RSA 私钥签名数据（PKCS1v15-SHA256）。
//
// 参数:
//   - privateKey: RSA 私钥
//   - data: 待签名数据
//
// 返回值:
//   - []byte: 签名
//   - error: 签名错误
func RSASign(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return nil, fmt.Errorf("https: rsa sign failed: %w", err)
	}
	return signature, nil
}

// RSAVerify 使用 RSA 公钥验证签名。
//
// 参数:
//   - publicKey: RSA 公钥
//   - data: 原始数据
//   - signature: 签名
//
// 返回值:
//   - error: 验证错误（nil 表示验证通过）
func RSAVerify(publicKey *rsa.PublicKey, data, signature []byte) error {
	hash := sha256.Sum256(data)
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature); err != nil {
		return fmt.Errorf("https: rsa verify failed: %w", err)
	}
	return nil
}

// ===== PEM 编解码 =====

// MarshalRSAPrivateKey 将 RSA 私钥编码为 PEM 格式。
func MarshalRSAPrivateKey(privateKey *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
}

// MarshalRSAPublicKey 将 RSA 公钥编码为 PEM 格式。
func MarshalRSAPublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("https: marshal public key failed: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}), nil
}

// ParseRSAPrivateKey 从 PEM 数据解析 RSA 私钥。
func ParseRSAPrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("https: failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("https: parse private key failed: %w", err2)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("https: key is not RSA private key")
		}
	}
	return privateKey, nil
}

// ParseRSAPublicKey 从 PEM 数据解析 RSA 公钥。
func ParseRSAPublicKey(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("https: failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("https: parse public key failed: %w", err)
	}

	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("https: key is not RSA public key")
	}
	return publicKey, nil
}
