package bot

import (
	"github.com/pkg/errors"
	tele "gopkg.in/telebot.v4"

	"git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/lang"
)

func (b *Bot) Send(userID int64, locID string, args lang.Args) error {
	_, err := b.tg.Send(tele.ChatID(userID), b.loc.Get(locID, args), tele.ModeMarkdownV2)

	return errors.Wrap(err, "sending message")
}
