package markdown

import (
	"strings"
	"unicode"
)

func Slugify(heading string) string {
	lower := strings.ToLower(heading)
	var buf strings.Builder
	lastHyphen := false
	for _, r := range lower {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
			lastHyphen = false
		} else {
			if !lastHyphen {
				buf.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	slug := buf.String()
	slug = strings.Trim(slug, "-")
	return slug
}
