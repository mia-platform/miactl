package renderer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Ref: https://play.golang.org/p/6KBd7Dd3UJy
func TestDateFormatOutput(t *testing.T) {
	date, err := time.Parse(time.UnixDate, "Sat Mar  7 11:06:39 PST 2015")
	require.NoError(t, err)

	require.Equal(t, "07 Mar 2015 11:06 PST", FormatDate(date))
}
