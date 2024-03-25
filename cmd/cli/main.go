package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/kljensen/snowball"
)


func main() {
	var str string

	flag.StringVar(&str, "str", "follower brings bunch of questions follow", "input a string to be stemmed")
	flag.Parse()

	words := strings.Split(str, " ")
	filterMap := dictionary()
	res := make([]string, 0)

	for i := 0; i < len(words); i++ {
		stemmed, err := snowball.Stem(words[i], "english", true)
		if err != nil {
			log.Fatal(err)
		}
		if !filterMap[stemmed] {
			res = append(res, stemmed)
		}
		filterMap[stemmed] = true
	}

	fmt.Println(strings.Join(res, " ")) 
}

