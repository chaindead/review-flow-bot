package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/do/v2"

	"github.com/chaindead/review-flow-bot/internal/db/models"
	"github.com/chaindead/review-flow-bot/internal/lang"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).With().
		Timestamp().
		Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	injector := do.New()
	do.Provide(injector, lang.New)

	loc := do.MustInvoke[*lang.Localizer](injector)

	args := lang.Args{
		"Teams":      []models.Team{},
		"Unassigned": []models.User{},
	}
	payload := `{"Unassigned":[{"AuthoredMRs":null,"CreatedAt":"2025-10-08T20:06:25Z","GitlabID":111,"GitlabUsername":"gitlab_user-name","ID":111,"ReviewingMRs":null,"TeamMembership":null,"TelegramUsername":"telegram_user_name","UpdatedAt":"2025-10-08T20:06:25Z"}],"Teams":[{"CreatedAt":"2025-10-08T20:06:25Z","Members":[{"Role":"reviewer","Team":null,"TeamID":"backend","User":{"AuthoredMRs":null,"CreatedAt":"2025-10-08T20:06:25Z","GitlabID":6,"GitlabUsername":"reviewer2","ID":7076621751,"ReviewingMRs":null,"TeamMembership":null,"TelegramUsername":"chaindead","UpdatedAt":"2025-10-08T20:06:25Z"},"UserID":7076621751},{"Role":"reviewer","Team":null,"TeamID":"backend","User":{"AuthoredMRs":null,"CreatedAt":"2025-10-08T20:06:25Z","GitlabID":5,"GitlabUsername":"reviewer1","ID":268850924,"ReviewingMRs":null,"TeamMembership":null,"TelegramUsername":"doejon","UpdatedAt":"2025-10-08T20:06:25Z"},"UserID":268850924},{"Role":"member","Team":null,"TeamID":"backend","User":{"AuthoredMRs":null,"CreatedAt":"2025-10-08T20:06:25Z","GitlabID":4,"GitlabUsername":"developer2","ID":5875418647,"ReviewingMRs":null,"TeamMembership":null,"TelegramUsername":"letztes","UpdatedAt":"2025-10-08T20:06:25Z"},"UserID":5875418647}],"Name":"backend"},{"CreatedAt":"2025-10-08T20:06:25Z","Members":null,"Name":"frontend"}]}`
	_ = json.Unmarshal([]byte(payload), &args)
	log.Info().Interface("data", args).Send()

	text := loc.Get("admin.list.teams", args)

	fmt.Println(text)

}
