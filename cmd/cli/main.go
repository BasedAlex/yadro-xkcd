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
	var str string

	flag.StringVar(&str, "s", "follower brings bunch of questions follow", "input a string to be stemmed")
	flag.Parse()
	
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
		fmt.Println(stemmed)

		// we don't need any word that includes apostrophes
		if slices.Contains([]byte(stemmed), 39) {
			continue
		}

		if seen[stemmed] == 0 {
			seen[stemmed]++
			res = res + " " + stemmed
		}

	}


	fmt.Println(strings.Join(strings.FieldsFunc(res, f), " ")) 
}

