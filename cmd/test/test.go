package main

import (
	"github.com/rs/zerolog/log"

	_ "github.com/chaindead/review-flow-bot/internal/logger"
)

func main() {
	log.Info().Send()
}
