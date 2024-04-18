package database

import (
	"encoding/json"
	"errors"
	"github.com/basedalex/yadro-xkcd/pkg/config"
	"log"
	"os"
	"path/filepath"
)

type Page struct {
	Img      string   `json:"img"`
	Keywords []string `json:"keywords"`
}

func SaveComics(cfg *config.Config, comics map[string]Page) {
	pathToFile := filepath.Join(cfg.DbPath, filepath.Base(cfg.DbFile))

	existingPages := make(map[string]Page)
	if _, err := os.Stat(pathToFile); !errors.Is(err, os.ErrNotExist) {
		existingData, err := os.ReadFile(pathToFile)
		if err != nil {
			log.Println(err)
			return
		}
		err = json.Unmarshal(existingData, &existingPages)
		if err != nil {
			log.Println(err)
			return
		}
	}

	f, err := os.OpenFile(pathToFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("error opening file:", err)
		return
	}
	defer f.Close()
	for key, value := range comics {
		existingPages[key] = value
	}

	file, err := json.Marshal(existingPages)
	if err != nil {
		log.Println("error marshalling JSON:", err)
		return
	}
	_, err = f.Write(file)
	if err != nil {
		log.Println("error writing to file:", err)
		return
	}
}
