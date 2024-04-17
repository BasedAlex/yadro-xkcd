package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
	"yardro-xkcd/pkg/config"
	"yardro-xkcd/pkg/xkcd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load(parseArgs())
	if err != nil {
		log.Fatalln("error loading config:", err)
	}

	// sm, err := indexer.Stem("I'm following your questions")

	// if err != nil {
	// 	log.Fatal(err)
	// } 
	// log.Println(sm)

	now := time.Now()
	xkcd.SetWorker(cfg, ctx)
	fmt.Println(time.Since(now))
}

func parseArgs() string {
	var configPath string
	flag.StringVar(&configPath, "c", "config.yaml", "path to config relative to executable")
	flag.Parse()
	return configPath
}