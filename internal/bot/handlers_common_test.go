package bot

import (
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	gm "github.com/rumenvasilev/go-gitlab-mock/mock"
	"github.com/stretchr/testify/mock"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/chaindead/review-flow-bot/internal/db/models"
)

func (s *Suite) TestHandlerID() {
	id := s.tc.Sender().ID

	s.tc.EXPECT().
		Send(
			mock.MatchedBy(func(s string) bool {
				return strings.Contains(s, strconv.FormatInt(id, 10))
			}),
			mock.Anything).
		Return(nil)

	s.Require().NoError(s.bot.idHandler(s.tc))
	s.loc.called = true // hack
}

func (s *Suite) TestHandlerLogin() {
	token := "glpat-12345"
	u := &gitlab.User{
		ID:       rand.Int(),
		Username: "gitlab_u",
	}

	s.gitlab(
		gm.WithRequestMatchHandler(gm.EndpointPattern{Pattern: "/api/v4/user", Method: "GET"},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				s.Assert().Equal(token, r.Header.Get("Private-Token"))
				_, _ = w.Write(gm.MustMarshal(u))
			}),
		),
	)
	s.tc.EXPECT().
		Args().
		Return([]string{token})

	s.Require().NoError(s.bot.loginHandler(s.tc))
	s.Equal(u.Username, s.loc.args["GitlabUsername"])
	s.Equal(s.tc.Sender().Username, s.loc.args["TelegramUsername"])

	dbU := models.User{ID: s.tc.Sender().ID}
	s.Require().NoError(s.bot.db.DB().NewSelect().Model(&dbU).WherePK().Scan(ctx))

	s.Equal(s.tc.Sender().ID, dbU.ID)
	s.Equal(s.tc.Sender().Username, dbU.TelegramUsername)
	s.Equal(int64(u.ID), dbU.GitlabID)
	s.Equal(u.Username, dbU.GitlabUsername)
}

func (s *Suite) TestHandlerLogin_Failed() {
	s.tc.EXPECT().Args().Return([]string{})

	s.Require().NoError(s.bot.loginHandler(s.tc))
	s.Equal("login.usage", s.loc.id)
}

func (s *Suite) TestHandlerStart() {
	s.Require().NoError(s.bot.startHandler(s.tc))
	s.Equal("start.welcome", s.loc.id)
}

func (s *Suite) TestHandlerHelp() {
	s.Require().NoError(s.bot.helpHandler(s.tc))
	s.Equal("help.commands", s.loc.id)
}
