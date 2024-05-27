package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/basedalex/yadro-xkcd/internal/db"
	"github.com/basedalex/yadro-xkcd/internal/router"
	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	configPath := parseArgs()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalln("error loading config:", err)
	}

	database, err := db.NewPostgres(ctx, cfg.DSN)
	if err != nil {
		log.Panic(err)
	}
	logrus.Info("connected to database")
	logrus.Info("server started on port:", cfg.SrvPort)

	// go scheduler.New(ctx, cfg)

	err = router.NewServer(ctx, cfg, database)
	if err != nil {
		log.Fatalln("error serving on port:", cfg.SrvPort)
	}
}

func parseArgs() string {
	var configPath string
	flag.StringVar(&configPath, "c", "config.yaml", "path to config relative to executable")

	flag.Parse()
	return configPath
}
