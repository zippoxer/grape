package grape

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"golang.org/x/net/html"
)

var ErrNoMatch = errors.New("no match")

func setStructField(name string, val reflect.Value, s reflect.Value) error {
	f := s.FieldByName(name)
	if val.Type().AssignableTo(f.Type()) {
		f.Set(val)
		return nil
	}
	switch val.Kind() {
	case reflect.String:
		switch f.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(val.Interface().(string), 10, 64)
			if err != nil {
				return err
			}
			if f.OverflowInt(n) {
				return fmt.Errorf("cannot assign to %s: overflow", name)
			}
			f.SetInt(n)
			return nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			n, err := strconv.ParseUint(val.Interface().(string), 10, 64)
			if err != nil {
				return err
			}
			if f.OverflowUint(n) {
				return fmt.Errorf("cannot assign to %s: overflow", name)
			}
			f.SetUint(n)
			return nil
		case reflect.Float32, reflect.Float64:
			n, err := strconv.ParseFloat(val.Interface().(string), f.Type().Bits())
			if err != nil {
				return err
			}
			if f.OverflowFloat(n) {
				return fmt.Errorf("cannot assign to %s: overflow", name)
			}
			f.SetFloat(n)
			return nil
		// TODO: temporary fix, won't work in the long run.
		case reflect.Slice:
			f.Set(reflect.Append(f, val))
			return nil
		}
	}
	return fmt.Errorf("cannot assign to %s: %s is not assignable to %s", name, val.Type(), f.Type())
}

func find(e *Expr, root *html.Node, v reflect.Value) error {
	if e.root {
		for _, ce := range e.children {
			if err := find(ce, root, v); err != nil {
				return err
			}
		}
		return nil
	}
	elem := v
	if elem.Kind() == reflect.Ptr {
		elem = reflect.Indirect(elem)
	}
	var slice bool
	if elem.Kind() == reflect.Slice {
		elem = reflect.Indirect(reflect.New(elem.Type().Elem()))
		slice = true
	}
	var findings int
loop:
	for n := root.FirstChild; n != nil; n = n.NextSibling {
		err := find(e, n, v)
		if err != nil && err != ErrNoMatch {
			return err
		}
		if err != ErrNoMatch {
			findings++
		}
		if !e.selector(n) {
			continue
		}
		if slice {
			elem = reflect.Indirect(reflect.New(elem.Type()))
		}
		for _, c := range e.captures {
			var attrVal string
			switch c.attr {
			case "text":
				attrVal = getNodeText(n, nil)
			case "html":
				var err error
				attrVal, err = getNodeHTML(n)
				if err != nil {
					return err
				}
			default:
				var exists bool
				for _, a := range n.Attr {
					if a.Key == c.attr {
						attrVal = a.Val
						exists = true
						break
					}
				}
				if !exists {
					continue loop
				}
			}
			if c.re != nil {
				match := c.re.FindStringSubmatch(attrVal)
				if len(match) == 0 {
					continue loop
				}
				for i := range match[1:] {
					assignment := reflect.ValueOf(match[i+1])
					for _, filterName := range c.targets[i].filters {
						filter, ok := findFilter(filterName, e)
						if !ok {
							return fmt.Errorf("%q is not a defined filter", filterName)
						}
						assignment = filter.Call([]reflect.Value{assignment})[0]
					}
					if err := setStructField(c.targets[i].name, assignment, elem); err != nil {
						return err
					}
				}
				continue
			}
			assignment := reflect.ValueOf(attrVal)
			for _, filterName := range c.targets[0].filters {
				filter, ok := findFilter(filterName, e)
				if !ok {
					return fmt.Errorf("%q is not a defined filter", filterName)
				}
				assignment = filter.Call([]reflect.Value{assignment})[0]
			}
			if err := setStructField(c.targets[0].name, assignment, elem); err != nil {
				return err
			}
		}
		for _, ce := range e.children {
			err := find(ce, n, elem)
			if err == ErrNoMatch {
				continue loop
			}
			if err != nil {
				return err
			}
		}
		findings++
		if slice {
			iv := reflect.Indirect(v)
			iv.Set(reflect.Append(iv, elem))
		}
	}
	if e.optional || findings > 0 {
		return nil
	}
	return ErrNoMatch
}

func getNodeText(node *html.Node, buf *bytes.Buffer) string {
	if node.Type == html.TextNode {
		return node.Data
	}
	if node.FirstChild == nil {
		return ""
	}
	if buf == nil {
		buf = &bytes.Buffer{}
	}
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		buf.WriteString(getNodeText(n, buf))
	}
	return buf.String()
}

func getNodeHTML(node *html.Node) (string, error) {
	var buf bytes.Buffer
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		if err := html.Render(&buf, n); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}
