package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"yardro-xkcd/pkg/config"
	"yardro-xkcd/pkg/xkcd"
)


func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalln("error loading config")
	}

	createFlags(cfg)

	// согласно заданию вызываем функцию при старте программы
	xkcd.GetPages(cfg)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", cfg.Port),
		Handler: xkcd.Routes(cfg),
	}

	err = srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}

}

// парсим флаги
func createFlags(cfg *config.Config) {
	var showPages bool
	var newLimit int

	flag.BoolVar(&showPages, "o", false, "select true to print results to console")
	flag.IntVar(&newLimit, "n", 0, "type number to limit pages")
	flag.Parse()

	cfg.Print = showPages
	if newLimit > 0 {
		cfg.Limit = cfg.Start + newLimit
	} else if newLimit < 0 {
		fmt.Println("please provide a positive value")
	}
}