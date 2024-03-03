package utils

import (
	"regexp"
	"strings"
)

var validEmailOnlyValidCharacters = regexp.MustCompile(`[^a-z0-9._%+\-]+`)

func TransformIntoValidEmailName(name string) string {
	name = strings.ToLower(name)
	name = validEmailOnlyValidCharacters.ReplaceAllString(name, "_")
	return name
}
