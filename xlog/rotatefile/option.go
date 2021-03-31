package rotatefile

import "time"

var UTC = clockFunc(func() time.Time { return time.Now().UTC() })
var Local = clockFunc(time.Now)

type Option interface {
	apply(*RotateFile)
}

type optionFunc func(*RotateFile)

func (f optionFunc) apply(rf *RotateFile) { f(rf) }

type Clock interface {
	Now() time.Time
}

type clockFunc func() time.Time

func (f clockFunc) Now() time.Time {
	return f()
}

func (rf *RotateFile) withOption(options ...Option) {
	for _, option := range options {
		option.apply(rf)
	}
}

func WithDir(dir string) Option {
	return optionFunc(func(rf *RotateFile) {
		rf.dir = dir
	})
}

func WithBiz(biz string) Option {
	return optionFunc(func(rf *RotateFile) {
		rf.biz = biz
	})
}

func WithClock(clock Clock) Option {
	return optionFunc(func(rf *RotateFile) {
		rf.clock = clock
	})
}

func WithSeverity(severity string) Option {
	return optionFunc(func(rf *RotateFile) {
		rf.severity = severity
	})
}

func WithTimeLayout(layout string) Option {
	return optionFunc(func(rf *RotateFile) {
		rf.timeLayout = layout
	})
}

func WithMaxSize(maxSize int64) Option {
	return optionFunc(func(rf *RotateFile) {
		rf.maxSize = maxSize
	})
}

func WithPeroid(peroid time.Duration) Option {
	return optionFunc(func(rf *RotateFile) {
		if peroid > 24*time.Hour {
			peroid = 24 * time.Hour
		} else if peroid < time.Minute {
			peroid = time.Minute
		}
		rf.peroid = peroid
	})
}

func WithMaxCount(maxCount int) Option {
	return optionFunc(func(rf *RotateFile) {
		rf.maxCount = maxCount
	})
}

func WithMaxAge(age time.Duration) Option {
	return optionFunc(func(rf *RotateFile) {
	})
}
