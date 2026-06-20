// Package tls 提供 TLS 客户端构建、TLS 证书管理以及通用加解密工具。
//
// 核心功能：
//   - TLS 证书加载和 TLS 配置构建
//   - TLS 客户端构建器
//   - 对称加密（AES-CBC/AES-GCM）和非对称加密（RSA）工具
//
// 快速开始:
//
//	// 加载 TLS 配置
//	tlsCfg, err := tls.LoadTLSConfig("cert.pem", "key.pem")
//
//	// 创建 TLS 客户端
//	client := tls.NewClient("https://example.com",
//	    tls.WithTLSConfig(tlsCfg),
//	)
//	resp, err := client.Get(ctx, "/api/users")
//
//	// AES 加密
//	ciphertext, err := tls.AESEncrypt([]byte("plaintext"), key, iv)
//
//	// RSA 加密
//	ciphertext, err := tls.RSAEncrypt(publicKey, []byte("plaintext"))
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// LoadTLSConfig 从证书文件和密钥文件加载 TLS 配置。
//
// 参数:
//   - certFile: PEM 格式的证书文件路径
//   - keyFile: PEM 格式的私钥文件路径
//
// 返回值:
//   - *tls.Config: TLS 配置
//   - error: 加载错误
//
// 示例:
//
//	tlsCfg, err := tls.LoadTLSConfig("server.crt", "server.key")
//	if err != nil {
//	    log.Fatal(err)
//	}
func LoadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("tls: load x509 key pair failed: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// LoadTLSConfigWithCA 加载 TLS 配置并启用 CA 证书验证（用于客户端验证服务端）。
//
// 参数:
//   - certFile: PEM 格式的证书文件路径
//   - keyFile: PEM 格式的私钥文件路径
//   - caFile: PEM 格式的 CA 证书文件路径
//
// 返回值:
//   - *tls.Config: TLS 配置
//   - error: 加载错误
func LoadTLSConfigWithCA(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("https: load x509 key pair failed: %w", err)
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("https: read CA file failed: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("https: failed to parse CA certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// LoadClientTLSConfig 加载仅客户端验证用的 TLS 配置（不需要服务端证书）。
//
// 参数:
//   - caFile: PEM 格式的 CA 证书文件路径，用于验证服务端证书
//
// 返回值:
//   - *tls.Config: TLS 配置
//   - error: 加载错误
func LoadClientTLSConfig(caFile string) (*tls.Config, error) {
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("https: read CA file failed: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("https: failed to parse CA certificate")
	}

	return &tls.Config{
		RootCAs:    caPool,
		MinVersion: tls.VersionTLS12,
	}, nil
}

// LoadCertFromPEM 从 PEM 编码的字节数据解析 TLS 证书对。
//
// 参数:
//   - certPEM: PEM 格式的证书数据
//   - keyPEM: PEM 格式的私钥数据
//
// 返回值:
//   - *tls.Config: TLS 配置
//   - error: 解析错误
func LoadCertFromPEM(certPEM, keyPEM []byte) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("https: parse x509 key pair failed: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// InsecureTLSConfig 创建跳过证书验证的 TLS 配置（仅用于开发测试）。
//
// 返回值:
//   - *tls.Config: 不安全但可用的 TLS 配置
func InsecureTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}
}

// MustLoadTLSConfig 加载 TLS 配置，失败时 panic。
//
// 参数:
//   - certFile: PEM 格式的证书文件路径
//   - keyFile: PEM 格式的私钥文件路径
//
// 返回值:
//   - *tls.Config: TLS 配置
func MustLoadTLSConfig(certFile, keyFile string) *tls.Config {
	cfg, err := LoadTLSConfig(certFile, keyFile)
	if err != nil {
		panic(fmt.Sprintf("https: must load TLS config: %v", err))
	}
	return cfg
}
