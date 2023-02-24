package context

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const missingContextName = "missing"
const contextName1 = "context1"

func TestContextLookUpEmptyContextMap(t *testing.T) {
	viper.SetConfigType("yaml")
	config := ``
	err := viper.ReadConfig(strings.NewReader(config))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}
	err = contextLookUp(missingContextName)
	require.EqualError(t, err, "no context specified in config file")
}

func TestContextLookUp(t *testing.T) {
	viper.SetConfigType("yaml")
	config := `contexts:
  context1:
    apibaseurl: http://url
    companyid: "123"
    projectid: "123"`
	err := viper.ReadConfig(strings.NewReader(config))
	if err != nil {
		t.Fatalf("unexpected error reading config: %v", err)
	}
	err = contextLookUp(missingContextName)
	require.EqualError(t, err, "context missing does not exist")

	err = contextLookUp(contextName1)
	require.Nil(t, err)
}
