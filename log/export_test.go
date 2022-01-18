package log

import "io"

func (s *StdOutputLogger) SetOut(o io.Writer) {
	s.out = o
}
