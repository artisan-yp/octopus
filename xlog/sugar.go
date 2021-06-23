package xlog

type SugaredLogger struct {
	base *Logger
}

func (s *SugaredLogger) Desugar() *Logger {
	base := s.base.clone()
	return base
}
