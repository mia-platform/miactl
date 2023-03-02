package login

import (
	"fmt"
	"testing"
)

func TestLoginOIDC(t *testing.T) {
	tokens, err := GetTokensWithOIDC("https://test.console.gcp.mia-platform.eu", "okta")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(tokens)
}
