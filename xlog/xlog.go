package xlog

type Logger struct {
	ctlr Ctlr
}

func (l *Logger) clone() *Logger {
	copy := *l
	return &copy
}

func (l *Logger) check(lvl Level, msg string) *Entry {
	const callerSkipOffset = 2
	if lvl < DPanicLevel && !l.ctlr.Enabled(lvl) {
		return nil
	}

	entry := Entry{}

	return &entry
}

func (l *Logger) Debug(msg string, fields ...Field) {

}

func (l *Logger) Info(msg string, fields ...Field) {

}

func (l *Logger) Warn(msg string, fields ...Field) {

}

func (l *Logger) Error(msg string, fields ...Field) {

}

func (l *Logger) DPanic(msg string, fields ...Field) {

}

func (l *Logger) Panic(msg string, fields ...Field) {

}

func (l *Logger) Fatal(msg string, fields ...Field) {

}
