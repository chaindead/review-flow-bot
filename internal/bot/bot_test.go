package bot

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/suite"

	"github.com/chaindead/review-flow-bot/internal/config"
	"github.com/chaindead/review-flow-bot/internal/db"
	"github.com/chaindead/review-flow-bot/internal/gitlab"
	"github.com/chaindead/review-flow-bot/internal/lang"
)

type Suite struct {
	suite.Suite
	bot Bot

	tc *Context
}

func (s *Suite) SetupSuite() {
	injector := do.New()

	do.ProvideNamedValue(injector, "cfg.db", config.DB{File: ":memory:"})
	do.ProvideNamedValue(injector, "cfg.tg", config.TG{})
	do.ProvideValue[*gitlab.Gitlab](injector, &gitlab.Gitlab{})
	do.Provide(injector, db.New)
	do.Provide(injector, lang.New)
	do.Provide(injector, New)

	b, err := do.InvokeStruct[Bot](injector)
	s.Require().NoError(err)
	s.bot = b
}

func Test(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) SetupTest() {
	s.tc = NewContext(s.T())
}

func (s *Suite) TearDownTest() {
	s.tc.AssertExpectations(s.T())
}
