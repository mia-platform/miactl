package renderer

import "time"

// dateFormat should be used when want to render a date in a standard fashion.
// This format is similar to RFC822 but displays the full year instead.
var dateFormat = "02 Jan 2006 15:04 MST"

// FormatDate formats date using custom format similar to RFC822.
func FormatDate(date time.Time) string {
	return date.Format(dateFormat)
}
