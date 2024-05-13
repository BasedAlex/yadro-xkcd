package scheduler

import (
	"context"
	"time"

	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/basedalex/yadro-xkcd/pkg/xkcd"
	"github.com/sirupsen/logrus"
)

func New(ctx context.Context, cfg *config.Config) {
    ticker := time.NewTicker(24 * time.Hour)

    runUpdate(ctx, cfg)

    go func() {
        for {
            select {
            case <-ticker.C:
                runUpdate(ctx, cfg)
            }
        }
    }()

    select {}

}

func runUpdate(ctx context.Context, cfg *config.Config) {
	xkcd.SetWorker(ctx, cfg)

    logrus.Println("Last updated at:", time.Now())
}