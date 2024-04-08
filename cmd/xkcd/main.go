package main

import (
	"flag"
	"fmt"
	"log"
	"yardro-xkcd/pkg/config"
	"yardro-xkcd/pkg/xkcd"
)


func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalln("error loading config:", err)
	}

	createFlags(cfg)

	// согласно заданию вызываем функцию при старте программы
	err = xkcd.WriteToDB(cfg)
	if err != nil {
		log.Fatalln(err)
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
		cfg.Limit = newLimit
	} else if newLimit < 0 {
		fmt.Println("please provide a positive value")
	}
}