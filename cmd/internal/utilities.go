package internal

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func CreateConfigurableTestServer(t testing.TB, path string, h http.HandlerFunc, serverCfg map[string]string) (*httptest.Server, error) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc(path, h)
	server := httptest.NewUnstartedServer(mux)

	if serverCfg != nil {
		cert, err := tls.LoadX509KeyPair(serverCfg["cert"], serverCfg["key"])
		if err != nil {
			return nil, err
		}
		server.TLS = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		server.StartTLS()
	} else {
		server.Start()
	}

	return server, nil
}
