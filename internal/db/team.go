package db

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/uptrace/bun"

	"github.com/chaindead/review-flow-bot/internal/db/models"
)

func (d *DB) GetAllTeams(ctx context.Context) ([]*models.Team, error) {
	var teams []*models.Team
	err := d.db.NewSelect().
		Model(&teams).
		Relation("Members").
		Relation("Members.User").
		Order("name ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (d *DB) AssignUserToTeam(ctx context.Context, userID int64, teamName, role string) error {
	// Create team if not exists
	team := &models.Team{
		Name:      teamName,
		CreatedAt: time.Now(),
	}
	_, err := d.db.NewInsert().
		Model(team).
		Ignore().
		Exec(ctx)
	if err != nil {
		return err
	}

	// Upsert team member
	teamMember := &models.TeamMember{
		UserID: userID,
		TeamID: teamName,
		Role:   role,
	}

	_, err = d.db.NewInsert().
		Model(teamMember).
		On("CONFLICT (user_id) DO UPDATE").
		Set("team_id = EXCLUDED.team_id").
		Set("role = EXCLUDED.role").
		Exec(ctx)

	return err
}

func (d *DB) SelectReviewers(ctx context.Context, authorID int64, forced []models.User, count int) ([]models.User, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be greater than zero")
	}

	u := models.User{ID: authorID}
	err := d.db.NewSelect().Model(&u).WherePK().Relation("TeamMembership").Scan(ctx)
	if err != nil {
		return nil, err
	}

	needed := count - len(forced)
	if needed < 0 {
		return forced, nil
	}

	if u.TeamMembership == nil {
		return forced, nil
	}

	excludeIDs := append(lo.Map(forced, func(item models.User, _ int) int64 {
		return item.ID
	}), u.ID)

	var selected []models.TeamMember
	err = d.db.NewSelect().
		Model(&selected).
		Where("tm.team_id = ?", u.TeamMembership.TeamID).
		Where("tm.role = ?", models.RoleReviewer).
		Where("tm.user_id NOT IN (?)", bun.In(excludeIDs)).
		Relation("User").
		OrderExpr("random()").
		Limit(needed).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	reviewers := append(lo.Map(selected, func(item models.TeamMember, _ int) models.User {
		return *item.User
	}), forced...)

	return reviewers, nil
}
