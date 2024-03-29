package main

import (
	"flag"
	"fmt"
	"log"
	"slices"
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
	"github.com/kljensen/snowball/english"
)


func main() {

	pronouns := map[string]bool{
		"i": true, "you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
		"me": true, "him": true, "her": true, "us": true, "them": true,
		"myself": true, "yourself": true, "himself": true, "herself": true, "itself": true, "ourselves": true, "themselves": true,
	}

	prepositions := map[string]bool{
		"aboard": true, "about": true, "above": true, "across": true, "after": true, "against": true, "along": true, "amid": true, "among": true, "around": true, "as": true, "at": true, "before": true, "behind": true, "below": true, "beneath": true, "beside": true, "between": true, "beyond": true, "but": true, "by": true, "concerning": true, "considering": true, "despite": true, "down": true, "during": true, "except": true, "excepting": true, "for": true, "from": true, "in": true, "inside": true, "into": true, "like": true, "near": true, "of": true, "off": true, "on": true, "onto": true, "out": true, "outside": true, "over": true, "past": true, "regarding": true, "round": true, "since": true, "through": true, "throughout": true, "till": true, "to": true, "toward": true, "under": true, "underneath": true, "until": true, "up": true, "upon": true, "with": true, "within": true, "without": true,
	}

	var str string

	flag.StringVar(&str, "s", "", "input a string to be stemmed")
	flag.Parse()

	if len(str) == 0 {
		log.Fatalln("Please provide a string to be stemmed")
	}
	
	seen := make(map[string]int)

	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}

	words := strings.Split(str, " ")
	var res string

	for _, word := range words {

		if english.IsStopWord(word) {
			continue
		}

		stemmed, err := snowball.Stem(word, "english", true)
		if err != nil {
			log.Fatal(err)
		}

		// we don't need any word that includes apostrophes
		if slices.Contains([]byte(stemmed), 39) || pronouns[stemmed] || prepositions[stemmed] {
			continue
		}


		if seen[stemmed] == 0 {
			seen[stemmed]++
			res = res + " " + stemmed
		}

	}

	if len(res) == 0 {
		log.Fatalln("result is empty, please provide a better string")
	}

	fmt.Println(strings.Join(strings.FieldsFunc(res, f), " ")) 
}

