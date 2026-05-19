package bot

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	tele "gopkg.in/telebot.v4"

	"git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/lang"
)

func (b *Bot) listHandler(c tele.Context) error {
	ctx := context.Background()
	teams, err := b.db.GetAllTeams(ctx)
	if err != nil {
		return fmt.Errorf("get all teams: %w", err)
	}

	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Name < teams[j].Name
	})
	for _, team := range teams {
		sort.Slice(team.Members, func(i, j int) bool {
			if team.Members[i].Role == team.Members[j].Role {
				return team.Members[i].User.TelegramUsername < team.Members[j].User.TelegramUsername
			}

			return team.Members[i].Role == "reviewer"
		})
	}

	unassigned, err := b.db.GetUnassignedUsers(ctx)
	if err != nil {
		return fmt.Errorf("get unassigned: %w", err)
	}

	args := lang.Args{
		"Teams":      teams,
		"Unassigned": unassigned,
	}

	return c.Send(b.loc.Get("admin.list.teams", args), tele.ModeMarkdownV2)
}

func (b *Bot) assignHandler(c tele.Context) error {
	ctx := context.Background()
	args := c.Args()
	if len(args) < 2 {
		return c.Send(b.loc.Get("admin.assign.usage", lang.NoArgs))
	}

	username := strings.TrimPrefix(args[0], "@")
	teamName := args[1]
	role := "member"
	if len(args) >= 3 {
		if args[2] != "reviewer" {
			return c.Send(b.loc.Get("admin.assign.bad_role", lang.NoArgs))
		}

		role = "reviewer"
	}

	log.Info().
		Str("username", username).
		Str("team", teamName).
		Str("role", role).
		Msg("assign user to team")

	user, err := b.db.GetUserByTelegramUsername(ctx, username)
	if err != nil || user == nil {
		return c.Send(b.loc.Get("admin.assign.not_logged", lang.Args{"TelegramUsername": username}))
	}

	err = b.db.AssignUserToTeam(ctx, user.ID, teamName, role)
	if err != nil {
		return err
	}

	return c.Send(b.loc.Get("admin.assign.success", lang.Args{
		"TelegramUsername": username,
		"Team":             teamName,
		"Role":             role,
	}), tele.ModeMarkdownV2)
}
