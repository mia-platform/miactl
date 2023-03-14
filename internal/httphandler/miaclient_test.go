package httphandler

import (
	"testing"

	"github.com/mia-platform/miactl/internal/cmd/login"
	"github.com/stretchr/testify/require"
)

func TestClientBuilding(t *testing.T) {
	mExpected := MiaClient{
		request: Request{
			url: "url",
		},
		browser:    login.Browser{},
		providerId: "id",
		clientUrl:  "url",
	}

	m2Expected := MiaClient{
		request: Request{
			url: "url2",
		},
	}

	b := login.Browser{}
	r := Request{
		url: "url",
	}
	r2 := Request{
		url: "url2",
	}

	m := NewMiaClientBuilder().
		withAuthentication(b, "id", "url").
		withRequest(r)

	m2 := NewMiaClientBuilder().
		withRequest(r2)

	require.Equal(t, *m, mExpected)
	require.Equal(t, *m2, m2Expected)

}
