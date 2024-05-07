package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/basedalex/yadro-xkcd/pkg/indexer"
	"github.com/basedalex/yadro-xkcd/pkg/xkcd"
	log "github.com/sirupsen/logrus"
)

type HTTPResponse struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type Handler struct {
	cfg *config.Config
}

func NewServer(ctx context.Context, cfg *config.Config) error {
	srv := &http.Server{
		Addr: ":" + cfg.SrvPort,
		Handler: newRouter(cfg),
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

func newRouter(cfg *config.Config) *http.ServeMux {

	handler := Handler{
		cfg: cfg,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", ping)
	mux.HandleFunc("/update", handler.updatePics)
	mux.HandleFunc("/pics", handler.getPics)

	return mux
}

func ping(w http.ResponseWriter, r *http.Request) {
	writeOkResponse(w, http.StatusOK, "Hello World")
}

func (h *Handler) updatePics(w http.ResponseWriter, r *http.Request) {
	xkcd.SetWorker(r.Context(), h.cfg)

	err := indexer.Reverse(h.cfg)
	if err != nil {
		writeErrResponse(w, http.StatusBadRequest, err)
	}
	
	writeOkResponse(w, http.StatusOK, "Updating comics!")
}

func (h *Handler) getPics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	search := query.Get("search")

	if search == "" {
		writeErrResponse(w, http.StatusBadRequest, fmt.Errorf("no comics to search"))
		return
	}

	sm, err := indexer.InvertSearch(h.cfg, search)
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