package bot

import (
	"cmp"

	"github.com/stretchr/testify/require"

	"github.com/chaindead/review-flow-bot/internal/db/models"
)

func (s *Suite) TestHandlerList() {
	teams := []models.Team{
		{Name: "team1"},
		{Name: "team2"},
	}

	users := []models.User{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}
	_, err := s.bot.db.DB().NewInsert().Model(&teams).Exec(ctx)
	_, err2 := s.bot.db.DB().NewInsert().Model(&users).Exec(ctx)
	require.NoError(s.T(), cmp.Or(err, err2))

	s.Require().NoError(s.bot.listHandler(s.tc))
	s.Len(s.loc.args["Teams"], 2)
	s.Len(s.loc.args["Unassigned"], 3)
}
