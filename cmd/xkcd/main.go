package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/basedalex/yadro-xkcd/pkg/indexer"
	"github.com/basedalex/yadro-xkcd/pkg/xkcd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	configPath, findIndex, useIndex := parseArgs()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalln("error loading config:", err)
	}
	getResults(cfg, findIndex, useIndex)
	xkcd.SetWorker(cfg, ctx)

	err = indexer.Reverse(cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func parseArgs() (string, string, bool) {
	var configPath string
	var findIndex string
	var useIndex bool

	flag.StringVar(&configPath, "c", "config.yaml", "path to config relative to executable")

	flag.StringVar(&findIndex, "s", "hello world!", "type string to find indexes for")

	flag.BoolVar(&useIndex, "i", false, "set true to use indexes to search")
	flag.Parse()
	return configPath, findIndex, useIndex
}

func getResults(cfg *config.Config, s string, useIndex bool) {
	log.Println("waiting for results... for ", s)
	for i := 0; i < 5; i++ {
		if useIndex {
			sm, err := indexer.LinearSearch(cfg, s)
			if err != nil {
				log.Println(err)
			}
			rm, err := indexer.InvertSearch(cfg, s)
			if err != nil {
				log.Println(err)
			}
			if len(rm) != 0 && len(sm) != 0 {
				log.Println(rm)
				log.Println(sm)
				return
			} else {
				time.Sleep(time.Second)
			}
		} else {
			sm, err := indexer.LinearSearch(cfg, s)
			if err != nil {
				log.Println(err)
			}
			if len(sm) != 0 {
				log.Println(sm)
				return
			} else {
				time.Sleep(time.Second)
			}
		}
	}
}
