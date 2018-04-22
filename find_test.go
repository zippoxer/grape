package grape

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/andybalholm/cascadia"
)

func TestFind(t *testing.T) {
	f, err := os.Open("testdata/apps.html")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Document(f)
	if err != nil {
		t.Fatal(err)
	}
	e := &Expr{
		root: true,
		children: []*Expr{
			{
				selector: cascadia.MustCompile("li.search-results-item"),
				captures: []capture{
					{
						attr:    "data-docid",
						targets: []target{{"ID", nil}},
					},
				},
				children: []*Expr{
					&Expr{
						selector: cascadia.MustCompile("a.title"),
						captures: []capture{
							{
								attr:    "title",
								targets: []target{{"Name", nil}},
							},
						},
					},
					&Expr{
						selector: cascadia.MustCompile("span.attribution"),
						children: []*Expr{
							&Expr{
								selector: cascadia.MustCompile("a"),
								captures: []capture{
									{
										attr:    "href",
										re:      regexp.MustCompile(`/store/apps/developer\?id=(.+)`),
										targets: []target{{"Author", nil}},
									},
								},
							},
						},
					},
					&Expr{
						selector: cascadia.MustCompile("a.thumbnail"),
						children: []*Expr{
							{
								selector: cascadia.MustCompile("img"),
								captures: []capture{
									{
										attr:    "src",
										targets: []target{{"Thumb", nil}},
									},
								},
							},
						},
					},
					&Expr{
						selector: cascadia.MustCompile("div.ratings"),
						captures: []capture{
							{
								attr:    "title",
								re:      regexp.MustCompile(`Rating: ([0-9\.]+)`),
								targets: []target{{"Rating", nil}},
							},
						},
						optional: true,
					},
					&Expr{
						selector: cascadia.MustCompile("span.buy-offer"),
						captures: []capture{
							{
								attr:    "data-doccurrencycode",
								targets: []target{{"Currency", nil}},
							},
							{
								attr:    "data-docpricemicros",
								targets: []target{{"Price", nil}},
							},
						},
					},
					&Expr{
						selector: cascadia.MustCompile("div.description"),
						captures: []capture{
							{
								attr:    "text",
								targets: []target{{"Desc", nil}},
							},
						},
					},
				},
			},
		},
	}
	var actual []App
	if err := e.Find(doc, &actual); err != nil {
		t.Fatal(err)
	}
	var expected []App
	b, err := ioutil.ReadFile("testdata/apps.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &expected); err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(actual, expected) {
		t.Error("not what expected!")
		/*b, err := json.MarshalIndent(actual, "", "\t")
		if err != nil {
			t.Fatal(err)
		}
		ioutil.WriteFile("find_test.out", b, 777)*/
	}
}
