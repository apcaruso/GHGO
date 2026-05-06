package vocab

import "strings"

func NormalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func Key(s string) string {
	s = strings.ToLower(NormalizeSpace(s))

	var b strings.Builder
	for _, r := range s {
		switch r {
		case '-', '_', '/', '\\', '.', ',', ';', ':', '(', ')', '[', ']', '{', '}', '\'', '"':
			b.WriteRune(' ')
		default:
			b.WriteRune(r)
		}
	}

	return NormalizeSpace(b.String())
}
