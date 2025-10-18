package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/samber/do/v2"
	"github.com/spf13/pflag"

	"github.com/chaindead/review-flow-bot/internal/bot"
	"github.com/chaindead/review-flow-bot/internal/config"
	"github.com/chaindead/review-flow-bot/internal/db"
	"github.com/chaindead/review-flow-bot/internal/gitlab"
	"github.com/chaindead/review-flow-bot/internal/http"
	"github.com/chaindead/review-flow-bot/internal/lang"
	_ "github.com/chaindead/review-flow-bot/internal/logger"
	"github.com/chaindead/review-flow-bot/internal/watcher"
)

var (
	buildTag = "dev"

	showVersion = pflag.BoolP("version", "v", false, "show version")
	showEnvs    = pflag.BoolP("envs", "e", false, "show setting options")
)

func main() {
	pflag.Parse()

	if *showVersion {
		fmt.Println("review-flow-bot", buildTag)
		os.Exit(0)
	}

	if *showEnvs {
		config.Help()
		os.Exit(0)
	}

	injector := do.New()
	if err := config.Init(injector); err != nil {
		log.Fatal().Err(err).Msg("init config")
	}
	do.Provide(injector, db.New)
	do.Provide(injector, bot.New)
	do.Provide(injector, gitlab.New)
	do.Provide(injector, lang.New)
	do.Provide(injector, watcher.New)
	do.Provide(injector, http.New)

	//w, err := do.Invoke[*watcher.Watcher](injector)
	//if err != nil {
	//	log.Fatal().Err(err).Msg("failed to initialize watcher")
	//}

	tg, err := do.Invoke[*bot.Bot](injector)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize bot")
	}

	srv, err := do.Invoke[*http.Server](injector)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize http")
	}

	go tg.Start()
	go srv.Start()

	_, report := injector.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
	if !report.Succeed {
		for service, err := range report.Errors {
			log.Err(err).Interface("service", service).Msg("failed to shutdown")
		}
	}
}
