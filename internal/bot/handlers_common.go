package bot

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	tele "gopkg.in/telebot.v4"

	"github.com/chaindead/review-flow-bot/internal/lang"
)

func (b *Bot) startHandler(c tele.Context) error {
	return c.Send(b.loc.Get("start.welcome", lang.NoArgs), tele.ModeMarkdownV2)
}

func (b *Bot) helpHandler(c tele.Context) error {
	return c.Send(b.loc.Get("help.commands", lang.Args{
		"IsAdmin": b.isAdmin(c.Sender().ID),
	}), tele.ModeMarkdownV2)
}

func (b *Bot) idHandler(c tele.Context) error {
	userID := c.Sender().ID
	username := c.Sender().Username

	msg := fmt.Sprintf("Your Telegram ID: `%d`\nUsername: @%s", userID, username)
	return c.Send(msg, tele.ModeMarkdown)
}

func (b *Bot) loginHandler(c tele.Context) error {
	args := c.Args()
	if len(args) < 1 {
		return c.Send(b.loc.Get("login.usage", lang.NoArgs))
	}

	gitlabToken := args[0]
	telegramID := c.Sender().ID
	telegramUsername := c.Sender().Username

	log.Info().
		Int64("telegram_id", telegramID).
		Str("telegram_username", telegramUsername).
		Msg("login attempt")

	ctx := context.Background()

	gitlabUser, err := b.git.GetCurrentUser(ctx, gitlabToken)
	if err != nil {
		log.Error().Err(err).Msg("failed to get gitlab user")
		return c.Send(b.loc.Get("login.fail_gitlab", lang.NoArgs))
	}

	err = b.db.SaveUser(ctx, telegramID, int64(gitlabUser.ID), telegramUsername, gitlabUser.Username)
	if err != nil {
		return err
	}

	return c.Send(b.loc.Get("login.success", lang.Args{
		"TelegramUsername": telegramUsername,
		"GitlabUsername":   gitlabUser.Username,
	}), tele.ModeMarkdownV2)
}
