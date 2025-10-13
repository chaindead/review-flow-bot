package bot

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/stretchr/testify/mock"
	tele "gopkg.in/telebot.v4"
)

func (s *Suite) TestHandlerID() {
	uid := rand.Int63()

	s.tc.EXPECT().Sender().Return(&tele.User{ID: uid})
	s.tc.EXPECT().
		Send(
			mock.MatchedBy(func(s string) bool {
				return strings.Contains(s, strconv.FormatInt(uid, 10))
			}),
			mock.Anything).
		Return(nil)

	s.Require().NoError(s.bot.idHandler(s.tc))
}
