package options

type LoggerOptions struct {
	Extras    any
	Protocol  Protocol
	RequestID string
	UserID    string
}

type Protocol string

type LoggerOption func(*LoggerOptions)

func WithExtras(data any) LoggerOption {
	return func(lo *LoggerOptions) {
		lo.Extras = data
	}
}

func WithUserID(data string) LoggerOption {
	return func(lo *LoggerOptions) {
		lo.UserID = data
	}
}
