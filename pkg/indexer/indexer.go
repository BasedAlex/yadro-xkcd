package indexer

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/basedalex/yadro-xkcd/pkg/database"
	"github.com/basedalex/yadro-xkcd/pkg/words"
)

func Stem(s string) (map[string][]int, error) {
	smap := make(map[string][]int)

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
		intIndex, err := strconv.Atoi(pagesIndex)
		if err != nil {
			fmt.Println("couldn't create index", err)
			continue
		}
		for key := range smap {
			for _, keyword := range pages.Keywords {
				if keyword == key && len(smap[keyword]) < 10 {
					smap[keyword] = append(smap[keyword], intIndex)
				}
			}
		}
	}

	return smap, nil
}

func Reverse(cfg *config.Config) error {
	pathToFile := filepath.Join(cfg.DbPath, filepath.Base(cfg.DbFile))
	pathToIndex := filepath.Join(cfg.DbPath, filepath.Base("indexer.json"))

	existingPages := make(map[string]database.Page)
	indexPages := make(map[string][]int)

	if _, err := os.Stat(pathToFile); !errors.Is(err, os.ErrNotExist) {
		fmt.Println(err)
		existingData, err := os.ReadFile(pathToFile)
		if err != nil {
			return err
		}
		err = json.Unmarshal(existingData, &existingPages)
		if err != nil {
			return err
		}
	}


	for pagesIndex, pages := range existingPages {
		intIndex, err := strconv.Atoi(pagesIndex)
		if err != nil {
			fmt.Println("couldn't create index", err)
			continue
		}
		for _, keyword := range pages.Keywords {
			indexPages[keyword] = append(indexPages[keyword], intIndex)
		}
	}

	f, err := os.OpenFile(pathToIndex, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("error opening file:", err)
		return err
	}
	defer f.Close()

	file, err := json.Marshal(indexPages)
	if err != nil {
		log.Println("error marshalling JSON:", err)
		return err
	}
	_, err = f.Write(file)
	if err != nil {
		log.Println("error writing to file:", err)
		return err
	}

	return nil
}

func InvertSearch(cfg *config.Config, s string) (map[string][]int, error) {
	pathToIndex := filepath.Join(cfg.DbPath, filepath.Base("indexer.json"))
	existingIndexPages := make(map[string][]int)

	if _, err := os.Stat(pathToIndex); !errors.Is(err, os.ErrNotExist) {
		fmt.Println(err)
		existingData, err := os.ReadFile(pathToIndex)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(existingData, &existingIndexPages)
		if err != nil {
			return nil, err
		}
	}

	smap, err := Stem(s)
	if err != nil {
		return nil, err
	}

	for key := range smap {
		pages := existingIndexPages[key]
		smap[key] = append(smap[key], pages...)
		if len(smap[key]) > 10 {
			smap[key] = append(smap[key][:9], smap[key][10])
		}
	}
	
	return smap, nil
}
