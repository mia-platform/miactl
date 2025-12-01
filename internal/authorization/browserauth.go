// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authorization

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"

	"github.com/mia-platform/miactl/internal/client"
	"github.com/mia-platform/miactl/internal/resources"
)

const (
	providerEndpointStringTemplate = "/api/apps/%s/providers/"
	refreshTokenEndpointString     = "/api/refreshtoken"
	authorizeEndpointString        = "/api/authorize"
	callbackEndpointString         = "/oauth/callback"
	// disable gosec for false positive in G101, because it is not an hardcoded credentials...
	getTokenEndpointString = "/api/oauth/token" // #nosec G101
	appIDKey               = "appId"
	providerIDKey          = "providerId"
)

const loginSuccessHTMLPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Authorized</title>
  <script>
    setTimeout(function() {
      window.close();
    }, 1500);
  </script>
  <style>
    body {
      background-color: #eee;
    }
    .placeholder {
      margin: 2em auto;
      padding: 2em;
      background-color: #fff;
      border-radius: 1em;
      position: absolute;
      left: 0;
      right: 0;
    }
  </style>
</head>
<body>
  <div class="placeholder">
    <h1>Authorized</h1>
    <p>If the page don't close by itself, You can close this window.</p>
  </div>
</body>
</html>
`

type Config struct {
	// AppID the app id to use for getting the correct authorization from the Mia-Platform console
	AppID string

	// LocalServerBindAddress hostname and port which the local server binds to, you can use the
	// port 0 for allocate the first free port
	LocalServerBindAddress []string

	// Client the http client to use for sending the request
	Client client.Interface

	// ServerReadyHandler function to call when the local server is ready to receive traffic
	ServerReadyHandler LocalServerReadyHandler
}

// GetToken performs the authorization flow against the Mia-Platform Console
// This perform the following actions:
// 1. Send a request for getting the provider id for the configured AppID
// 2. Start a local server for intercetting the callback
// 3. Open the browser and start the login flow
// 4. Wait for the user authorization
// 5. Exchange the code received via callback
// 6. Return the oauth2 token
func (c *Config) GetToken(ctx context.Context) (*oauth2.Token, error) {
	if len(c.AppID) == 0 {
		return nil, errors.New("missing appId for browser login flow")
	}

	if c.Client == nil {
		return nil, errors.New("cannot setup browser login flow without a valid client")
	}

	return c.startLoginFlow(ctx)
}

// RefreshToken perform a refresh token request and return an error if something went wrong or the new oauth2 token
func (c *Config) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if len(refreshToken) == 0 {
		return nil, errors.New("missing refresh token")
	}

	if c.Client == nil {
		return nil, errors.New("cannot refresh token without a valid client")
	}

	return c.startRefreshFlow(ctx, refreshToken)
}

func (c *Config) startLoginFlow(ctx context.Context) (*oauth2.Token, error) {
	providerID, err := providerIDForApplication(ctx, c.AppID, c.Client)
	if err != nil {
		return nil, err
	}

	listener, err := newListener(c.LocalServerBindAddress)
	if err != nil {
		return nil, err
	}

	startFlowURL := c.Client.Get().
		APIPath(authorizeEndpointString).
		SetParam(appIDKey, c.AppID).
		SetParam(providerIDKey, providerID).URL().String()

	authResponse, err := startLocalServerForToken(ctx, startFlowURL, listener, c.ServerReadyHandler)
	if err != nil {
		return nil, err
	}

	return jwtToken(ctx, authResponse, c.Client)
}

// newListener return the first listener that can be opened or an error if all the ports are already used
func newListener(addresses []string) (net.Listener, error) {
	if len(addresses) == 0 {
		addresses = append(addresses, "127.0.0.1:0")
	}

	for _, address := range addresses {
		listener, err := net.Listen("tcp", address)
		if err != nil {
			continue
		}
		return listener, nil
	}

	return nil, fmt.Errorf("could not listen to any specified addresses: %s", strings.Join(addresses, ", "))
}

// return the providerId for appId to use in the login flow
func providerIDForApplication(ctx context.Context, appID string, client client.Interface) (string, error) {
	response, err := client.
		Get().
		APIPath(fmt.Sprintf(providerEndpointStringTemplate, appID)).
		Do(ctx)

	if err != nil {
		return "", err
	}

	if err := response.Error(); err != nil {
		return "", err
	}

	providers := make([]*resources.AuthProvider, 0)
	if err := response.ParseResponse(&providers); err != nil {
		return "", err
	}
	if len(providers) == 0 {
		return "", fmt.Errorf("no providers found for %s", appID)
	}

	// TODO: in case of multiple providers  made the user choose the one he wants
	// Temporarily disable gosec G602, which produces a false positive.
	// See https://github.com/securego/gosec/issues/1005.
	return providers[0].ID, nil // #nosec G602
}

// startLocalServerForToken start a server for listening to callback requests for the login flow
func startLocalServerForToken(ctx context.Context, startFlowURL string, listener net.Listener, readyFn LocalServerReadyHandler) (*authResponse, error) {
	shutdownChannel := make(chan int)
	responseChannel := make(chan *authResponse)

	httpHandler := &httpHandler{
		startFlowURL:    startFlowURL,
		responseChannel: responseChannel,
	}

	server := http.Server{
		Handler:           httpHandler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	hostAddress := "http://" + listener.Addr().String()
	var taskGroup errgroup.Group
	var respOut *authResponse
	taskGroup.Go(func() error {
		defer close(responseChannel)

		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	taskGroup.Go(func() error {
		defer close(shutdownChannel)

		select {
		case response, ok := <-responseChannel:
			if ok {
				respOut = response
			}
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	taskGroup.Go(func() error {
		<-shutdownChannel

		// gracefully shutdown the server with a timout of half a second, if not, kill it
		timeoutCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()
		if err := server.Shutdown(timeoutCtx); err != nil {
			_ = server.Close()
			return nil
		}
		return nil
	})
	taskGroup.Go(func() error {
		if readyFn == nil {
			return nil
		}
		return readyFn(hostAddress)
	})

	if err := taskGroup.Wait(); err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}

	return respOut, respOut.Err
}

// jwtToken exhange code and status of the authorization response for a jwt token
func jwtToken(ctx context.Context, response *authResponse, client client.Interface) (*oauth2.Token, error) {
	request := &resources.JWTTokenRequest{Code: response.Code, State: response.State}
	bodydata, err := resources.EncodeResourceToJSON(&request)
	if err != nil {
		return nil, err
	}

	jwtResponse, err := client.
		Post().
		APIPath(getTokenEndpointString).
		Body(bodydata).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	return parseJWTResponse(jwtResponse)
}

type authResponse struct {
	Err   error
	Code  string `json:"code"`
	State string `json:"state"`
}

type httpHandler struct {
	startFlowURL    string
	responseChannel chan<- *authResponse
	syncOneResponse sync.Once
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	switch {
	case r.Method == http.MethodGet && r.URL.Path == callbackEndpointString && q.Get("error") != "":
		h.syncOneResponse.Do(func() {
			q := r.URL.Query()
			errorCode, errorDescription := q.Get("error"), q.Get("error_description")

			http.Error(w, "authorization error", http.StatusInternalServerError)
			h.responseChannel <- &authResponse{
				Err: fmt.Errorf("authorization error from server: %s %s", errorCode, errorDescription),
			}
		})
	case r.Method == http.MethodGet && r.URL.Path == callbackEndpointString && q.Get("code") != "":
		h.syncOneResponse.Do(func() {
			q := r.URL.Query()
			code, state := q.Get("code"), q.Get("state")
			w.Header().Add("Content-Type", "text/html")
			if _, err := w.Write([]byte(loginSuccessHTMLPage)); err != nil {
				fmt.Fprintf(os.Stderr, "error during writing HTML page: %s", err)
			}
			h.responseChannel <- &authResponse{Code: code, State: state}
		})
	case r.Method == http.MethodGet && r.RequestURI == "/":
		http.Redirect(w, r, h.startFlowURL, http.StatusFound)
	default:
		http.NotFound(w, r)
		h.responseChannel <- &authResponse{
			Err: fmt.Errorf("callback not recognized: %s %s", r.Method, r.URL.Path),
		}
	}
}

func open(url string) error {
	return openBrowser(url)
}

func parseJWTResponse(jwtResponse *client.Response) (*oauth2.Token, error) {
	if jwtResponse.Error() != nil {
		return nil, jwtResponse.Error()
	}

	jwt := new(resources.UserToken)
	if err := jwtResponse.ParseResponse(&jwt); err != nil {
		return nil, err
	}

	return jwt.JWTToken(), nil
}

func (c *Config) startRefreshFlow(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	bodydata, err := resources.EncodeResourceToJSON(&resources.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		return nil, err
	}

	jwtResponse, err := c.Client.
		Post().
		APIPath(refreshTokenEndpointString).
		Body(bodydata).
		Do(ctx)

	if err != nil {
		return nil, err
	}
	return parseJWTResponse(jwtResponse)
}
