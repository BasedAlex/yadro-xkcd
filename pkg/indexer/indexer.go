package indexer

import (
	"fmt"
	"yardro-xkcd/pkg/words"
)

func Stem(s string) (map[string][]int, error)  {
	sm := make(map[string][]int, 0)

	stems, err := words.Steminator(s)
	if err != nil {
		fmt.Println("error stemming: ", err)
		return nil, err
	}

	for _, stem := range stems {
		var sarr []int
		sm[stem] = sarr
	}


	return sm, nil
}