package lang

import (
	"embed"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/do/v2"
	"golang.org/x/text/language"
)

//go:embed *.toml
var fs embed.FS

var (
	funcs = map[string]interface{}{
		"bold": func(s string) string {
			return "*" + s + "*"
		},
		"italic": func(s string) string {
			return "_" + s + "_"
		},
		"code": func(s string) string {
			return "`" + s + "`"
		},
		"url": func(t, u string) string {
			return fmt.Sprintf("[%s](%s)", t, u)
		},
		"mention": func(t string, u int) string {
			return fmt.Sprintf("[%s](tg://user?id=%d)", t, u)
		},
		"par": func(s string) string {
			return `\(` + s + `\)`
		},
		"esc":    escape,
		"role":   role,
		"status": status,
		"urle":   urle,
		"join":   join,
	}
)

type Localizer struct {
	loc *i18n.Localizer
}

func New(_ do.Injector) (*Localizer, error) {
	bundle := i18n.NewBundle(language.Russian)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, err := bundle.LoadMessageFileFS(fs, "ru.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to load ru.toml: %w", err)
	}

	return &Localizer{loc: i18n.NewLocalizer(bundle, "ru")}, nil
}

var NoArgs = map[string]interface{}{}

type Args map[string]interface{}

func (l *Localizer) Get(id string, args map[string]interface{}) string {
	return l.loc.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: id,
		},
		TemplateData: args,
		Funcs:        funcs,
	})
}
