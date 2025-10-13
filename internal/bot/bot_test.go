package bot

import (
	"context"
	"math/rand"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/suite"
	tele "gopkg.in/telebot.v4"

	"github.com/rumenvasilev/go-gitlab-mock/mock"

	"github.com/chaindead/review-flow-bot/internal/config"
	"github.com/chaindead/review-flow-bot/internal/db"
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
	inj do.Injector

	// per test
	tc *Context
}

func (s *Suite) SetupSuite() {
	s.inj = do.New()

	do.ProvideNamedValue(s.inj, "cfg.db", config.DB{File: ":memory:"})
	do.ProvideNamedValue(s.inj, "cfg.tg", config.TG{})
	do.ProvideValue[*gitlab.Gitlab](s.inj, &gitlab.Gitlab{})
	do.Provide(s.inj, db.New)
	do.Provide(s.inj, lang.New)
	do.Provide(s.inj, New)

	b, err := do.InvokeStruct[Bot](s.inj)
	s.Require().NoError(err)
	s.bot = b
}

func (s *Suite) setGitlab(opts ...mock.MockBackendOption) {
	mockedURL := mock.NewMockedHTTPServer(opts...)

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
}

func (s *Suite) TearDownTest() {
	s.tc.AssertExpectations(s.T())
}
