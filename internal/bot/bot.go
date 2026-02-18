package bot

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/samber/do/v2"
	tele "gopkg.in/telebot.v4"

	"github.com/chaindead/review-flow-bot/internal/config"
	"github.com/chaindead/review-flow-bot/internal/db"
	"github.com/chaindead/review-flow-bot/internal/gitlab"
)

type Localizer interface {
	Get(id string, args map[string]any) string
}

type Bot struct {
	cfgTg config.TG      `do:"cfg.tg"`
	db    *db.DB         `do:""`
	git   *gitlab.Gitlab `do:""`
	loc   Localizer      `do:""`

	tg *tele.Bot
}

func New(i do.Injector) (*Bot, error) {
	b, err := do.InvokeStruct[Bot](i)
	if err != nil {
		return nil, fmt.Errorf("load bot struct: %w", err)
	}

	pref := tele.Settings{
		Token:   b.cfgTg.Token,
		Poller:  &tele.LongPoller{Timeout: 10 * time.Second},
		OnError: b.onError,
	}

	b.tg, err = tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("create tg bot: %w", err)
	}

	b.setupRoutes()

	return &b, nil
}

func (b *Bot) Start() {
	log.Info().Str("username", b.tg.Me.Username).Msg("starting bot")

	go func() {
		b.tg.Start()
	}()
}

func (b *Bot) Shutdown() error {
	log.Info().Msg("stoping bot")

	b.tg.Stop()

	return nil
}

func (b *Bot) setupRoutes() {
	tg := b.tg

	// Public commands
	tg.Handle("/start", b.startHandler)
	tg.Handle("/help", b.helpHandler)
	tg.Handle("/id", b.idHandler)
	tg.Handle("/login", b.loginHandler)

	// Authenticated commands group
	authGroup := tg.Group()
	authGroup.Use(b.requireAuth)
	authGroup.Handle("/my", b.myHandler)
	authGroup.Handle("/review", b.reviewHandler)

	// Admin commands group
	adminGroup := tg.Group()
	adminGroup.Use(b.requireAdmin)
	adminGroup.Handle("/list", b.listHandler)
	adminGroup.Handle("/assign", b.assignHandler)
}
