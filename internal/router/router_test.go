package router

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/basedalex/yadro-xkcd/internal/db"
	mock_router "github.com/basedalex/yadro-xkcd/internal/router/mocks"
	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/benbjohnson/clock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/ratelimit"
)

func TestHandler_NewServer(t *testing.T) {
	t.Run("server success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Millisecond)
		defer cancel()
	
		c := gomock.NewController(t)
		service := mock_router.NewMockxkcdService(c)
	
		cfg := &config.Config{
			JWTSecret:        "secret",
			TokenMaxTime:     24,
			RateLimit:        1,
			ConcurrencyLimit: 1,
		}
	
		err := NewServer(ctx, cfg, service)
		require.NoError(t, err)
	}) 
	t.Run("server incorrect port", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100 * time.Millisecond)
		defer cancel()
	
		c := gomock.NewController(t)
		service := mock_router.NewMockxkcdService(c)
	
		cfg := &config.Config{
			JWTSecret:        "secret",
			TokenMaxTime:     24,
			RateLimit:        1,
			ConcurrencyLimit: 1,
			SrvPort: "test",
		}
	
		err := NewServer(ctx, cfg, service)
		require.EqualError(t, err, "error with the server: listen tcp: lookup tcp/test: unknown port")
		}) 
}

func TestHandler_doTask(t *testing.T) {
	t.Run("invalid url", func(t *testing.T) {
		buffer := strings.Builder{}
		cfg := &config.Config{
			Path: " http://example.com/file[/].html",
		}
		done := make(chan struct{})
		doTask(context.Background(), nil, nil, cfg, 1, &buffer, done)
		expectedError := "couldn't make request: parse \" http://example.com/file[/].html1/info.0.json\": first path segment in URL cannot contain colon\n"
		assert.Equal(t, expectedError, buffer.String())
		close(done)
	})

	t.Run("request error", func(t *testing.T) {
		buffer := strings.Builder{}
		cfg := &config.Config{}
		done := make(chan struct{})
		doTask(context.Background(), nil, &http.Client{}, cfg, 1, &buffer, done)
		expectedError := "problem getting info from url: 1/info.0.json Get \"1/info.0.json\": unsupported protocol scheme \"\"\n"
		assert.Equal(t, expectedError, buffer.String())
		close(done)
	})

	t.Run("invalid status", func(t *testing.T) {
		buffer := strings.Builder{}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log("Request received, returning bad request status")
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer srv.Close()
		cfg := &config.Config{
			Path: srv.URL + "/",
		}
		done := make(chan struct{})
		
		t.Log("Starting doTask")
		go func() {
			doTask(context.Background(), nil, &http.Client{}, cfg, 1, &buffer, done)
			close(done)
		}()
		
		select {
		case <-done:
			expectedError := fmt.Sprintf("couldn't get info from url: %s/1/info.0.json\n", srv.URL)
			assert.Equal(t, expectedError, buffer.String())
		case <-time.After(10 * time.Second):
			t.Fatal("Test timed out")
		}
	})

	t.Run("invalid character", func(t *testing.T) {
		buffer := strings.Builder{}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("no content"))
		}))
		defer srv.Close()
		cfg := &config.Config{
			Path: srv.URL + "/",
		}
		done := make(chan struct{})
		doTask(context.Background(), nil, &http.Client{}, cfg, 1, &buffer, done)
		expectedError := "invalid character 'o' in literal null (expecting 'u')\n"
		assert.Equal(t, expectedError, buffer.String())
		close(done)
	})

	t.Run("can't parse json", func(t *testing.T) {
		buffer := strings.Builder{}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()
		cfg := &config.Config{
			Path: srv.URL + "/",
		}
		done := make(chan struct{})
		doTask(context.Background(), nil, &http.Client{}, cfg, 1, &buffer, done)
		expectedError := "unexpected end of JSON input\n"
		assert.Equal(t, expectedError, buffer.String())
		close(done)
	})
}

func TestHandler_login(t *testing.T) {
	type userType struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	type mockBehavior func(s *mock_router.MockxkcdService, user userType)

	testTable := []struct {
		name                string
		inputBody           string
		inputUser           userType
		mockBehavior        mockBehavior
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"name":"Test", "login":"admin", "password":"admin"}`,
			inputUser: userType{
				Login:    "admin",
				Password: "admin",
			},
			mockBehavior: func(s *mock_router.MockxkcdService, user userType) {
				ctx := context.Background()
				s.EXPECT().GetUserPasswordByLogin(ctx, user.Login).Return("admin", nil)
			},
			expectedStatusCode: 200,
		},
		{
			name:      "Incorrect Password",
			inputBody: `{"name":"Test", "login":"admin", "password":"123"}`,
			inputUser: userType{
				Login:    "admin",
				Password: "123",
			},
			mockBehavior: func(s *mock_router.MockxkcdService, user userType) {
				ctx := context.Background()
				s.EXPECT().GetUserPasswordByLogin(ctx, user.Login).Return("admin", nil)
			},
			expectedStatusCode:  401,
			expectedRequestBody: `{"error":"invalid credentials"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			service := mock_router.NewMockxkcdService(c)
			testCase.mockBehavior(service, testCase.inputUser)

			cfg := &config.Config{
				JWTSecret:        "secret",
				TokenMaxTime:     24,
				RateLimit:        1,
				ConcurrencyLimit: 1,
			}

			handler := &Handler{
				limiter:     ratelimit.New(1),
				concurrency: make(chan struct{}, 1),
				cfg:         cfg,
				service:     service,
				userToken:   "test",
			}

			r := http.NewServeMux()

			r.HandleFunc("/login", handler.login)

			w := httptest.NewRecorder()

			req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			t.Log(w.Body)
		})
	}
}

func TestHandler_Guard(t *testing.T) {
	type mockBehavior func(s *mock_router.MockxkcdService, token string)

	testTable := []struct {
		name               string
		token              string
		mockBehavior       mockBehavior
		expectedStatusCode int
	}{
		{
			name: "Valid Token",
			token: func() string {
				claims := jwt.MapClaims{
					"login": "admin",
					"exp":   time.Now().Add(time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("secret"))
				return tokenString
			}(),
			mockBehavior: func(s *mock_router.MockxkcdService, login string) {
				user := db.User{Role: "admin"}
				s.EXPECT().GetUserByLogin(gomock.Any(), login).Return(user, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Invalid Token",
			token:              "invalidToken",
			mockBehavior:       func(s *mock_router.MockxkcdService, login string) {},
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			service := mock_router.NewMockxkcdService(c)
			testCase.mockBehavior(service, "admin")

			cfg := &config.Config{
				JWTSecret:        "secret",
				TokenMaxTime:     24,
				RateLimit:        1,
				ConcurrencyLimit: 1,
			}

			handler := &Handler{
				limiter:     ratelimit.New(1),
				concurrency: make(chan struct{}, 1),
				cfg:         cfg,
				service:     service,
				userToken:   "test",
			}

			r := http.NewServeMux()

			protectedHandler := handler.Guard()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			r.Handle("/protected", protectedHandler)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("token", testCase.token)

			r.ServeHTTP(w, req)
			assert.Equal(t, testCase.expectedStatusCode, w.Code)

			if w.Code != testCase.expectedStatusCode {
				t.Errorf("expected status %d, got %d", testCase.expectedStatusCode, w.Code)
			}
			t.Log(w.Body)
		})
	}
}

func TestHandler_IsAuth(t *testing.T) {
	testTable := []struct {
		name               string
		contextUser        any
		expectedStatusCode int
	}{
		{
			name:               "User OK",
			contextUser:        "testUser",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "User Not Present",
			contextUser:        nil,
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {

			type contextKey string
			const userKey contextKey = "user"

			ctx := context.Background()
			if testCase.contextUser != nil {
				ctx = context.WithValue(ctx, userKey, testCase.contextUser)
			}

			handler := isAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/protected", nil)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != testCase.expectedStatusCode {
				t.Errorf("expected status %d, got %d", testCase.expectedStatusCode, w.Code)
			}
		})
	}
}

func TestHandler_CheckRole(t *testing.T) {
	testTable := []struct {
		name               string
		contextUser        any
		requiredRole       string
		expectedStatusCode int
	}{
		{
			name:               "Role Matches",
			contextUser:        "admin",
			requiredRole:       "admin",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Role Does Not Match",
			contextUser:        "user",
			requiredRole:       "admin",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "No User in Context",
			contextUser:        nil,
			requiredRole:       "admin",
			expectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			type contextKey string
			const userKey contextKey = "user"

			if testCase.contextUser != nil {
				ctx = context.WithValue(ctx, userKey, testCase.contextUser)
			}

			handler := checkRole(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}), testCase.requiredRole)

			req := httptest.NewRequest("GET", "/protected", nil)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != testCase.expectedStatusCode {
				t.Errorf("expected status %d, got %d", testCase.expectedStatusCode, w.Code)
			}
		})
	}
}

func TestHandler_updatePics(t *testing.T) {
	type mockBehavior func(s *mock_router.MockxkcdService)

	testTable := []struct {
		name               string
		mockBehavior       mockBehavior
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "Success",
			mockBehavior: func(s *mock_router.MockxkcdService) {
				ctx := context.Background()
				s.EXPECT().Reverse(ctx, gomock.Any()).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"data":"Updated comics..."}`,
		},
		{
			name: "Internal Server Error",
			mockBehavior: func(s *mock_router.MockxkcdService) {
				ctx := context.Background()
				s.EXPECT().Reverse(ctx, gomock.Any()).Return(errors.New("some error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   `{"error":"some error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			mockService := mock_router.NewMockxkcdService(c)
			testCase.mockBehavior(mockService)

			cfg := &config.Config{
				RateLimit:        1,
				ConcurrencyLimit: 1,
			}

			handler := &Handler{
				limiter:     ratelimit.New(cfg.RateLimit),
				concurrency: make(chan struct{}, cfg.ConcurrencyLimit),
				cfg:         cfg,
				service:     mockService,
				userToken:   "test",
			}

			req := httptest.NewRequest("POST", "/update", nil)
			w := httptest.NewRecorder()

			handler.updatePics(w, req)

			resp := w.Result()
			if resp.StatusCode != testCase.expectedStatusCode {
				t.Errorf("expected status %d, got %d", testCase.expectedStatusCode, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}

			expectedResponse := testCase.expectedResponse + "\n"
			if string(body) != expectedResponse {
				t.Errorf("expected response %q, got %q", expectedResponse, string(body))
			}
		})
	}
}

func TestHandler_getPics(t *testing.T) {
	type mockBehavior func(s *mock_router.MockxkcdService, search string)

	testTable := []struct {
		name               string
		searchQuery        string
		mockBehavior       mockBehavior
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:        "Success",
			searchQuery: "test",
			mockBehavior: func(s *mock_router.MockxkcdService, search string) {
				ctx := context.Background()
				results := map[string][]int{
					"comics": {1, 2, 3},
				}
				s.EXPECT().InvertSearch(ctx, gomock.Any(), search).Return(results, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"data":["/1","/2","/3"]}`,
		},
		{
			name:               "No Search Query",
			searchQuery:        "",
			mockBehavior:       func(s *mock_router.MockxkcdService, search string) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"error":"no comics to search"}`,
		},
		{
			name:        "Internal Server Error",
			searchQuery: "error",
			mockBehavior: func(s *mock_router.MockxkcdService, search string) {
				ctx := context.Background()
				s.EXPECT().InvertSearch(ctx, gomock.Any(), search).Return(nil, errors.New("some error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   `{"error":"some error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			mockService := mock_router.NewMockxkcdService(c)
			testCase.mockBehavior(mockService, testCase.searchQuery)

			cfg := &config.Config{
				RateLimit:        1,
				ConcurrencyLimit: 1,
				Path:             "/",
			}

			handler := &Handler{
				limiter:     ratelimit.New(cfg.RateLimit),
				concurrency: make(chan struct{}, cfg.ConcurrencyLimit),
				cfg:         cfg,
				service:     mockService,
				userToken:   "test",
			}

			req := httptest.NewRequest("GET", fmt.Sprintf("/pics?search=%s", testCase.searchQuery), nil)
			w := httptest.NewRecorder()

			handler.getPics(w, req)

			resp := w.Result()
			if resp.StatusCode != testCase.expectedStatusCode {
				t.Errorf("expected status %d, got %d", testCase.expectedStatusCode, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}

			expectedResponse := testCase.expectedResponse + "\n"
			if string(body) != expectedResponse {
				t.Errorf("expected response %q, got %q", expectedResponse, string(body))
			}
		})
	}
}

func TestHandler_NewScheduler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mock_router.NewMockxkcdService(ctrl)

	cfg := &config.Config{
		Parallel: 10,
		Path:     "http://example.com/",
	}

	h := &Handler{
		service:   mockService,
		cfg:       cfg,
		userToken: "test",
		clock:     clock.NewMock(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockService.EXPECT().SaveComics(gomock.Any(), cfg, gomock.Any()).AnyTimes()
	mockService.EXPECT().Reverse(gomock.Any(), cfg).AnyTimes()

	go h.NewScheduler(ctx)

	h.clock.(*clock.Mock).Add(24 * time.Hour)

	time.Sleep(100 * time.Millisecond)
}

func TestHandler_runUpdate(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	cfg := &config.Config{
		JWTSecret:        "secret",
		TokenMaxTime:     24,
		RateLimit:        1,
		ConcurrencyLimit: 1,
	}
	c := gomock.NewController(t)
	defer c.Finish()

	service := mock_router.NewMockxkcdService(c)
	handler := &Handler{
		limiter:     ratelimit.New(1),
		concurrency: make(chan struct{}, 1),
		cfg:         cfg,
		service:     service,
		userToken:   "test",
	}

	ctx := context.Background()
	handler.runUpdate(ctx)
	logOutput := buf.String()
	t.Log(logOutput)
}

func Test_WriteOkResponse(t *testing.T) {
	testCases := []struct {
		name               string
		statusCode         int
		data               any
		expectedStatusCode int
		expectedBody       string
		expectedLog        string
	}{
		{
			name:               "successful request with data",
			statusCode:         http.StatusOK,
			data:               map[string]string{"message": "success"},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"data":{"message":"success"}}` + "\n",
			expectedLog:        "successful request with statusCode 200 and data type map[string]string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			logrus.SetOutput(&logBuffer)
			defer logrus.SetOutput(os.Stderr)

			rr := httptest.NewRecorder()
			writeOkResponse(rr, tc.statusCode, tc.data)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
			require.Equal(t, tc.expectedBody, rr.Body.String())
			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, tc.expectedLog)
		})
	}
}

func Test_WriteErrResponse(t *testing.T) {
	testCases := []struct {
		name               string
		statusCode         int
		err                error
		expectedStatusCode int
		expectedBody       string
		expectedLog        string
	}{
		{
			name:               "error response",
			statusCode:         http.StatusInternalServerError,
			err:                fmt.Errorf("internal server error"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"error":"internal server error"}` + "\n",
			expectedLog:        "internal server error",
		},
		{
			name:               "not found error",
			statusCode:         http.StatusNotFound,
			err:                fmt.Errorf("not found"),
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       `{"error":"not found"}` + "\n",
			expectedLog:        "not found",
		},
		{
			name:               "bad request error",
			statusCode:         http.StatusBadRequest,
			err:                fmt.Errorf("bad request"),
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"error":"bad request"}` + "\n",
			expectedLog:        "bad request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var logBuffer bytes.Buffer
			logrus.SetOutput(&logBuffer)
			defer logrus.SetOutput(os.Stderr)

			rr := httptest.NewRecorder()

			writeErrResponse(rr, tc.statusCode, tc.err)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)

			require.Equal(t, tc.expectedBody, rr.Body.String())

			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, tc.expectedLog)
		})
	}
}
