package config

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"
	"github.com/samber/do/v2"
)

type Config struct {
	TG      TG      `envPrefix:"TG_"`
	DB      DB      `envPrefix:"DB_"`
	Gitlab  Gitlab  `envPrefix:"GITLAB_"`
	Watcher Watcher `envPrefix:"WATCH_"`
}

type TG struct {
	Token    string  `env:"TOKEN,required"`
	AdminIDs []int64 `env:"ADMIN_IDS" envSeparator:","`
}

type DB struct {
	File string `env:"DB_FILE" envDefault:"review-flow-bot.db"`
	Seed bool   `env:"SEED" envDefault:"false"`
}

type Gitlab struct {
	Token string `env:"TOKEN,required"`
	URL   string `env:"URL" envDefault:"https://gitlab.com"`
}

type Watcher struct {
	PollInterval time.Duration `env:"POLL_INTERVAL" envDefault:"5s"`
}

func Init(i do.Injector) error {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	// setup providers
	v := reflect.ValueOf(cfg)
	t := reflect.TypeOf(cfg)
	for idx := 0; idx < v.NumField(); idx++ {
		tag := t.Field(idx).Tag.Get("envPrefix")
		value := v.Field(idx).Interface()
		provideName := envPrefixName(tag)

		do.ProvideNamedValue(i, provideName, value)
		log.Debug().Str("name", provideName).Interface("value", value).Msg("register config")
	}

	return nil
}

func envPrefixName(tag string) string {
	name := "cfg." + strings.ToLower(tag)

	return strings.TrimSuffix(name, "_")
}
