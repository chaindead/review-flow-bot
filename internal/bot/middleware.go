package bot

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"
	tele "gopkg.in/telebot.v4"
)

// isAdmin checks if user is admin
func (b *Bot) isAdmin(userID int64) bool {
	for _, adminID := range b.cfgTg.AdminIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

// requireAuth middleware - requires user to be authenticated
func (b *Bot) requireAuth(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		ctx := context.Background()
		userID := c.Sender().ID

		user, err := b.db.GetUserByTelegramID(ctx, userID)
		if err != nil || user == nil {
			log.Debug().Int64("user_id", userID).Msg("unauthorized access attempt")
			return c.Send("❌ Сначала авторизуйтесь: /login <gitlab_token>")
		}

		return next(c)
	}
}

// requireAdmin middleware - requires user to be admin
func (b *Bot) requireAdmin(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		userID := c.Sender().ID

		if !b.isAdmin(userID) {
			log.Warn().Int64("user_id", userID).Msg("unauthorized admin access attempt")
			return c.Send("❌ Эта команда доступна только администраторам")
		}

		return next(c)
	}
}

// Helper functions

func (b *Bot) sendError(c tele.Context, err error) error {
	log.Error().Err(err).Msg("handler error")
	return c.Send(fmt.Sprintf("❌ Ошибка: %v", err))
}

func (b *Bot) sendSuccess(c tele.Context, msg string) error {
	return c.Send(fmt.Sprintf("✅ %s", msg))
}

func (b *Bot) onError(err error, c tele.Context) {
	log.Error().Err(err).Msg("handler error")

	s := func(s string) string {
		runes := []rune(s)
		rand.Seed(time.Now().UnixNano())
		for i := range runes {
			j := rand.Intn(len(runes))
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	}

	const lol_seed = "⸮ ﹖ ︖ ⁇ ¿ ‽ "
	lol := lol_seed + s(lol_seed) + s(lol_seed)

	msg := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n",
		lol, err.Error(), s(lol), s(lol), s(lol), s(lol), s(lol), s(lol), s(lol))

	err = c.Send(msg)
	if err != nil {
		log.Error().Err(err).Msg("handler send")
	}
}
