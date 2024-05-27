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

func Steminator(str string) ([]string, error) {
	pronouns := map[string]interface{}{
		"i": true, "you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
		"me": true, "him": true, "her": true, "us": true, "them": true, "my": true,
		"myself": true, "yourself": true, "himself": true, "herself": true, "itself": true, "ourselves": true, "themselves": true,
	}

	prepositions := map[string]interface{}{
		"aboard": true, "about": true, "above": true, "across": true, "after": true, "against": true, "along": true, "amid": true, "among": true, "around": true, "as": true, "at": true, "before": true, "behind": true, "below": true, "beneath": true, "beside": true, "between": true, "beyond": true, "but": true, "by": true, "concerning": true, "considering": true, "despite": true, "down": true, "during": true, "except": true, "excepting": true, "for": true, "from": true, "in": true, "inside": true, "into": true, "like": true, "near": true, "of": true, "off": true, "on": true, "onto": true, "out": true, "outside": true, "over": true, "past": true, "regarding": true, "round": true, "since": true, "through": true, "throughout": true, "till": true, "to": true, "toward": true, "under": true, "underneath": true, "until": true, "up": true, "upon": true, "with": true, "within": true, "without": true, "alt": true,
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
