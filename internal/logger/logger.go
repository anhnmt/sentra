package logger

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var systemFields = map[string]bool{
	zerolog.TimestampFieldName: true,
	zerolog.LevelFieldName:     true,
	zerolog.MessageFieldName:   true,
	zerolog.CallerFieldName:    true,
}

func Init() {
	zerolog.TimeFieldFormat = time.RFC3339

	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			zerolog.MessageFieldName,
		},
		FormatExtra: func(fields map[string]interface{}, buf *bytes.Buffer) error {
			var hasFields bool
			for k := range fields {
				if !systemFields[k] {
					hasFields = true
					break
				}
			}

			if !hasFields {
				return nil
			}

			for k, v := range fields {
				if systemFields[k] {
					continue
				}
				fmt.Fprintf(buf, "\n  %-14s %v", k+":", v)
			}
			buf.WriteString("\n")
			return nil
		},
	}).
		With().
		Timestamp().
		Logger()
}
