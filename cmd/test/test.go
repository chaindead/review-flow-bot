package main

import (
	"github.com/rs/zerolog/log"

	_ "git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/logger"
)

func main() {
	log.Info().Send()
}
