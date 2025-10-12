package db

import (
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	"github.com/chaindead/review-flow-bot/internal/db/models"
)

func TestDB_MergeReqCreate(t *testing.T) {
	db := getDB(t)

	res, err := db.MergeReqCreate(ctx,
		1, 2, 5000,
		"Example MR",
		"https:/gitlab.com/root/test-projec/-/merge_requests/1",
		[]models.User{
			{ID: 1000},
			{ID: 2000},
		},
	)
	require.NoError(t, err)

	require.Len(t, res.Reviewers, 2)
	log.Info().Interface("actual", res).Msg("actual")
}
