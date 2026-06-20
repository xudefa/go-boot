package tls

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	goNet "net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func generateTestCert(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []goNet.IP{goNet.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("CreateCertificate error = %v", err)
	}

	certPEM = pemEncode("CERTIFICATE", certDER)
	keyPEM = pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(privateKey))
	return certPEM, keyPEM
}

func pemEncode(typ string, derBytes []byte) []byte {
	block := &pem.Block{
		Type:  typ,
		Bytes: derBytes,
	}
	return pem.EncodeToMemory(block)
}

func TestLoadCertFromPEM(t *testing.T) {
	certPEM, keyPEM := generateTestCert(t)

	tlsCfg, err := LoadCertFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("LoadCertFromPEM() error = %v", err)
	}
	if tlsCfg == nil {
		t.Fatal("LoadCertFromPEM() returned nil")
	}
	if len(tlsCfg.Certificates) != 1 {
		t.Errorf("expected 1 certificate, got %d", len(tlsCfg.Certificates))
	}
}

func TestInsecureTLSConfig(t *testing.T) {
	cfg := InsecureTLSConfig()
	if cfg == nil {
		t.Fatal("InsecureTLSConfig() returned nil")
	}
	if !cfg.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %d, want %d", cfg.MinVersion, tls.VersionTLS12)
	}
}

func TestMustLoadTLSConfig_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid cert file")
		}
	}()
	_ = MustLoadTLSConfig("nonexistent.pem", "nonexistent.key")
}

func TestNewClient(t *testing.T) {
	certPEM, keyPEM := generateTestCert(t)

	tlsCfg, err := LoadCertFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("LoadCertFromPEM() error = %v", err)
	}

	client := NewClient("https://localhost:8443",
		WithTLSConfig(tlsCfg),
		WithTimeout(5*time.Second),
		WithDefaultHeader("User-Agent", "go-boot-test"),
	)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestNewClient_Insecure(t *testing.T) {
	client := NewClient("https://example.com",
		WithInsecureTLS(),
	)
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestHttpsClientIntegration(t *testing.T) {
	certPEM, keyPEM := generateTestCert(t)

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("X509KeyPair error = %v", err)
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	ts.StartTLS()
	defer ts.Close()

	client := NewClient(ts.URL,
		WithInsecureTLS(),
	)

	resp, err := client.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestHttpsClientIntegration_AllMethods(t *testing.T) {
	certPEM, keyPEM := generateTestCert(t)

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("X509KeyPair error = %v", err)
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	ts.StartTLS()
	defer ts.Close()

	client := NewClient(ts.URL, WithInsecureTLS())
	ctx := context.Background()

	t.Run("GET", func(t *testing.T) {
		resp, err := client.Get(ctx, "/")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("POST", func(t *testing.T) {
		resp, err := client.Post(ctx, "/", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("Post() error = %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
		}
	})

	t.Run("PUT", func(t *testing.T) {
		resp, err := client.Put(ctx, "/", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("Put() error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("DELETE", func(t *testing.T) {
		resp, err := client.Delete(ctx, "/")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
		}
	})
}

func TestLoadTLSConfig(t *testing.T) {
	_, err := LoadTLSConfig("nonexistent.pem", "nonexistent.key")
	if err == nil {
		t.Error("expected error for nonexistent files")
	}
}
