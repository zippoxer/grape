package grape

import (
	"io"
	"net/http"
	"reflect"
	"regexp"
	"sync"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

type target struct {
	name    string
	filters []string
}

type capture struct {
	attr    string
	re      *regexp.Regexp
	targets []target
}

// Expr is the representation of a compiled Grape expression.
// An Expr is safe for concurrent use by multiple goroutines.
type Expr struct {
	root      bool
	selector  cascadia.Selector
	captures  []capture
	optional  bool
	children  []*Expr
	filters   map[string]reflect.Value
	muFilters sync.RWMutex
}

// Filter adds filters from the given FilterMap to Expr's filter map.
// If the expression calls non-builtin filters, Filter must be called before Find.
// Panics if a filter in the map has an invalid name.
func (e *Expr) Filters(filterMap FilterMap) {
	e.muFilters.Lock()
	defer e.muFilters.Unlock()
	if e.filters == nil {
		e.filters = make(map[string]reflect.Value)
	}
	addFilters(e.filters, filterMap)
}

func (e *Expr) Find(doc *Doc, v interface{}) error {
	return find(e, doc.root, reflect.ValueOf(v))
}

// Doc represents an HTML document.
type Doc struct {
	root *html.Node
}

//
func Document(r io.Reader) (*Doc, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	return &Doc{root: doc}, nil
}

func Fetch(url string) (*Doc, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return Document(resp.Body)
}
