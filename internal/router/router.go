package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/basedalex/yadro-xkcd/internal/db"
	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/basedalex/yadro-xkcd/pkg/words"
	log "github.com/sirupsen/logrus"
)

type HTTPResponse struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type xkcdService interface {
	SaveComics(ctx context.Context, cfg *config.Config, comics db.Page) error
	Reverse(ctx context.Context, cfg *config.Config) error
	InvertSearch(ctx context.Context, cfg *config.Config, s string) (map[string][]int, error)
}

type Handler struct {
	service xkcdService
	cfg *config.Config
}

func NewServer(ctx context.Context, cfg *config.Config, service xkcdService) error {
	srv := &http.Server{
		Addr: ":" + cfg.SrvPort,
		Handler: newRouter(cfg, service),
		ReadHeaderTimeout: 3 * time.Second,
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	go func() {
		<-ctx.Done()

		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Warn(err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error with the server: %w", err)
	}

	return nil
}

func newRouter(cfg *config.Config, service xkcdService) *http.ServeMux {
	handler := &Handler{
		cfg: cfg,
		service: service,
	}

	mux := http.NewServeMux()

	handler.NewScheduler(context.Background())

	mux.HandleFunc("/", ping)
	mux.HandleFunc("/update", handler.updatePics)
	mux.HandleFunc("/pics", handler.getPics)

	return mux
}

func ping(w http.ResponseWriter, r *http.Request) {
	writeOkResponse(w, http.StatusOK, "Hello World")
}

func (h *Handler) updatePics(w http.ResponseWriter, r *http.Request) {
	h.SetWorker(r.Context(), h.cfg)
	err := h.service.Reverse(r.Context(), h.cfg)
	if err != nil {
		writeErrResponse(w, http.StatusInternalServerError, err)
	}
	writeOkResponse(w, http.StatusOK, "Updated comics...")
}

func (h *Handler) getPics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	search := query.Get("search")

	if search == "" {
		writeErrResponse(w, http.StatusBadRequest, fmt.Errorf("no comics to search"))
		return
	}

	sm, err := h.service.InvertSearch(r.Context(), h.cfg, search)
	if err != nil {
		writeErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	links := make([]string, 0)

	for _, values := range sm {
		for _, v := range values {
			link := fmt.Sprintf("%s%d", h.cfg.Path, v)
			links = append(links, link)
		}
	}

	writeOkResponse(w, http.StatusOK, links)
}

func (h *Handler) NewScheduler(ctx context.Context) {
    ticker := time.NewTicker(24 * time.Hour)

    h.runUpdate(ctx)

    go func() {
        for {
            select {
            case <-ticker.C:
                h.runUpdate(ctx)
            }
        }
    }()

    select {}

}

func (h *Handler) runUpdate(ctx context.Context) {
	h.SetWorker(ctx, h.cfg)

    log.Println("Last updated at:", time.Now())
}

const clientTimeout = 10


type rawPage struct {
	Num		int `json:"num"`
	Alt        string `json:"alt"`
	Transcript string `json:"transcript"`
	Img        string `json:"img"`
}

func (h *Handler) SetWorker(ctx context.Context, cfg *config.Config) {
	results := make(chan db.Page, cfg.Parallel)
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
			err := h.service.SaveComics(ctx, cfg, result)
			if err != nil {
				log.Info(err)
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
		close(generatorDoneCh)
	} ()
	
loop:
	for i := 1; ;i++ {
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

func task(ctx context.Context, results chan<- db.Page, client *http.Client, cfg *config.Config, intCh chan int) {
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

		var page db.Page

		page.Keywords = stemmedKeywords
		page.Img = raw.Img
		page.Index = strconv.Itoa(raw.Num)
		results <- page
	}
}

func writeOkResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	log.Infof("successful request with statusCode %d and data type %T", statusCode, data)
	if data != nil {
		err := json.NewEncoder(w).Encode(HTTPResponse{Data: data})
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func writeErrResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	log.Error(err)

	jsonErr := json.NewEncoder(w).Encode(HTTPResponse{Error: err.Error()})
	if jsonErr != nil {
		log.Error(jsonErr)
	}
}