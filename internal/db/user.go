package db

import (
	"context"
	"database/sql"
	"errors"

	"git.dev.bi.zone/pam/onprem/dev/tools/review-flow-bot/internal/db/models"
)

func (d *DB) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	user := &models.User{ID: telegramID}
	err := d.db.NewSelect().
		Model(user).
		WherePK().
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *DB) GetUserByTelegramUsername(ctx context.Context, username string) (*models.User, error) {
	user := new(models.User)
	err := d.db.NewSelect().
		Model(user).
		Where("telegram_username = ?", username).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *DB) SaveUser(ctx context.Context, telegramID, gitlabID int64, telegramUsername, gitlabUsername string) error {
	user := &models.User{
		ID:               telegramID,
		TelegramUsername: telegramUsername,
		GitlabUsername:   gitlabUsername,
		GitlabID:         gitlabID,
	}

	_, err := d.db.NewInsert().
		Model(user).
		On("CONFLICT (id) DO UPDATE").
		Exec(ctx)

	return err
}

func (d *DB) GetUnassignedUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User

	err := d.db.NewSelect().
		Model(&users).
		Where("NOT EXISTS (SELECT 1 FROM team_members tm WHERE tm.user_id = u.id)").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return users, nil
}
