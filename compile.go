package grape

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/andybalholm/cascadia"
)

func Compile(expr string) (*Expr, error) {
	s := &scanner{src: expr}
	e := &Expr{root: true}
	if err := compile(0, e, -1, s); err != nil {
		return nil, err
	}
	return e, nil
}

func MustCompile(expr string) *Expr {
	e, err := Compile(expr)
	if err != nil {
		panic(fmt.Errorf("grape.MustCompile: %v", err))
	}
	return e
}

func compile(rootIndent int, root *Expr, tk int, s *scanner) error {
	if tk == -1 {
		var err error
		tk, err = s.scan()
		if err != nil {
			return fmt.Errorf("line %d: %s", s.line, err)
		}
		if tk == tkEOF {
			return nil
		}
	}
	var indent int
	if tk == tkIndent {
		indent = len(s.data())
	}
	if indent < rootIndent {
		return nil
	}
	if indent > rootIndent {
		err := compile(indent, root.children[len(root.children)-1], tk, s)
		if err != nil {
			return err
		}
	}
	e := &Expr{}
	var capt capture
	for {
		switch tk {
		case tkSelector:
			var err error
			e.selector, err = cascadia.Compile(s.data())
			if err != nil {
				return err
			}
		case tkLeftBracket:
			capt = capture{}
		case tkRightBracket:
			e.captures = append(e.captures, capt)
		case tkQuestionMark:
			e.optional = true
		case tkIdentifier:
			if capt.attr == "" {
				capt.attr = strings.ToLower(s.data())
				break
			}
			a := strings.Split(s.data(), ".")
			t := target{name: a[0]}
			if len(a) > 1 {
				t.filters = a[1:]
			}
			capt.targets = append(capt.targets, t)
		case tkRegexp:
			var err error
			capt.re, err = regexp.Compile(s.regexp.String())
			if err != nil {
				return fmt.Errorf("line %d: %s", s.line, err)
			}
		case tkLineBreak:
			if e.selector != nil {
				root.children = append(root.children, e)
			}
			return compile(rootIndent, root, -1, s)
		case tkEOF:
			if e.selector != nil {
				root.children = append(root.children, e)
			}
			return nil
		}
		var err error
		tk, err = s.scan()
		if err != nil {
			return fmt.Errorf("line %d: %s", s.line, err)
		}
	}
}
