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

	// Capture extra fields before ConsoleWriter touches them
	var capturedExtra []struct{ k, v string }

	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		PartsOrder: []string{
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			zerolog.MessageFieldName,
		},
		FormatPrepare: func(fields map[string]interface{}) error {
			capturedExtra = capturedExtra[:0]
			for k, v := range fields {
				if !systemFields[k] {
					capturedExtra = append(capturedExtra, struct{ k, v string }{
						k: k,
						v: fmt.Sprintf("%v", v),
					})
					// Remove so ConsoleWriter won't auto-append to message line
					delete(fields, k)
				}
			}
			return nil
		},
		FormatExtra: func(fields map[string]interface{}, buf *bytes.Buffer) error {
			if len(capturedExtra) == 0 {
				return nil
			}
			for _, f := range capturedExtra {
				fmt.Fprintf(buf, "\n  %-14s %s", f.k+":", f.v)
			}
			buf.WriteString("\n")
			return nil
		},
	}).
		With().
		Timestamp().
		Logger()
}
