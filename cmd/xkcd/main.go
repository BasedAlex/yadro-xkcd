package main

import (
	"flag"
	"log"

	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/basedalex/yadro-xkcd/pkg/indexer"
)

func main() {
	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	// defer cancel()

	cfg, err := config.Load(parseArgs())
	if err != nil {
		log.Fatalln("error loading config:", err)
	}
	sm, err := indexer.LinearSearch(cfg, "I'm following your questions")

	if err != nil {
		log.Fatal(err)
	}
	log.Println(sm)
	// now := time.Now()
	// xkcd.SetWorker(cfg, ctx)
	// fmt.Println(time.Since(now))
}

func parseArgs() string {
	var configPath string
	flag.StringVar(&configPath, "c", "config.yaml", "path to config relative to executable")
	flag.Parse()
	return configPath
}
