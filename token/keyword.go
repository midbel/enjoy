package token

import (
	"slices"
)

var keywords = []string{
	"as",
	"let",
	"const",
	"for",
	"in",
	"of",
	"import",
	"export",
	"from",
	"default",
	"if",
	"else",
	"switch",
	"case",
	"function",
	"return",
	"break",
	"continue",
	"try",
	"catch",
	"finally",
	"while",
	"do",
	"null",
	"undefined",
	"true",
	"false",
	"delete",
	"typeof",
	"instanceof",
	"new",
}

func IsKeyword(str string) bool {
	if !slices.IsSorted(keywords) {
		slices.Sort(keywords)
	}
	_, ok := slices.BinarySearch(keywords, str)
	return ok
}
