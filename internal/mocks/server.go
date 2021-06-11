package mocks

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type ServerConfig struct {
	Endpoint       string
	Method         string
	QueryParams    map[string]interface{}
	RequestHeaders map[string]string
	RequestBody    interface{}
	Reply          interface{}
	ReplyStatus    int
}

type ServerConfigs []ServerConfig

type CertificatesConfig struct {
	CertPath string
	KeyPath  string
}

func HTTPServer(t *testing.T, cfgs ServerConfigs, tlsCfg *CertificatesConfig) (*httptest.Server, error) {
	t.Helper()

	mux := http.NewServeMux()

	for _, c := range cfgs {
		expectedRequestBody, _ := json.Marshal(c.RequestBody)
		serverResponse, _ := json.Marshal(c.Reply)

		mux.HandleFunc(c.Endpoint, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			require.Equal(t, c.Endpoint, req.URL.Path, "correct endpoint")
			require.Equal(t, c.Method, req.Method)
			for key, expectedValue := range c.QueryParams {
				require.Equal(t, expectedValue, req.FormValue(key), "query param is the expected one")
			}
			for key, expectedValue := range c.RequestHeaders {
				require.Equal(t, expectedValue, req.Header.Get(key), "header value is the expected one")
			}

			// check body only if required (e.g. not needed for GET/DELETE method)
			if c.RequestBody != nil {
				rawRequestBody, _ := ioutil.ReadAll(req.Body)
				// remove the EOF/EOL (0xa byte) added by the ioutil.ReadAll to the request body bytes
				requestBody := rawRequestBody[:len(rawRequestBody)-1]

				require.Equal(t, expectedRequestBody, requestBody)
			}

			w.WriteHeader(c.ReplyStatus)
			if c.Reply != nil {
				w.Write(serverResponse)
			}
		}))
	}

	server := httptest.NewUnstartedServer(mux)

	if tlsCfg != nil {
		cert, err := tls.LoadX509KeyPair(tlsCfg.CertPath, tlsCfg.KeyPath)
		if err != nil {
			t.Fatalf("error loading TLS conf: %v", err)
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
