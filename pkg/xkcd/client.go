package xkcd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"yardro-xkcd/pkg/config"
	"yardro-xkcd/pkg/database"
	"yardro-xkcd/pkg/words"
)

type rawPage struct {
	Alt string `json:"alt"`
	Transcript string `json:"transcript"`
	Img string `json:"img"`
}


func WriteToDB(cfg *config.Config) error {
	
	newPages, err := GetPages(cfg)
	if err != nil {
		return err
	}
	
	err = database.SaveComics(newPages, cfg)
	if err != nil {
		return err
	}
	return nil
}


const clientTimeout = 10

// эта функция будет вызвана при создании приложения
func GetPages(cfg *config.Config) (map[string]database.Page, error) {
	newPages := make(map[string]database.Page)

	counter := 0

	client := &http.Client{
		Timeout: clientTimeout * time.Second,
	}

	defer client.CloseIdleConnections()
	
	for i := 1; i <= cfg.Limit; i++ {
		
		url := fmt.Sprintf("%s%d/info.0.json", cfg.Path, i)
		
		res, err := client.Get(url)

		// увеличиваем счётчик только при клиентских ошибках
		if res.StatusCode != http.StatusOK {
			counter++
		}

		if err != nil {
			fmt.Println("problem getting info from url:", url)

			// возвращаемся если слишком часто получаем ошибки, т.к. либо на сервере проблема, либо кончились комиксы 
			if counter > 10 {
				fmt.Println("too many missed pages: ", counter)
				return nil, err
			}
			continue
		}

		content, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("nothing found")
			return nil, err
		}

		var raw rawPage
		err = json.Unmarshal(content, &raw)
		if err != nil {
			return nil, err
		}

		keywords := raw.Alt + " " + raw.Transcript

		stemmedKeywords, err := words.Steminator(keywords)
		if err != nil {
			fmt.Println("error stemming: ", err)
			return nil, err
		}

		var page database.Page

		page.Keywords = stemmedKeywords
		page.Img = raw.Img
		index := strconv.Itoa(i)
		newPages[index] = page
	}
	return newPages, nil
}
