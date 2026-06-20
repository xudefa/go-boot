# https 包 — HTTPS 客户端与加解密工具

## 概述

`https` 包提供 HTTPS 客户端构建、TLS 证书管理以及通用加解密工具。核心功能：

- **HTTPS 客户端** — 基于 `net.NetClient` 封装，支持 TLS 配置、超时、默认请求头
- **AES 加解密** — CBC 模式（带 PKCS7 填充）和 GCM 模式（推荐，自带认证）
- **RSA 加解密** — OAEP-SHA256 加密/解密
- **RSA 签名** — PKCS1v15-SHA256 签名/验签
- **TLS 配置** — 从文件或 PEM 数据加载证书，支持 CA 验证
- **PEM 编解码** — 密钥的 PEM 格式序列化与反序列化

---

## HTTPS 客户端

### NewClient — 创建 HTTPS 客户端

```go
func NewClient(baseURL string, opts ...ClientOption) *net.NetClient
```

```go
client := https.NewClient("https://api.example.com",
    https.WithTimeout(30*time.Second),
    https.WithDefaultHeader("Authorization", "Bearer token123"),
)
resp, err := client.Get(ctx, "/api/users")
```

### ClientOption

| 选项 | 说明 |
|------|------|
| `WithTLSConfig(tlsConfig)` | 设置 TLS 配置 |
| `WithInsecureTLS()` | 跳过证书验证（仅开发测试） |
| `WithTimeout(timeout)` | 设置请求超时时间 |
| `WithDefaultHeader(key, value)` | 设置默认请求头 |
| `WithTransport(transport)` | 设置自定义 HTTP Transport |

#### WithTLSConfig

```go
tlsCfg, _ := https.LoadTLSConfig("server.crt", "server.key")
client := https.NewClient("https://example.com",
    https.WithTLSConfig(tlsCfg),
)
```

#### WithInsecureTLS

```go
// 仅用于开发测试，跳过证书验证
client := https.NewClient("https://localhost:8443",
    https.WithInsecureTLS(),
)
```

#### WithTimeout

```go
client := https.NewClient("https://api.example.com",
    https.WithTimeout(10*time.Second),
)
```

#### WithDefaultHeader

```go
client := https.NewClient("https://api.example.com",
    https.WithDefaultHeader("Authorization", "Bearer eyJhbGciOiJ..."),
    https.WithDefaultHeader("X-Request-Id", "uuid-123"),
)
```

#### WithTransport

```go
transport := &http.Transport{
    MaxIdleConns:    50,
    IdleConnTimeout: 60 * time.Second,
}
client := https.NewClient("https://api.example.com",
    https.WithTransport(transport),
)
```

### 完整使用示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "github.com/xudefa/go-boot/https"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    client := https.NewClient("https://jsonplaceholder.typicode.com",
        https.WithTimeout(10*time.Second),
    )

    ctx := context.Background()

    // GET 请求
    resp, err := client.Get(ctx, "/users/1")
    if err != nil {
        log.Fatal(err)
    }

    if resp.IsSuccess() {
        var user User
        resp.Bind(&user)
        fmt.Printf("用户: %+v\n", user)
    }

    // POST 请求
    newUser := map[string]any{"name": "张三"}
    resp, err = client.Post(ctx, "/users", newUser)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("创建成功:", resp.String())
}
```

---

## TLS 配置

### LoadTLSConfig — 从文件加载 TLS 配置

```go
func LoadTLSConfig(certFile, keyFile string) (*tls.Config, error)
```

```go
tlsCfg, err := https.LoadTLSConfig("server.crt", "server.key")
if err != nil {
    log.Fatal(err)
}
```

### LoadTLSConfigWithCA — 加载 TLS 配置并启用 CA 验证

```go
func LoadTLSConfigWithCA(certFile, keyFile, caFile string) (*tls.Config, error)
```

```go
tlsCfg, err := https.LoadTLSConfigWithCA("client.crt", "client.key", "ca.crt")
if err != nil {
    log.Fatal(err)
}
```

### LoadClientTLSConfig — 仅客户端验证（无需客户端证书）

```go
func LoadClientTLSConfig(caFile string) (*tls.Config, error)
```

```go
tlsCfg, err := https.LoadClientTLSConfig("ca.crt")
if err != nil {
    log.Fatal(err)
}
// 验证服务端证书，但不需要客户端证书
```

### LoadCertFromPEM — 从 PEM 字节数据加载

```go
func LoadCertFromPEM(certPEM, keyPEM []byte) (*tls.Config, error)
```

```go
certPEM := []byte("-----BEGIN CERTIFICATE-----\n...")
keyPEM := []byte("-----BEGIN RSA PRIVATE KEY-----\n...")
tlsCfg, err := https.LoadCertFromPEM(certPEM, keyPEM)
```

### InsecureTLSConfig — 跳过证书验证

```go
func InsecureTLSConfig() *tls.Config
```

```go
cfg := https.InsecureTLSConfig()
// cfg.InsecureSkipVerify = true, MinVersion = tls.VersionTLS12
```

### MustLoadTLSConfig — 失败时 panic

```go
func MustLoadTLSConfig(certFile, keyFile string) *tls.Config
```

```go
cfg := https.MustLoadTLSConfig("server.crt", "server.key")
// 加载失败直接 panic，适用于初始化阶段
```

---

## AES 加解密

### AESEncrypt / AESDecrypt — CBC 模式（带 PKCS7 填充）

```go
func AESEncrypt(plaintext, key, iv []byte) ([]byte, error)
func AESDecrypt(ciphertext, key, iv []byte) ([]byte, error)
```

- `key`：16/24/32 字节，对应 AES-128/AES-192/AES-256
- `iv`：16 字节初始向量

```go
key := []byte("1234567890123456")  // AES-128
iv := []byte("1234567890123456")   // 16 字节

plaintext := []byte("Hello, World!")

ciphertext, err := https.AESEncrypt(plaintext, key, iv)
if err != nil {
    log.Fatal(err)
}

decrypted, err := https.AESDecrypt(ciphertext, key, iv)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(decrypted)) // "Hello, World!"
```

### AESGCMEncrypt / AESGCMDecrypt — GCM 模式（推荐）

```go
func AESGCMEncrypt(plaintext, key, additionalData []byte) ([]byte, error)
func AESGCMDecrypt(ciphertext, key, additionalData []byte) ([]byte, error)
```

- GCM 模式自带认证标签，能检测密文篡改
- 输出格式：`nonce + ciphertext + tag`（nonce 为随机生成 12 字节）
- `additionalData` 为可选的附加认证数据

```go
key := []byte("1234567890123456") // AES-128

plaintext := []byte("Hello, World!")

// 加密
ciphertext, err := https.AESGCMEncrypt(plaintext, key, nil)
if err != nil {
    log.Fatal(err)
}

// 解密
decrypted, err := https.AESGCMDecrypt(ciphertext, key, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(decrypted)) // "Hello, World!"

// 带附加认证数据
aad := []byte("additional-data")
ciphertext, _ = https.AESGCMEncrypt(plaintext, key, aad)
decrypted, _ = https.AESGCMDecrypt(ciphertext, key, aad)
```

---

## RSA 加解密

### RSAGenerateKey — 生成密钥对

```go
func RSAGenerateKey(bits int) (*rsa.PrivateKey, error)
```

```go
privateKey, err := https.RSAGenerateKey(2048)
if err != nil {
    log.Fatal(err)
}
publicKey := &privateKey.PublicKey
```

### RSAEncrypt / RSADecrypt — OAEP-SHA256

```go
func RSAEncrypt(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error)
func RSADecrypt(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error)
```

```go
privateKey, _ := https.RSAGenerateKey(2048)
publicKey := &privateKey.PublicKey

plaintext := []byte("Hello, RSA!")
ciphertext, err := https.RSAEncrypt(publicKey, plaintext)
if err != nil {
    log.Fatal(err)
}

decrypted, err := https.RSADecrypt(privateKey, ciphertext)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(decrypted)) // "Hello, RSA!"
```

---

## RSA 签名与验签

### RSASign / RSAVerify — PKCS1v15-SHA256

```go
func RSASign(privateKey *rsa.PrivateKey, data []byte) ([]byte, error)
func RSAVerify(publicKey *rsa.PublicKey, data, signature []byte) error
```

```go
privateKey, _ := https.RSAGenerateKey(2048)
publicKey := &privateKey.PublicKey

data := []byte("待签名的数据")

// 签名
signature, err := https.RSASign(privateKey, data)
if err != nil {
    log.Fatal(err)
}

// 验签
err = https.RSAVerify(publicKey, data, signature)
if err != nil {
    log.Fatal("签名验证失败")
} else {
    fmt.Println("签名验证通过")
}
```

---

## PEM 编解码

### 编码

```go
func MarshalRSAPrivateKey(privateKey *rsa.PrivateKey) []byte
func MarshalRSAPublicKey(publicKey *rsa.PublicKey) ([]byte, error)
```

```go
privateKey, _ := https.RSAGenerateKey(2048)

// 私钥 -> PEM
privPEM := https.MarshalRSAPrivateKey(privateKey)
os.WriteFile("private.pem", privPEM, 0600)

// 公钥 -> PEM
pubPEM, _ := https.MarshalRSAPublicKey(&privateKey.PublicKey)
os.WriteFile("public.pem", pubPEM, 0644)
```

### 解码

```go
func ParseRSAPrivateKey(pemData []byte) (*rsa.PrivateKey, error)
func ParseRSAPublicKey(pemData []byte) (*rsa.PublicKey, error)
```

```go
// 从 PEM 文件读取
pemData, _ := os.ReadFile("private.pem")
privateKey, err := https.ParseRSAPrivateKey(pemData)
if err != nil {
    log.Fatal(err)
}

pemData, _ = os.ReadFile("public.pem")
publicKey, err := https.ParseRSAPublicKey(pemData)
if err != nil {
    log.Fatal(err)
}
```

---

## 完整 API 参考

| 函数/类型 | 说明 |
|-----------|------|
| `NewClient(baseURL, opts...)` | 创建 HTTPS 客户端 |
| `WithTLSConfig(cfg)` | 设置 TLS 配置 |
| `WithInsecureTLS()` | 跳过证书验证 |
| `WithTimeout(timeout)` | 设置超时 |
| `WithDefaultHeader(k, v)` | 设置默认请求头 |
| `WithTransport(t)` | 设置自定义 Transport |
| `LoadTLSConfig(certFile, keyFile)` | 从文件加载 TLS 配置 |
| `LoadTLSConfigWithCA(cert, key, ca)` | 加载 TLS + CA 验证 |
| `LoadClientTLSConfig(caFile)` | 加载客户端 TLS 配置 |
| `LoadCertFromPEM(certPEM, keyPEM)` | 从 PEM 字节加载 |
| `InsecureTLSConfig()` | 创建不安全 TLS 配置 |
| `MustLoadTLSConfig(cert, key)` | 加载 TLS 配置（失败 panic） |
| `AESEncrypt(plain, key, iv)` | AES-CBC 加密 |
| `AESDecrypt(cipher, key, iv)` | AES-CBC 解密 |
| `AESGCMEncrypt(plain, key, aad)` | AES-GCM 加密（推荐） |
| `AESGCMDecrypt(cipher, key, aad)` | AES-GCM 解密 |
| `RSAGenerateKey(bits)` | 生成 RSA 密钥对 |
| `RSAEncrypt(pub, plain)` | RSA OAEP 加密 |
| `RSADecrypt(priv, cipher)` | RSA OAEP 解密 |
| `RSASign(priv, data)` | RSA PKCS1v15 签名 |
| `RSAVerify(pub, data, sig)` | RSA 验签 |
| `MarshalRSAPrivateKey(key)` | RSA 私钥→PEM |
| `MarshalRSAPublicKey(key)` | RSA 公钥→PEM |
| `ParseRSAPrivateKey(pem)` | PEM→RSA 私钥 |
| `ParseRSAPublicKey(pem)` | PEM→RSA 公钥 |
