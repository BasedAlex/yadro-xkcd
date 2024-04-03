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


func ReadJSON(cfg *config.Config) (map[string]Page, error) {

	file, err := os.Open(filepath.Join(cfg.DbPath, filepath.Base(cfg.DbFile)))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	comics := make(map[string]Page, 0)

	dist, err := os.ReadFile(filepath.Join(cfg.DbPath, filepath.Base(cfg.DbFile)))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	json.Unmarshal(dist, &comics)

	return comics, nil
}