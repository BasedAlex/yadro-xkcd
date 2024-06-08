package words

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
	"github.com/kljensen/snowball/english"
)

type Stemmer interface {
	Steminator(str string) ([]string, error)
}

func Steminator(str string) ([]string, error) {
	pronouns := map[string]struct{}{
		"i": {}, "you": {}, "he": {},"she": {}, "it": {},"we": {},
		"they": {},"me": {},"him": {},"her": {}, "us": {}, "them": {}, "my": {},
		"myself": {}, "yourself": {}, "himself": {}, "herself": {}, "itself": {}, "ourselves": {}, "themselves": {},
	}
	

	prepositions := map[string]struct{}{
		"aboard": {}, "about": {}, "above": {}, "across": {}, "after": {}, "against": {}, "along": {}, "amid": {}, "among": {}, "around": {}, "as": {}, "at": {}, "before": {}, "behind": {}, "below": {}, "beneath": {}, "beside": {}, "between": {}, "beyond": {}, "but": {}, "by": {}, "concerning": {}, "considering": {}, "despite": {}, "down": {}, "during": {}, "except": {}, "excepting": {}, "for": {}, "from": {}, "in": {}, "inside": {}, "into": {}, "like": {}, "near": {}, "of": {}, "off": {}, "on": {}, "onto": {}, "out": {}, "outside": {}, "over": {}, "past": {}, "regarding": {}, "round": {}, "since": {}, "through": {}, "throughout": {}, "till": {}, "to": {}, "toward": {}, "under": {}, "underneath": {}, "until": {}, "up": {}, "upon": {}, "with": {}, "within": {}, "without": {}, "alt": {},
	}

	if len(str) == 0 {
		return nil, errors.New("please provide a string to be stemmed")
	}

	seen := make(map[string]int)

	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}

	regex := regexp.MustCompile("[^a-zA-Z]+")
	newStr := regex.ReplaceAllString(str, " ")

	words := strings.Split(newStr, " ")
	var res string

	for _, word := range words {
		if english.IsStopWord(word) {
			continue
		}

		stemmed, err := snowball.Stem(word, "english", true)
		if err != nil {
			return nil, err
		}

		if slices.Contains([]byte(stemmed), 39) {
			continue
		}

		if _, ok := pronouns[stemmed]; ok {
			continue
		}

		if _, ok := prepositions[stemmed]; ok {
			continue
		}

		if seen[stemmed] == 0 {
			seen[stemmed]++
			res = res + " " + stemmed
		}
	}

	if len(res) == 0 {
		return nil, errors.New("result is empty, please provide a better string")
	}

	result := strings.FieldsFunc(res, f)

	return result, nil
}
