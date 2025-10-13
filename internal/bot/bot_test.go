package bot

import (
	"context"
	"math/rand"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	tele "gopkg.in/telebot.v4"

	gmock "github.com/rumenvasilev/go-gitlab-mock/mock"

	"github.com/chaindead/review-flow-bot/internal/config"
	"github.com/chaindead/review-flow-bot/internal/db"
	"github.com/chaindead/review-flow-bot/internal/db/models"
	"github.com/chaindead/review-flow-bot/internal/gitlab"
	"github.com/chaindead/review-flow-bot/internal/lang"
	_ "github.com/chaindead/review-flow-bot/internal/logger"
)

var (
	ctx = context.Background()
)

type Suite struct {
	suite.Suite
	bot Bot
	loc *mockLocalizer
	inj do.Injector

	// per test
	tc *Context

	// cleanup
	users   []models.User
	teams   []models.Team
	members []models.TeamMember
	mrs     []models.MergeRequest
	reviews []models.MRReviewer
}

func (s *Suite) SetupSuite() {
	s.inj = do.New()
	s.loc = &mockLocalizer{}

	do.ProvideNamedValue(s.inj, "cfg.db", config.DB{File: ":memory:"})
	do.ProvideNamedValue(s.inj, "cfg.tg", config.TG{})
	do.ProvideValue(s.inj, &gitlab.Gitlab{})
	do.ProvideValue(s.inj, s.loc)
	do.Provide(s.inj, db.New)
	do.Provide(s.inj, New)

	b, err := do.InvokeStruct[Bot](s.inj)
	s.Require().NoError(err)
	s.bot = b
}

func (s *Suite) gitlab(opts ...gmock.MockBackendOption) {
	mockedURL := gmock.NewMockedHTTPServer(opts...)

	do.ProvideNamedValue(s.inj, "cfg.gitlab", config.Gitlab{URL: mockedURL})
	do.Override[*gitlab.Gitlab](s.inj, gitlab.New)

	git, err := do.Invoke[*gitlab.Gitlab](s.inj)
	s.Require().NoError(err)

	s.bot.git = git
}

func Test(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) SetupTest() {
	s.tc = NewContext(s.T())

	testUser := tele.User{
		ID:       rand.Int63(),
		Username: s.T().Name(),
	}
	s.tc.EXPECT().Sender().Return(&testUser).Maybe()

	s.tc.EXPECT().Send(mock.Anything).Return(nil).Maybe()
	s.tc.EXPECT().Send(mock.Anything, mock.Anything).Return(nil).Maybe()

	s.cleanDB()
	s.loc.reset()
}

func (s *Suite) cleanDB() {
	sql := `
	PRAGMA foreign_keys = OFF;

	-- Generate and execute delete statements for all user tables
	WITH tables AS (
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
	)
	SELECT 'DELETE FROM "' || name || '";' AS stmt FROM tables;

	PRAGMA foreign_keys = ON;
	`

	_, err := s.bot.db.DB().ExecContext(ctx, sql)
	s.NoError(err)
}

func (s *Suite) TearDownTest() {
	s.tc.AssertExpectations(s.T())
	s.True(s.loc.called)
}

type mockLocalizer struct {
	id   string
	args lang.Args

	called bool
}

func (m *mockLocalizer) Get(id string, args map[string]interface{}) string {
	m.called = true
	m.id = id
	m.args = args

	return "mocked"
}

func (m *mockLocalizer) reset() {
	m.called = false
	m.args = nil
	m.id = "unknown"
}
