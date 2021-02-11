package cmd

import (
	"fmt"
)

var apiKeyValue = "foo"
var apiKeyFlag = fmt.Sprintf(`--apiKey="%s"`, apiKeyValue)
var sidValue = "my-sid"
var apiCookieFlag = fmt.Sprintf(`--apiCookie="sid=%s"`, sidValue)
var apiBaseURLValue = "https://local.io/base-path/"
var apiBaseURLFlag = fmt.Sprintf(`--apiBaseUrl=%s`, apiBaseURLValue)
