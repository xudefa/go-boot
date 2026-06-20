package tls

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestAESEncryptDecrypt(t *testing.T) {
	key := []byte("0123456789abcdef") // 16 bytes for AES-128
	iv := []byte("1234567890abcdef")  // 16 bytes
	plaintext := []byte("Hello, go-boot HTTPS!")

	ciphertext, err := AESEncrypt(plaintext, key, iv)
	if err != nil {
		t.Fatalf("AESEncrypt() error = %v", err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext should not equal plaintext")
	}

	decrypted, err := AESDecrypt(ciphertext, key, iv)
	if err != nil {
		t.Fatalf("AESDecrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %s, want %s", decrypted, plaintext)
	}
}

func TestAESEncryptDecrypt_InvalidKey(t *testing.T) {
	key := []byte("short") // invalid length
	iv := []byte("1234567890abcdef")

	_, err := AESEncrypt([]byte("data"), key, iv)
	if err == nil {
		t.Error("expected error for invalid key size")
	}
}

func TestAESGCMEncryptDecrypt(t *testing.T) {
	key := []byte("0123456789abcdef") // 16 bytes
	plaintext := []byte("Hello, go-boot GCM!")
	additionalData := []byte("auth-data")

	ciphertext, err := AESGCMEncrypt(plaintext, key, additionalData)
	if err != nil {
		t.Fatalf("AESGCMEncrypt() error = %v", err)
	}

	decrypted, err := AESGCMDecrypt(ciphertext, key, additionalData)
	if err != nil {
		t.Fatalf("AESGCMDecrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %s, want %s", decrypted, plaintext)
	}
}

func TestAESGCMEncryptDecrypt_WrongAdditionalData(t *testing.T) {
	key := []byte("0123456789abcdef")
	plaintext := []byte("Hello, go-boot GCM!")

	ciphertext, err := AESGCMEncrypt(plaintext, key, []byte("auth-data"))
	if err != nil {
		t.Fatalf("AESGCMEncrypt() error = %v", err)
	}

	_, err = AESGCMDecrypt(ciphertext, key, []byte("wrong-auth"))
	if err == nil {
		t.Error("expected error for wrong additional data")
	}
}

func TestRSAGenerateKey(t *testing.T) {
	key, err := RSAGenerateKey(2048)
	if err != nil {
		t.Fatalf("RSAGenerateKey() error = %v", err)
	}
	if key.N.BitLen() != 2048 {
		t.Errorf("key size = %d, want 2048", key.N.BitLen())
	}
}

func TestRSAEncryptDecrypt(t *testing.T) {
	privateKey, err := RSAGenerateKey(2048)
	if err != nil {
		t.Fatalf("RSAGenerateKey() error = %v", err)
	}

	plaintext := []byte("Hello, go-boot RSA!")

	ciphertext, err := RSAEncrypt(&privateKey.PublicKey, plaintext)
	if err != nil {
		t.Fatalf("RSAEncrypt() error = %v", err)
	}

	decrypted, err := RSADecrypt(privateKey, ciphertext)
	if err != nil {
		t.Fatalf("RSADecrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %s, want %s", decrypted, plaintext)
	}
}

func TestRSASignVerify(t *testing.T) {
	privateKey, err := RSAGenerateKey(2048)
	if err != nil {
		t.Fatalf("RSAGenerateKey() error = %v", err)
	}

	data := []byte("data to sign")

	signature, err := RSASign(privateKey, data)
	if err != nil {
		t.Fatalf("RSASign() error = %v", err)
	}

	if err := RSAVerify(&privateKey.PublicKey, data, signature); err != nil {
		t.Errorf("RSAVerify() error = %v", err)
	}

	// Wrong data should fail
	if err := RSAVerify(&privateKey.PublicKey, []byte("wrong data"), signature); err == nil {
		t.Error("expected verify error for wrong data")
	}
}

func TestMarshalParseRSAPrivateKey(t *testing.T) {
	privateKey, err := RSAGenerateKey(2048)
	if err != nil {
		t.Fatalf("RSAGenerateKey() error = %v", err)
	}

	pemData := MarshalRSAPrivateKey(privateKey)

	parsed, err := ParseRSAPrivateKey(pemData)
	if err != nil {
		t.Fatalf("ParseRSAPrivateKey() error = %v", err)
	}

	if !privateKey.Equal(parsed) {
		t.Error("parsed private key does not match original")
	}
}

func TestMarshalParseRSAPublicKey(t *testing.T) {
	privateKey, err := RSAGenerateKey(2048)
	if err != nil {
		t.Fatalf("RSAGenerateKey() error = %v", err)
	}

	pemData, err := MarshalRSAPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("MarshalRSAPublicKey() error = %v", err)
	}

	parsed, err := ParseRSAPublicKey(pemData)
	if err != nil {
		t.Fatalf("ParseRSAPublicKey() error = %v", err)
	}

	if !privateKey.PublicKey.Equal(parsed) {
		t.Error("parsed public key does not match original")
	}
}

func TestPKCS7PadUnpad(t *testing.T) {
	data := []byte("hello")
	blockSize := 16

	padded := pkcs7Pad(data, blockSize)
	if len(padded) != 16 {
		t.Errorf("padded length = %d, want 16", len(padded))
	}

	unpadded, err := pkcs7Unpad(padded, blockSize)
	if err != nil {
		t.Fatalf("pkcs7Unpad() error = %v", err)
	}

	if !bytes.Equal(unpadded, data) {
		t.Errorf("unpadded = %s, want %s", unpadded, data)
	}
}

func TestPKCS7PadFullBlock(t *testing.T) {
	data := make([]byte, 16)
	_, _ = rand.Read(data)

	padded := pkcs7Pad(data, 16)
	if len(padded) != 32 {
		t.Errorf("full block padded length = %d, want 32", len(padded))
	}
}

func TestPKCS7UnpadInvalid(t *testing.T) {
	_, err := pkcs7Unpad([]byte{}, 16)
	if err == nil {
		t.Error("expected error for empty data")
	}

	_, err = pkcs7Unpad([]byte{1, 2, 3}, 16)
	if err == nil {
		t.Error("expected error for invalid length")
	}
}

func TestAESGCM_RoundTripRandom(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("rand.Read error = %v", err)
	}

	plaintext := make([]byte, 256)
	if _, err := rand.Read(plaintext); err != nil {
		t.Fatalf("rand.Read error = %v", err)
	}

	ciphertext, err := AESGCMEncrypt(plaintext, key, nil)
	if err != nil {
		t.Fatalf("AESGCMEncrypt() error = %v", err)
	}

	decrypted, err := AESGCMDecrypt(ciphertext, key, nil)
	if err != nil {
		t.Fatalf("AESGCMDecrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("decrypted data does not match original")
	}
}

func TestRSAEncryptWithPEMRoundTrip(t *testing.T) {
	privateKey, err := RSAGenerateKey(2048)
	if err != nil {
		t.Fatalf("RSAGenerateKey() error = %v", err)
	}

	privPEM := MarshalRSAPrivateKey(privateKey)
	pubPEM, err := MarshalRSAPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("MarshalRSAPublicKey() error = %v", err)
	}

	parsedPriv, err := ParseRSAPrivateKey(privPEM)
	if err != nil {
		t.Fatalf("ParseRSAPrivateKey() error = %v", err)
	}

	parsedPub, err := ParseRSAPublicKey(pubPEM)
	if err != nil {
		t.Fatalf("ParseRSAPublicKey() error = %v", err)
	}

	plaintext := []byte("PEM round trip test")
	ciphertext, err := RSAEncrypt(parsedPub, plaintext)
	if err != nil {
		t.Fatalf("RSAEncrypt() error = %v", err)
	}

	decrypted, err := RSADecrypt(parsedPriv, ciphertext)
	if err != nil {
		t.Fatalf("RSADecrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("PEM round trip decrypted data does not match")
	}
}
