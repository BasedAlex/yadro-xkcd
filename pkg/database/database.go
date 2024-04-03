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

func WriteJSON(comics map[string]Page, cfg *config.Config) {
	file, err := json.MarshalIndent(comics, "", " ")
	if err != nil {
		fmt.Println(err)
		return
	}
	
	dst, err := os.Create(filepath.Join(cfg.DbPath, filepath.Base(cfg.DbFile)))
	if err != nil {
		fmt.Println(err)
	}
	defer dst.Close()

	dst.Write(file)
}
