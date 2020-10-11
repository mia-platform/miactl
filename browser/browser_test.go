package browser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenBrowserByOS(t *testing.T) {
	url := "http://local.url/to-open"

	type args struct {
		goos string
		url  string
	}
	type expected struct {
		err  error
		args []string
	}
	tests := []struct {
		name     string
		args     args
		expected expected
	}{
		{
			name:     "linux",
			args:     args{"linux", url},
			expected: expected{args: []string{"xdg-open", url}},
		},
		{
			name:     "darwin",
			args:     args{"darwin", url},
			expected: expected{args: []string{"open", url}},
		},
		{
			name:     "windows",
			args:     args{"windows", url},
			expected: expected{args: []string{"rundll32", "url.dll,FileProtocolHandler", url}},
		},
		{
			name:     "unsupported",
			args:     args{"not-supported-os", url},
			expected: expected{err: fmt.Errorf("unsupported platform not-supported-os")},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd, err := commandForOS(test.args.goos, test.args.url)

			require.Equal(t, test.expected.err, err)
			if test.expected.args != nil {
				require.Equal(t, test.expected.args, cmd.Args)
			}
		})
	}
}
