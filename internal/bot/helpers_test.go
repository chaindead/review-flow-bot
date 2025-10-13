package bot

import (
	"math/rand"

	tele "gopkg.in/telebot.v4"

	"github.com/chaindead/review-flow-bot/internal/db/models"
)

func rndString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a' + byte(rand.Intn(26))
	}

	return string(b)
}

func (s *Suite) NoError2(_ any, err error) {
	s.Require().NoError(err)
}

func (s *Suite) Run(fn func(c tele.Context) error) {
	s.Require().NoError(fn(s.tc))
}

func (s *Suite) createUser() models.User {
	s.T().Helper()

	u := models.User{
		ID:               rand.Int63(),
		GitlabUsername:   "gitlab_" + rndString(8),
		TelegramUsername: "telegram_" + rndString(8),
		GitlabID:         rand.Int63(),
	}

	_, err := s.DB().NewInsert().Model(&u).Exec(ctx)
	s.Require().NoError(err)

	return u
}
