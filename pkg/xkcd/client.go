package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/basedalex/yadro-xkcd/pkg/database"
	"github.com/basedalex/yadro-xkcd/pkg/words"
)

const clientTimeout = 10

type rawPage struct {
	Num		int `json:"num"`
	Alt        string `json:"alt"`
	Transcript string `json:"transcript"`
	Img        string `json:"img"`
}

func task(ctx context.Context, results chan<- database.Page, client *http.Client, cfg *config.Config,  intCh chan int) {

	for w := range intCh {
		url := fmt.Sprintf("%s%d/info.0.json", cfg.Path, w)
		req, err  := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			fmt.Println("couldn't make request:", err)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			fmt.Println("problem getting info from url:", url, err)
			return
		}
		if res.StatusCode != http.StatusOK {
			fmt.Println("couldn't get info from url:", url)
			return
		}

		content, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("nothing found")
			continue
		}

		var raw rawPage
		err = json.Unmarshal(content, &raw)
		if err != nil {
			fmt.Println(err)
			continue
		}

		keywords := raw.Alt + " " + raw.Transcript
		stemmedKeywords, err := words.Steminator(keywords)
		if err != nil {
			fmt.Println("error stemming: ", err)
			return
		}

		var page database.Page

		page.Keywords = stemmedKeywords
		page.Img = raw.Img
		page.Index = strconv.Itoa(raw.Num)
		results <- page
	}
}

func SetWorker(ctx context.Context, cfg *config.Config) {
	results := make(chan database.Page, cfg.Parallel)
	intCh := make(chan int)

	client := &http.Client{
		Timeout: clientTimeout * time.Second,
	}

	var wg sync.WaitGroup

	for i := 1; i <= cfg.Parallel; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			task(ctx, results, client, cfg, intCh)
		}(i)
	}

	resultDoneCh := make(chan struct{})
	generatorDoneCh := make(chan struct{})
	
	go func() {
		defer close(resultDoneCh)
		for result := range results {
			database.SaveComics(cfg, result)
		}
	}()

	go func() {
		wg.Wait()
		close(results)
		close(generatorDoneCh)
	} ()
	
loop:
	for i := 0; ;i++ {
		if i == 404 {
			continue
		}
		select {
		case <-generatorDoneCh:
			break loop
		case intCh <- i:
		}
		
	}

	<-resultDoneCh
	fmt.Println("finished fetching data...")
}
