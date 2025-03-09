package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"

	"github.com/vadicheck/gofermart/pkg/logger/options"
)

type JSONFormatter struct {
	TimestampFormat string

	DisableTimestamp bool

	PrettyPrint bool
}

type LogDataFields struct {
	Time   string         `json:"time"`
	Level  string         `json:"level"`
	Msg    LogDataMessage `json:"msg"`
	Labels *LogDataLabels `json:"labels,omitempty"`
}

type LogDataLabels struct {
	UserID string `json:"user_id,omitempty"`
}

type LogDataMessage struct {
	Message string `json:"message,omitempty"`
	Extras  any    `json:"extras,omitempty"`
	Error   string `json:"error,omitempty"`
	Func    string `json:"func,omitempty"`
	File    string `json:"file,omitempty"`
}

func (f *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := LogDataFields{}

	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			data.Msg.Error = v.Error()
		default:
			if k == LogOptionsField {
				opts, ok := v.(options.LoggerOptions)
				if ok {
					showLabels := opts.UserID != ""
					showExtras := !(opts.Extras == nil ||
						reflect.ValueOf(opts.Extras).IsNil())
					if showLabels {
						data.Labels = &LogDataLabels{}
						data.Labels.UserID = opts.UserID
					}
					if showExtras {
						data.Msg.Extras = opts.Extras
					}
				}
			}
		}
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = DefaultTimestampFormat
	}

	if !f.DisableTimestamp {
		data.Time = entry.Time.Format(timestampFormat)
	}
	data.Msg.Message = entry.Message
	data.Level = entry.Level.String()

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	encoder := json.NewEncoder(b)
	if f.PrettyPrint {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %w", err)
	}

	return b.Bytes(), nil
}
