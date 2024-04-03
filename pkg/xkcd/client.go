package xkcd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"yardro-xkcd/pkg/config"
	"yardro-xkcd/pkg/database"
	"yardro-xkcd/pkg/words"
)

func Routes(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", ping)
	mux.HandleFunc("/getPages", func(w http.ResponseWriter, r *http.Request) {
		getPages(w, r, cfg)
	})

	return mux
}


func ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello")
}

type prepare struct {
	Alt string `json:"alt"`
	Transcript string `json:"transcript"`
	Img string `json:"img"`
}

// эта функция будет вызвана при создании приложения
func GetPages(cfg *config.Config) {
	newPages := make(map[string]database.Page)

	for i := cfg.Start; i <= cfg.Limit; i++ {
		link := fmt.Sprintf("%s%d/info.0.json", cfg.Path, i)
		res, err := http.Get(link)
		if err != nil {
			fmt.Println("problem getting info from link: ", link)
			return
		}

		content, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("nothing found")
			return
		}
		
		var prep prepare
		json.Unmarshal(content, &prep)

		keywords := prep.Alt + " " + prep.Transcript

		stemmedKeywords, err := words.Steminator(keywords)
		if err != nil {
			fmt.Println("error stemming: ", err)
			return
		}

		var page database.Page

		page.Keywords = stemmedKeywords
		page.Img = prep.Img
		index := strconv.Itoa(i)
		newPages[index] = page
	}
	if cfg.Print {
		for i, v := range newPages {
			fmt.Printf("Index:%s\nImage: %s\nKeywords:%v\n", i, v.Img, v.Keywords)
		}
	}
	database.WriteJSON(newPages, cfg)
}


// эта функция будет вызвана при запросе на /getPages
func getPages(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// жалко удалять ResponseWriter оставлю потомкам :^)
	_ = w

	if r.Method != http.MethodGet {
		fmt.Println("method not implemented")
		return
	}

	GetPages(cfg)
}