package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/basedalex/yadro-xkcd/internal/router"
	"github.com/basedalex/yadro-xkcd/internal/scheduler"
	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	configPath:= parseArgs()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalln("error loading config:", err)
	}

	logrus.Info("server started on port:", cfg.SrvPort)

	go scheduler.New(ctx, cfg)

	err = router.NewServer(ctx, cfg)
	if err != nil {
		log.Fatalln("error serving on port:", cfg.SrvPort)
	}

	// getResults(cfg, findIndex, useIndex)
	// xkcd.SetWorker(ctx, cfg)

	

	// err = indexer.Reverse(cfg)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func parseArgs() (string, ) {
	var configPath string
	flag.StringVar(&configPath, "c", "config.yaml", "path to config relative to executable")

	flag.Parse()
	return configPath
}

// func getResults(cfg *config.Config, s string, useIndex bool) {
// 	log.Println("waiting for results... for ", s)
// 	for i := 0; i < 5; i++ {
// 		if useIndex {
// 			sm, err := indexer.LinearSearch(cfg, s)
// 			if err != nil {
// 				log.Println(err)
// 			}
// 			rm, err := indexer.InvertSearch(cfg, s)
// 			if err != nil {
// 				log.Println(err)
// 			}
// 			if len(rm) != 0 && len(sm) != 0 {
// 				log.Println(rm)
// 				log.Println(sm)
// 				return
// 			} else {
// 				time.Sleep(time.Second)
// 			}
// 		} else {
// 			sm, err := indexer.LinearSearch(cfg, s)
// 			if err != nil {
// 				log.Println(err)
// 			}

// 			if len(sm) != 0 {
// 				log.Println(sm)
// 				return
// 			} else {
// 				time.Sleep(time.Second)
// 			}
// 		}
// 	}
// }
