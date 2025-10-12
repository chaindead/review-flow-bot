package bot

import (
	"context"
	"database/sql"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	tele "gopkg.in/telebot.v4"

	"github.com/chaindead/review-flow-bot/internal/db/models"
	"github.com/chaindead/review-flow-bot/internal/lang"
)

func (b *Bot) myHandler(c tele.Context) error {
	u := models.User{ID: c.Sender().ID}
	err := b.db.DB().NewSelect().Model(&u).WherePK().
		Relation("TeamMembership").
		Relation("AuthoredMRs.Reviewers.Reviewer").
		Relation("ReviewingMRs.MergeRequest.Author").
		Scan(context.Background())
	if err != nil {
		return err
	}

	return c.Send(b.loc.Get("my.success", lang.Args{"User": u}), tele.ModeMarkdownV2)
}

func (b *Bot) reviewHandler(c tele.Context) error {
	ctx := context.Background()
	args := c.Args()
	if len(args) < 1 {
		return c.Send(b.loc.Get("review.usage", lang.NoArgs))
	}

	mrLink := args[0]
	forceUsernames := args[1:]
	forceUsernames = lo.Map(forceUsernames, func(s string, _ int) string {
		return strings.TrimPrefix(s, "@")
	})

	var failedUnames []string
	var forceUsers []models.User
	for _, uname := range forceUsernames {
		var fr models.User
		err := b.db.DB().NewSelect().Model(&fr).Where("telegram_username = ?", uname).Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				failedUnames = append(failedUnames, uname)
				continue
			}

			return err
		}

		forceUsers = append(forceUsers, fr)
	}
	if len(failedUnames) > 0 {
		return c.Send(b.loc.Get("review.bad_force", lang.Args{"Nicks": failedUnames}))
	}

	pID, mrID, err := b.git.ParseMRURL(mrLink)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse mr_link")
		return c.Send(b.loc.Get("review.bad_link", lang.NoArgs))
	}

	telegramID := c.Sender().ID
	log.Info().
		Int64("telegram_id", telegramID).
		Str("mr_link", mrLink).
		Strs("force_reviewers", forceUsernames).
		Msg("review request")

	mrInfo, err := b.git.GetMergeRequest(ctx, pID, mrID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get merge request")
		return err
	}

	mergeInfo, _, err := b.git.C().MergeRequests.GetMergeRequestApprovals(pID, mrID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get merge request approvals")
		return err
	}

	if mergeInfo.ApprovalsLeft == 0 {
		mergeInfo.ApprovalsLeft = 1
	}

	reviewers, err := b.db.SelectReviewers(ctx, telegramID, forceUsers, mergeInfo.ApprovalsLeft)
	if err != nil {
		log.Error().Err(err).Msg("failed to select reviewers")
		return err
	}

	if len(reviewers) == 0 {
		return c.Send(b.loc.Get("review.bad_reviewers", lang.NoArgs))
	}

	log.Info().
		Interface("reviewers", reviewers).
		Interface("mrInfo", mrInfo).
		Interface("mergeInfo", mergeInfo).
		Msg("review data")

	dbMR, err := b.db.MergeReqCreate(ctx, mrID, pID, telegramID,
		mrInfo.Title,
		mrInfo.WebURL,
		reviewers)
	if err != nil {
		return err
	}

	arg := lang.Args{
		"MR": dbMR,
	}
	for _, reviewer := range reviewers {
		if _, err = b.tg.Send(
			tele.ChatID(reviewer.ID),
			b.loc.Get("review.notify", arg),
			tele.ModeMarkdownV2); err != nil {
			log.Error().Err(err).Msg("failed to send message")
		}
	}

	return c.Send(b.loc.Get("review.success", arg), tele.ModeMarkdownV2)
}
