package testutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"
)

var testToken = ""

type Test struct {
	Key string `json:"key"`
}

type MockValidToken struct{}
type MockExpiredToken struct{}
type MockFailAuth struct{}
type MockFailRefresh struct{}

func (a *MockValidToken) Authenticate() (string, error) {
	return "valid_token", nil
}

func (a *MockExpiredToken) Authenticate() (string, error) {
	if testToken == "" {
		testToken = "expired_token"
	} else {
		testToken = "valid_token"
	}
	return testToken, nil
}

func (a *MockFailAuth) Authenticate() (string, error) {
	return "", fmt.Errorf("authentication failed")
}

func (a *MockFailRefresh) Authenticate() (string, error) {
	if testToken == "" {
		testToken = "expired_token"
		return testToken, nil
	}
	return "", fmt.Errorf("authentication failed")
}

func GenerateMockCert(t *testing.T) (string, string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")}, // IP SAN for 127.0.0.1
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	testDirPath := t.TempDir()
	testCertPath := path.Join(testDirPath, "testcert.pem")
	certOut, err := os.Create(testCertPath)
	if err != nil {
		return "", "", err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	testKeyPath := path.Join(testDirPath, "testkey.pem")
	keyOut, err := os.Create(testKeyPath)
	if err != nil {
		panic(err)
	}
	defer keyOut.Close()
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		panic(err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	return testCertPath, testKeyPath, nil
}

func CreateMockServer() *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/notfound" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			auth := r.Header.Get("Authorization")
			switch auth {
			case "Bearer valid_token":
				w.WriteHeader(http.StatusOK)
			default:
				w.WriteHeader(http.StatusUnauthorized)
			}
		}
		if r.RequestURI == "/invalidbody" {
			w.Write([]byte(`invalid json`))
		} else if r.RequestURI == "/getprojects" {
			w.Write([]byte(`[{"_id": "123"}]`))
		}
	}))
	return server
}
