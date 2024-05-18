package helpers

import (
	"golang.org/x/exp/utf8string"
)

func TruncateText(s string, max int) string {

	if len(s) == 0 {
		return "-"
	}
	us := utf8string.NewString(s)
	if us.RuneCount() >= max {
		return us.Slice(0, max) + "..."
	}
	return s
}
