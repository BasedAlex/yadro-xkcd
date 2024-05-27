package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/basedalex/yadro-xkcd/pkg/words"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type Postgres struct {
	db *pgxpool.Pool
}

func NewPostgres(ctx context.Context, dbConnect string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(dbConnect)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	err = db.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("error pinging to database: %w", err)
	}

	return &Postgres{db: db}, nil
}


type Page struct {
	Index 	string `json:"index"`
	Img      string   `json:"img"`
	Keywords []string `json:"keywords"`
}


func (db *Postgres) SaveComics(ctx context.Context, cfg *config.Config, comics Page) error {
	// check if comic already exists by its index
	query := `
	SELECT id FROM comics WHERE index = $1;`

	row := db.db.QueryRow(ctx, query, comics.Index)

	var comicIndex string
	err := row.Scan(&comicIndex)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("database: %w", err)
	}
	if comicIndex != "" {
		return nil
	}

	stmt := `
	INSERT INTO comics (index, image, keywords)
	VALUES ($1, $2, $3);`

	_, err = db.db.Exec(ctx, stmt, comics.Index, comics.Img, comics.Keywords)
	if err != nil {
		return fmt.Errorf("database: %w", err) 
	}

	return nil
}

func (db *Postgres) Reverse(ctx context.Context, cfg *config.Config) error {
	existingPages := make(map[string]Page)
	indexPages := make(map[string][]int)

	query := `SELECT index, image, keywords FROM comics`

	rows, err := db.db.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	
	for rows.Next() {
		var page Page
		rows.Scan(&page.Index, &page.Img, &page.Keywords)
		existingPages[page.Index] = page
	}

	for pagesIndex, pages := range existingPages {
		intIndex, err := strconv.Atoi(pagesIndex)
		if err != nil {
			fmt.Println("couldn't create index", err)
			continue
		}
		for _, keyword := range pages.Keywords {
			indexPages[keyword] = append(indexPages[keyword], intIndex)
		}
	}


	stmt := `
	INSERT INTO indexes (stem, comics)
	VALUES ($1, $2);`

	for key, value := range indexPages {
		_, err = db.db.Exec(ctx, stmt, key, value)
		if err != nil {
			logrus.Info(err)
			return fmt.Errorf("database: %w", err)
		}
	}

	return nil
}

type User struct {
	Login string
	Role string
}

func (db *Postgres) GetUserByLogin(ctx context.Context, login string) (User, error) {
	stmt := `SELECT login, role FROM users WHERE login = $1;`

	row := db.db.QueryRow(ctx, stmt, login)
	var user User
	err := row.Scan(&user.Login, &user.Role)
	if err != nil {
		logrus.Info(err)
		return User{}, fmt.Errorf("database: %w", err)
	}
	
	return user, nil
}

func (db *Postgres) GetUserPasswordByLogin(ctx context.Context, login string) (string, error) {
	stmt := `SELECT password FROM users WHERE login = $1;`

	row := db.db.QueryRow(ctx, stmt, login)

	var password string 

	err := row.Scan(&password)
	if err != nil {
		logrus.Info(err)
		return "", fmt.Errorf("database: %w", err)
	}
	
	return password, nil
}

func (db *Postgres) InvertSearch(ctx context.Context, cfg *config.Config, s string) (map[string][]int, error) {
	indexedPages := make(map[string][]int)

	stems, err := words.Steminator(s)
	if err != nil {
		fmt.Println("error stemming: ", err)
		return nil, err
	}

	query := `
	SELECT comics FROM indexes
	WHERE stem = $1;`

	for _, v := range stems {
		var comic []int
		row := db.db.QueryRow(ctx, query, v)
		err = row.Scan(&comic)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			logrus.Info(err)
		} 
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		indexedPages[v] = comic
	}
	
	return indexedPages, nil
}
