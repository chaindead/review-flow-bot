package config

import (
	"os"
	"sort"

	"github.com/caarlos0/env/v11"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
)

func Help() {
	params, err := env.GetFieldParams(&Config{})
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Environment Variable", "Default Value", "Current Value"})
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

		rows = append(rows, []string{field.Key, defaultValue, currentValue})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	return rows
}
