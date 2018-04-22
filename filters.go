package grape

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"
)

type FilterMap map[string]func(s string) string

var regexpTag = regexp.MustCompile(`<[^>]+>`)

var builtins = map[string]reflect.Value{
	"lower": reflect.ValueOf(func(s string) string {
		return strings.ToLower(s)
	}),
	"upper": reflect.ValueOf(func(s string) string {
		return strings.ToUpper(s)
	}),
	"strip": reflect.ValueOf(func(s string) string {
		return strings.TrimSpace(s)
	}),
	"nocomma": reflect.ValueOf(func(s string) string {
		return strings.Replace(s, ",", "", -1)
	}),

	"notags": reflect.ValueOf(func(s string) string {
		return regexpTag.ReplaceAllString(s, "")
	}),
	"tag2space": reflect.ValueOf(func(s string) string {
		return regexpTag.ReplaceAllString(s, " ")
	}),
}

func addFilters(out map[string]reflect.Value, in FilterMap) {
	for name, filter := range in {
		if !goodName(name) {
			panic(fmt.Errorf("filter name %q is not a valid identifier", name))
		}
		out[name] = reflect.ValueOf(filter)
	}
}

func findFilter(name string, expr *Expr) (reflect.Value, bool) {
	if expr != nil && expr.filters != nil {
		expr.muFilters.RLock()
		defer expr.muFilters.RUnlock()
		if filter := expr.filters[name]; filter.IsValid() {
			return filter, true
		}
	}
	if filter := builtins[name]; filter.IsValid() {
		return filter, true
	}
	return reflect.Value{}, false
}

// goodName reports whether the function name is a valid identifier.
func goodName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		switch {
		case r == '_':
		case i == 0 && !unicode.IsLetter(r):
			return false
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			return false
		}
	}
	return true
}
