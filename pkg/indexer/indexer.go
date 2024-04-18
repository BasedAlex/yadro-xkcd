package indexer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/basedalex/yadro-xkcd/pkg/database"
	"github.com/basedalex/yadro-xkcd/pkg/words"
)

func Stem(s string) (map[string][]int, error)  {
	smap := make(map[string][]int, 0)

	stems, err := words.Steminator(s)
	if err != nil {
		fmt.Println("error stemming: ", err)
		return nil, err
	}

	for _, stem := range stems {
		var sarr []int
		smap[stem] = sarr
	}


	return smap, nil
}

func LinearSearch(cfg *config.Config, s string) (map[string][]int, error) {
	pathToFile := filepath.Join(cfg.DbPath, filepath.Base(cfg.DbFile))
	
	existingPages := make(map[string]database.Page)

	if _, err := os.Stat(pathToFile); !errors.Is(err, os.ErrNotExist) {
		existingData, err := os.ReadFile(pathToFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(existingData, &existingPages)
		if err != nil {
			return nil, err
		}
	}

	smap, err := Stem(s)

	if err != nil {
		return nil, err
	}

	for pagesIndex, pages := range existingPages {
		intIndex, _ := strconv.Atoi(pagesIndex)
		for key := range smap {
			for _, keyword := range pages.Keywords {
				if keyword == key && len(smap[keyword]) <= 10{
					smap[keyword] = append(smap[keyword], intIndex)
				} 
			}
		}
	}

	return smap, nil 
}