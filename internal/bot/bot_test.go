package bot

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bun"
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

	do.ProvideNamedValue(s.inj, "cfg.db", config.DB{File: ":memory:", Seed: true})
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
	var tables []string
	err := s.DB().NewSelect().
		With("tables", s.DB().NewSelect().
			Column("name").
			TableExpr("sqlite_master").
			Where("type = 'table' AND name NOT LIKE 'sqlite_%'"),
		).
		Table("tables").
		Column("name").
		Scan(ctx, &tables)
	s.NoError(err)

	// Disable FKs, delete all rows, re-enable FKs
	_, _ = s.DB().ExecContext(ctx, `PRAGMA foreign_keys = OFF;`)
	for _, t := range tables {
		_, err := s.DB().ExecContext(ctx, fmt.Sprintf(`DELETE FROM "%s";`, t))
		s.NoError(err)
	}
	_, _ = s.DB().ExecContext(ctx, `PRAGMA foreign_keys = ON;`)
}

func (s *Suite) TearDownTest() {
	s.tc.AssertExpectations(s.T())
	s.True(s.loc.called)
}

func (s *Suite) DB() *bun.DB {
	return s.bot.db.DB()
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
