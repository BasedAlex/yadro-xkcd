package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"yardro-xkcd/pkg/config"
)

type Page struct {
	Img      string `json:"img"`
	Keywords []string `json:"keywords"`
}

func SaveComics(comics map[string]Page, cfg *config.Config) error {

	if cfg.Print {
		for i, v := range comics {
			fmt.Printf("Index:%s\nImage: %s\nKeywords:%v\n", i, v.Img, v.Keywords)
		}
	}

	file, err := json.Marshal(comics)
	if err != nil {
		return err
	}
	
	dst, err := os.Create(filepath.Join(cfg.DbPath, filepath.Base(cfg.DbFile)))
	if err != nil {
		return err
	}
	defer dst.Close()

	dst.Write(file)
	return nil
}
