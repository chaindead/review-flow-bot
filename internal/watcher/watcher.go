package watcher

import (
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/do/v2"

	"git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/config"
	"git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/db"
	"git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/gitlab"
	"git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/lang"
)

type Sender interface {
	Send(userID int64, locID string, args lang.Args) error
}
type Watcher struct {
	cfg config.Watcher `do:"cfg.watch"`
	tg  Sender         `do:""`
	db  *db.DB         `do:""`
	git *gitlab.Gitlab `do:""`

	stopChan chan struct{}
}

func New(i do.Injector) (*Watcher, error) {
	w, err := do.InvokeStruct[*Watcher](i)
	if err != nil {
		return nil, errors.Wrap(err, "initializing watcher")
	}

	w.stopChan = make(chan struct{})

	return w, nil
}

func (w *Watcher) Start() {
	log.Info().Msg("starting watcher")
	go w.Watch()
}

func (w *Watcher) Shutdown() error {
	log.Info().Msg("shutting down watcher")
	close(w.stopChan)

	time.Sleep(1 * time.Second)

	return nil
}
