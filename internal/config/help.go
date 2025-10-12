package config

import (
	"os"
	"sort"

	"github.com/caarlos0/env/v11"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/rs/zerolog/log"
)

var descriptionMap = map[string]string{
	"DB_DB_FILE":          "path to database file",
	"DB_SEED":             "populate db with fake data for testing purposes",
	"GITLAB_TOKEN":        "token should have accesses [api, read_api, read_repository]",
	"GITLAB_URL":          "main url for your gitlab instance",
	"TG_ADMIN_IDS":        "tg ids of users, who have access to admin commands",
	"TG_TOKEN":            "tg bot token",
	"WATCH_POLL_INTERVAL": "period of watching MR status updates",
}

func Help() {
	params, err := env.GetFieldParams(&Config{})
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	table := tablewriter.NewTable(
		os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{Alignment: tw.CellAlignment{
				PerColumn: []tw.Align{
					"",
					"",
					"",
					"",
					tw.AlignCenter},
			},
			},
		}))
	table.Header([]string{
		"Environment Variable",
		"Description",
		"Default Value",
		"Current Value",
		"Valid",
	})
	for _, row := range newTableRows(params) {
		table.Append(row)
	}

	table.Render()
}

func newTableRows(fields []env.FieldParams) [][]string {
	const undefined = ""

	rows := make([][]string, 0, len(fields))
	for _, field := range fields {
		defaultValue := undefined
		if field.HasDefaultValue {
			defaultValue = field.DefaultValue
		}

		currentValue := undefined
		if val, ok := os.LookupEnv(field.Key); ok {
			currentValue = val
		}

		valid := "✔"
		if field.Required && currentValue == undefined {
			valid = "✖✖✖"
		}

		rows = append(rows, []string{
			field.Key,
			descriptionMap[field.Key],
			defaultValue,
			currentValue,
			valid,
		})
		delete(descriptionMap, field.Key) // for check unused
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	if len(descriptionMap) != 0 {
		log.Warn().Interface("unused", descriptionMap).Msg("found unused descriptions")
	}

	return rows
}
