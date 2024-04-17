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
	"yardro-xkcd/pkg/config"
	"yardro-xkcd/pkg/database"
	"yardro-xkcd/pkg/words"
)

const clientTimeout = 10

type rawPage struct {
	Alt        string `json:"alt"`
	Transcript string `json:"transcript"`
	Img        string `json:"img"`
}

func worker(id int, results chan<- map[string]database.Page, client *http.Client, cfg *config.Config, ctx context.Context) {
	j := 1
	count := 0
	for {
		if count == 10 || j == 100 {
			return
		}
		select {
		case <-ctx.Done():
			fmt.Println("shutdown signal received, exiting")
			return
		default:
			// Continue fetching data
		}

		fmt.Printf("worker %d started job %d\n", id, j)
		newPages := make(map[string]database.Page)

		url := fmt.Sprintf("%s%d/info.0.json", cfg.Path, j)

		res, err := client.Get(url)
		if res.StatusCode != http.StatusOK {
			count++
			j++
			continue
		}

		if err != nil {
			fmt.Println("problem getting info from url:", url)
			count++
			j++
			continue
		}
		content, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("nothing found")
			count++
			j++
			continue
		}

		var raw rawPage
		err = json.Unmarshal(content, &raw)
		if err != nil {
			count++
			j++
			fmt.Println(err)
			return
		}

		keywords := raw.Alt + " " + raw.Transcript
		stemmedKeywords, err := words.Steminator(keywords)
		if err != nil {
			count++
			j++
			fmt.Println("error stemming: ", err)
			return
		}

		var page database.Page

		page.Keywords = stemmedKeywords
		page.Img = raw.Img
		index := strconv.Itoa(j)
		newPages[index] = page
		results <- newPages
		j++
	}

}

func SetWorker(cfg *config.Config, ctx context.Context) {
	numJobs := cfg.Parallel
	results := make(chan map[string]database.Page, numJobs)

	client := &http.Client{
		Timeout: clientTimeout * time.Second,
	}

	var wg sync.WaitGroup
	go func() {
		<-ctx.Done()
		fmt.Println("context canceled")
	}()

	wg.Add(5)

	for w := 1; w <= 5; w++ {
		go func(workerID int) {
			defer wg.Done()
			worker(workerID, results, client, cfg, ctx)
		}(w)
	}

	doneCh := make(chan struct{}, 1)

	go func() {
		for result := range results {
			database.SaveComicsConcWithWorkers(cfg, result)
		}
	}()

	go func() {
		wg.Wait()
		close(results)
		close(doneCh)
	}()

	<-doneCh
}
