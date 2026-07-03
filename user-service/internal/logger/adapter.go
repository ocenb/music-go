package logger

import (
	"fmt"
	"log/slog"
)

type SlogAdapter struct {
	log *slog.Logger
}

func NewSlogAdapter(log *slog.Logger) *SlogAdapter {
	return &SlogAdapter{log: log}
}

func (s *SlogAdapter) Printf(format string, v ...any) {
	s.log.Info(fmt.Sprintf(format, v...))
}

func (s *SlogAdapter) Fatalf(format string, v ...any) {
	s.log.Error(fmt.Sprintf(format, v...))
}
