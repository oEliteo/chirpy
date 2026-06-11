package main

import (
	"strings"
)

var bannedWordsMap = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func filterChirp(body string, bannedWords map[string]struct{}) string {
	words := make([]string, 0)
	for word := range strings.FieldsSeq(body) {
		_, exists := bannedWords[strings.ToLower(word)]
		if exists {
			words = append(words, "****")
		} else {
			words = append(words, word)
		}
	}
	return strings.Join(words, " ")
}
