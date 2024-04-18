package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/basedalex/yadro-xkcd/pkg/xkcd"

	"github.com/basedalex/yadro-xkcd/pkg/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load(parseConfigPath())
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

	// go func() {
	// 	<-ctx.Done()
	// }()

}

// парсим флаги
func parseConfigPath() string {
	var configPath string
	flag.StringVar(&configPath, "c", "config.yaml", "path to config relative to executable")
	flag.Parse()
	return configPath
}
