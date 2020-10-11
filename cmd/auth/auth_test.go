package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthCommand(t *testing.T) {
	cmd := NewAuthCmd()

	require.Equal(t, "auth <command>", cmd.Use)
	require.Equal(t, "login", cmd.Short)
}
