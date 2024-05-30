package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/basedalex/yadro-xkcd/internal/db"
	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/basedalex/yadro-xkcd/pkg/words"
	"github.com/benbjohnson/clock"
	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"go.uber.org/ratelimit"
)

//go:generate mockgen -source=router.go -destination=mocks/mock.go

type HTTPResponse struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type xkcdService interface {
	SaveComics(ctx context.Context, cfg *config.Config, comics db.Page) error
	Reverse(ctx context.Context, cfg *config.Config) error
	InvertSearch(ctx context.Context, cfg *config.Config, s string) (map[string][]int, error)
	GetUserByLogin(ctx context.Context, login string) (db.User, error)
	GetUserPasswordByLogin(ctx context.Context, login string) (string, error)
}

type Handler struct {
	limiter     ratelimit.Limiter
	concurrency chan struct{}
	userToken   string
	service     xkcdService
	cfg         *config.Config
	clock       clock.Clock
}

func NewServer(ctx context.Context, cfg *config.Config, service xkcdService) error {
	srv := &http.Server{
		Addr:              ":" + cfg.SrvPort,
		Handler:           newRouter(cfg, service),
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
		limiter:     ratelimit.New(cfg.RateLimit),
		concurrency: make(chan struct{}, cfg.ConcurrencyLimit),
		cfg:         cfg,
		service:     service,
		userToken:   "test",
	}

	mux := http.NewServeMux()

	// handler.NewScheduler(context.Background())

	mux.Handle("/update", handler.Guard()(checkRole(
		isAuth(http.HandlerFunc(handler.updatePics)), "admin")))

	mux.HandleFunc("/login", handler.login)
	mux.Handle("/pics", handler.Guard()(isAuth(http.HandlerFunc(handler.getPics))))

	return mux
}

func checkRole(next http.Handler, role string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user")
		fmt.Println("CHECKING ROLE", user)
		if user != role {
			writeErrResponse(w, http.StatusUnauthorized, fmt.Errorf("invalid role"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user")
		fmt.Printf("user %v", user)
		if user == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) Guard() func(http.Handler) http.Handler {
	type MyCustomClaims struct {
		Login string `json:"login"`
		jwt.MapClaims
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("token")
			token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(t *jwt.Token) (interface{}, error) {
				return []byte(h.cfg.JWTSecret), nil
			})
			if err != nil {
				writeErrResponse(w, http.StatusBadRequest, err)
				return
			}
			log.Debug(token.Claims)
			if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
				log.Printf("%v %v", claims.Login, claims.MapClaims)
				user, err := h.service.GetUserByLogin(r.Context(), claims.Login)
				if err != nil {
					writeErrResponse(w, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
					return
				}
				fmt.Println(user)
				ctxWithValue := context.WithValue(r.Context(), "user", user.Role)
				next.ServeHTTP(w, r.WithContext(ctxWithValue))

			}

			w.WriteHeader(http.StatusForbidden)
		})
	}
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeErrResponse(w, http.StatusBadRequest, err)
		return
	}

	storedPassword, err := h.service.GetUserPasswordByLogin(r.Context(), creds.Login)
	if err != nil {
		writeErrResponse(w, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
		return
	}

	if storedPassword != creds.Password {
		writeErrResponse(w, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": creds.Login,
		"exp":   time.Now().Add(time.Hour * time.Duration(h.cfg.TokenMaxTime)).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		writeErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	writeOkResponse(w, http.StatusOK, map[string]string{"token": tokenString})
}

func (h *Handler) updatePics(w http.ResponseWriter, r *http.Request) {
	h.limiter.Take()
	h.concurrency <- struct{}{}
	defer func() {
		<-h.concurrency
	}()

	h.SetWorker(r.Context(), h.cfg)
	err := h.service.Reverse(r.Context(), h.cfg)
	if err != nil {
		writeErrResponse(w, http.StatusInternalServerError, err)
		return
	}
	writeOkResponse(w, http.StatusOK, "Updated comics...")
}

func (h *Handler) getPics(w http.ResponseWriter, r *http.Request) {
	h.limiter.Take()
	h.concurrency <- struct{}{}
	defer func() {
		<-h.concurrency
	}()

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
	if h.clock == nil {
		h.clock = clock.New()
	}

	ticker := h.clock.Ticker(24 * time.Hour) // Используем mock clock

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
	Num        int    `json:"num"`
	Alt        string `json:"alt"`
	Transcript string `json:"transcript"`
	Img        string `json:"img"`
}

func (h *Handler) SetWorker(ctx context.Context, cfg *config.Config) {
	results := make(chan db.Page, cfg.Parallel)
	intCh := make(chan int)
	errorCh := make(chan struct{})

	client := &http.Client{
		Timeout: clientTimeout * time.Second,
	}

	var wg sync.WaitGroup

	for i := 1; i <= cfg.Parallel; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			task(ctx, results, client, cfg, intCh, errorCh)
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
	}()

loop:
	for i := 1; ; i++ {
		if i == 404 {
			continue
		}
		select {
		case <-generatorDoneCh:
			//worker close intch
			close(intCh)
			break loop
		case intCh <- i:
		case <-ctx.Done():
			close(intCh)
			break loop
		case <-errorCh:
			close(intCh)
			break loop
		}

	}

	// <-resultDoneCh
	fmt.Println("finished fetching data...")
}

func doTask(ctx context.Context, results chan<- db.Page, client *http.Client, cfg *config.Config, w int, out io.Writer, errorCh chan<- struct{}) {
	url := fmt.Sprintf("%s%d/info.0.json", cfg.Path, w)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintln(out, "couldn't make request:", err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(out, "problem getting info from url:", url, err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		fmt.Fprintln(out, "couldn't get info from url:", url)
		errorCh <- struct{}{}
		return
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintln(out, "nothing found")
		return
	}

	var raw rawPage
	err = json.Unmarshal(content, &raw)
	if err != nil {
		fmt.Fprintln(out, err)
		return
	}

	keywords := raw.Alt + " " + raw.Transcript
	stemmedKeywords, err := words.Steminator(keywords)
	if err != nil {
		fmt.Fprintln(out, "error stemming: ", err)
		return
	}

	var page db.Page

	page.Keywords = stemmedKeywords
	page.Img = raw.Img
	page.Index = strconv.Itoa(raw.Num)
	results <- page
}

func task(ctx context.Context, results chan<- db.Page, client *http.Client, cfg *config.Config, intCh chan int, errorCh chan<- struct{}) {
	// for w := range intCh {
	// 	doTask(ctx, results, client, cfg, w, os.Stdout)
	// }

	for {
		select {
		case w, ok := <-intCh:
			if !ok {
				return
			}
			doTask(ctx, results, client, cfg, w, os.Stdout, errorCh)
		case <-ctx.Done():
			return
		}
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
