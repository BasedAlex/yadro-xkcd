package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/basedalex/yadro-xkcd/internal/db"
	"github.com/basedalex/yadro-xkcd/internal/router"
	"github.com/basedalex/yadro-xkcd/pkg/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	database, cfg := prepare(ctx, os.Stdout)
	// go scheduler.New(ctx, cfg)

	err := router.NewServer(ctx, cfg, database)
	if err != nil {
		log.Fatalln("error serving on port:", cfg.SrvPort)
	}
}

func prepare(ctx context.Context, out io.Writer) (*db.Postgres, *config.Config) {
	configPath := parseArgs()
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprint(out, "error loading config:", err)
		log.Fatalln("error loading config:", err)
	}

	database, err := db.NewPostgres(ctx, cfg.DSN)
	if err != nil {
		log.Panic(err)
	}
	fmt.Fprint(out, "connected to database")
	fmt.Fprint(out, "server started on port:", cfg.SrvPort)
	return database, cfg
}

func parseArgs() string {
	var configPath string
	flag.StringVar(&configPath, "c", "config.yaml", "path to config relative to executable")

	flag.Parse()
	return configPath
}
