package xlog

type Config struct {
}

type EncoderConfig struct {
	MessageKey    string
	LevelKey      string
	TimeKey       string
	NameKey       string
	CallerKey     string
	FunctionKey   string
	StacktraceKey string
}
