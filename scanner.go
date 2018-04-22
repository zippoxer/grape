package grape

import (
	"bytes"
	"fmt"
)

const (
	tkIndent = iota
	tkQuestionMark
	tkIdentifier
	tkSelector
	tkRegexp
	tkLeftBracket
	tkRightBracket
	tkLineBreak
	tkEOF
)

type scanner struct {
	start, i int
	src      string
	line     int

	inBrackets bool
	regexp     bytes.Buffer
}

func (s *scanner) prev() {
	s.i--
	if s.src[s.i] == '\n' {
		s.line--
	}
}

func (s *scanner) next() (byte, bool) {
	if s.i < len(s.src) {
		c := s.src[s.i]
		if c == '\n' {
			s.line++
		}
		s.i++
		return c, false
	}
	return 0, true
}

func (s *scanner) pos() (int, int, int) {
	return s.line, s.start, s.i
}

func (s *scanner) data() string {
	return s.src[s.start:s.i]
}

func (s *scanner) scan() (int, error) {
	s.start = s.i
	c, eof := s.next()
	if eof {
		return tkEOF, nil
	}
	switch c {
	case '\t':
		if s.skip(func(c byte) bool { return c != '\t' }) {
			return tkEOF, nil
		}
		return tkIndent, nil
	case '\n':
		if s.skip(func(c byte) bool { return c != ' ' && c != '\n' }) {
			return tkEOF, nil
		}
		return tkLineBreak, nil
	case ' ':
		if s.skip(func(c byte) bool { return c != ' ' && c != '\n' }) {
			return tkEOF, nil
		}
	case '(':
		s.inBrackets = true
		return tkLeftBracket, nil
	case ')':
		s.inBrackets = false
		return tkRightBracket, nil
	case '=':
		c, eof := s.next()
		if eof {
			return tkEOF, fmt.Errorf("unexpected EOF")
		}
		if c != '/' {
			return tkEOF, fmt.Errorf("expected '/', found %q", c)
		}
		s.regexp.Reset()
		var escape bool
		for {
			c, eof := s.next()
			if eof {
				return tkEOF, fmt.Errorf("unexpected EOF")
			}
			if escape {
				switch c {
				case '/':
					s.regexp.WriteByte('/')
				case '\\':
					s.regexp.WriteByte('\\')
				case 'n':
					s.regexp.WriteByte('\n')
				case 't':
					s.regexp.WriteByte('\t')
				default:
					s.regexp.WriteByte('\\')
					s.regexp.WriteByte(c)
				}
				escape = false
			} else {
				switch c {
				case '/':
					return tkRegexp, nil
				case '\\':
					escape = true
				default:
					s.regexp.WriteByte(c)
				}
			}
		}
	case '?':
		return tkQuestionMark, nil
	default:
		if s.inBrackets {
			s.skip(func(c byte) bool {
				return c == '\t' || c == ' ' || c == '\n' || c == '(' || c == ')' || c == '='
			})
			return tkIdentifier, nil
		}
		s.skip(func(c byte) bool { return c == '\t' || c == ' ' || c == '\n' })
		return tkSelector, nil
	}
	return s.scan()
}

func (s *scanner) skip(stop func(byte) bool) bool {
	for {
		c, eof := s.next()
		if eof {
			return true
		}
		if stop(c) {
			s.prev()
			return false
		}
	}
}

func (s *scanner) skipWhitespace() bool {
	return s.skip(func(c byte) bool { return c != '\n' && c != '\t' && c != ' ' })
}
