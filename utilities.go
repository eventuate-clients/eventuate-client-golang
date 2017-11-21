package eventuate

import "strings"

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func lastPart(input, delim string) string {
	result := ""
	if delim == "" {
		delim = "."
	}
	inputParts := strings.Split(input, delim)
	if len(inputParts) == 0 {
		return result
	}
	return inputParts[len(inputParts)-1]
}
