package indexer

import (
	"log"
	"testing"

	"github.com/basedalex/yadro-xkcd/pkg/config"
)

func BenchmarkLinearSearch(b *testing.B) {
	cfg, err := config.Load("../../config.yaml")
	if err != nil {
		log.Fatalln("error loading config:", err)
	}
	for i := 0; i < b.N; i++ {
		_, err = LinearSearch(cfg, "I'm following your questions")
		if err != nil {
			log.Print(err)
		}
	}
}

func BenchmarkInvertSearch(b *testing.B) {
	cfg, err := config.Load("../../config.yaml")
	if err != nil {
		log.Fatalln("error loading config:", err)
	}
	for i := 0; i < b.N; i++ {
		_, err = InvertSearch(cfg, "I'm following your questions")
		if err != nil {
			log.Print(err)
		}
	}
}
