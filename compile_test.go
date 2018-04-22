package grape

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var testExpr = `
li.search-results-item (data-docid ID)
	a.title (title Name)
	span.attribution
		a (href=/\/store\/apps\/developer\?id=(.+)/ Author)
	a.thumbnail
		img (src Thumb)
	?div.ratings (title=/Rating: ([0-9\.]+)/ Rating)
	span.buy-offer (data-docPriceMicros Price) (data-docCurrencyCode Currency)
	div.description (text Desc)
`

type App struct {
	ID       string
	Name     string
	Desc     string
	Thumb    string
	Author   string
	Rating   float64
	Installs int
	Ver      string
	Price    float64
	Currency string
}

func TestCompile(t *testing.T) {
	e, err := Compile(testExpr)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open("testdata/apps.html")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Document(f)
	if err != nil {
		t.Fatal(err)
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
		ioutil.WriteFile("compile_test.out", b, 777)*/
	}
}
