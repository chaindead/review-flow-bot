package bot

import (
	"cmp"
	"math/rand"

	"github.com/stretchr/testify/require"

	"github.com/chaindead/review-flow-bot/internal/db/models"
)

func rndString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a' + byte(rand.Intn(26))
	}

	return string(b)
}

func (s *Suite) createUser() models.User {
	s.T().Helper()

	u := models.User{
		ID:               rand.Int63(),
		GitlabUsername:   "gitlab_" + rndString(8),
		TelegramUsername: "telegram_" + rndString(8),
		GitlabID:         rand.Int63(),
	}

	_, err := s.bot.db.DB().NewInsert().Model(&u).Exec(ctx)
	s.Require().NoError(err)

	return u
}

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

func (s *Suite) TestHandlerAssignUsage() {
	s.tc.EXPECT().Args().Return(nil).Once()
	s.Require().NoError(s.bot.assignHandler(s.tc))

	s.Equal("admin.assign.usage", s.loc.id)
}

func (s *Suite) TestHandlerAssignBadRole() {
	s.tc.EXPECT().Args().Return([]string{
		"tg_user",
		"test_team",
		"bad_role",
	})
	s.Require().NoError(s.bot.assignHandler(s.tc))

	s.Equal("admin.assign.bad_role", s.loc.id)
}

func (s *Suite) TestHandlerAssignBadUser() {
	s.tc.EXPECT().Args().Return([]string{
		"not_logged_in",
		"",
	})
	s.Require().NoError(s.bot.assignHandler(s.tc))

	s.Equal("admin.assign.not_logged", s.loc.id)
}

func (s *Suite) TestHandlerAssignNewTeam() {
	u := s.createUser()
	newTeam := "not_exists_team"

	s.tc.EXPECT().Args().Return([]string{
		u.TelegramUsername,
		newTeam,
	})
	s.Require().NoError(s.bot.assignHandler(s.tc))

	s.Equal("admin.assign.success", s.loc.id)
	s.bot.db.DB().NewSelect().Model(&u).WherePK().Relation("TeamMembership").Scan(ctx)
	s.Equal(newTeam, u.TeamMembership.TeamID)
}

func (s *Suite) TestHandlerAssignExistTeam() {
	u := s.createUser()
	team := "exists_team"
	s.DB().NewInsert().Model(&models.Team{Name: team}).Exec(ctx)

	s.tc.EXPECT().Args().Return([]string{
		u.TelegramUsername,
		team,
	})
	s.Require().NoError(s.bot.assignHandler(s.tc))

	s.Equal("admin.assign.success", s.loc.id)
	s.DB().NewSelect().Model(&u).Relation("TeamMembership").WherePK().Scan(ctx)
	s.Equal(team, u.TeamMembership.TeamID)
}
