package logger

import (
	"bytes"
	"fmt"
	"io"
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

// Init khởi tạo logger mặc định ra os.Stdout
func Init() {
	initWithWriter(os.Stdout)
}

// InitWithWriter cho phép inject writer — dùng khi cần mix với progress bar
func InitWithWriter(w io.Writer) {
	initWithWriter(w)
}

func initWithWriter(w io.Writer) {
	zerolog.TimeFieldFormat = time.RFC3339

	var capturedExtra []struct{ k, v string }

	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        w,
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
					delete(fields, k)
				}
			}
			return nil
		},
		FormatExtra: func(_ map[string]interface{}, buf *bytes.Buffer) error {
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
