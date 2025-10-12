package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/chaindead/review-flow-bot/internal/db/models"
)

var ctx = context.Background()

func TestDB_SelectReviewers(t *testing.T) {
	db := getDB(t)

	var forced []models.User
	err := db.DB().NewSelect().Model(&forced).OrderExpr("random()").Limit(2).Scan(ctx)
	require.NoError(t, err)

	reviewers, err := db.SelectReviewers(ctx, 5000, forced, 10)
	require.NoError(t, err)

	fmt.Println(lo.Map(reviewers, func(item models.User, _ int) int64 {
		return item.ID
	}))
	require.True(t, len(reviewers) > 3)
}
