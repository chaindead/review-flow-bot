package bot

import (
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
	s.NoError2(s.DB().NewInsert().Model(&teams).Exec(ctx))
	s.NoError2(s.DB().NewInsert().Model(&users).Exec(ctx))

	s.Run(s.bot.listHandler)
	s.Len(s.loc.args["Teams"], 2)
	s.Len(s.loc.args["Unassigned"], 3)
}

func (s *Suite) TestHandlerAssignUsage() {
	s.tc.EXPECT().Args().Return(nil).Once()
	s.Run(s.bot.assignHandler)

	s.Equal("admin.assign.usage", s.loc.id)
}

func (s *Suite) TestHandlerAssignBadRole() {
	s.tc.EXPECT().Args().Return([]string{
		"tg_user",
		"test_team",
		"bad_role",
	})
	s.Run(s.bot.assignHandler)

	s.Equal("admin.assign.bad_role", s.loc.id)
}

func (s *Suite) TestHandlerAssignBadUser() {
	s.tc.EXPECT().Args().Return([]string{
		"not_logged_in",
		"",
	})
	s.Run(s.bot.assignHandler)

	s.Equal("admin.assign.not_logged", s.loc.id)
}

func (s *Suite) TestHandlerAssignNewTeam() {
	u := s.createUser()
	newTeam := "not_exists_team"

	s.tc.EXPECT().Args().Return([]string{
		u.TelegramUsername,
		newTeam,
	})
	s.Run(s.bot.assignHandler)

	s.Equal("admin.assign.success", s.loc.id)
	s.DB().NewSelect().Model(&u).WherePK().Relation("TeamMembership").Scan(ctx)
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
	s.Run(s.bot.assignHandler)

	s.Equal("admin.assign.success", s.loc.id)
	s.DB().NewSelect().Model(&u).Relation("TeamMembership").WherePK().Scan(ctx)
	s.Equal(team, u.TeamMembership.TeamID)
}
