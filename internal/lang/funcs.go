package lang

import (
	"fmt"
	"strings"

	"github.com/chaindead/review-flow-bot/internal/db/models"
)

var mdEscaper = strings.NewReplacer(
	`_`, `\_`,
	`*`, `\*`,
	`[`, `\[`,
	`]`, `\]`,
	`(`, `\(`,
	`)`, `\)`,
	`~`, `\~`,
	"`", "\\`",
	`>`, `\>`,
	`#`, `\#`,
	`+`, `\+`,
	`-`, `\-`,
	`=`, `\=`,
	`|`, `\|`,
	`{`, `\{`,
	`}`, `\}`,
	`.`, `\.`,
	`!`, `\!`,
)

func escape(s string) string {
	return mdEscaper.Replace(s)
}

func role(s string) string {
	switch s {
	case models.RoleReviewer:
		return "👁"
	case models.RoleMember:
		return "👤"
	default:
		return "❔"
	}
}

func status(s string) string {
	switch s {
	case models.MRStatusApproved:
		return "👍"
	case models.MRStatusPending:
		return "⏳"
	case models.MRStatusRejected:
		return "👎"
	default:
		return "❔"
	}
}

func urle(t, u string) string {
	if strings.HasPrefix(u, "http://") || strings.Contains(u, "localhost") {
		u = ""
	}

	return fmt.Sprintf("[%s](%s)", escape(t), u)
}

func join(s []string, sep string) string {
	return strings.Join(s, sep)
}
