# Grape, a CSS-like language for scraping HTML documents

Example:
```go
package main

import (
	"log"
	"strings"

	"github.com/zippoxer/grape"
)

const document = `
<ul>
	<li>
		<h1>Tesla Model S</h1>
		<h2>FEATURES:</h2>
		<ul>
			<li>FAST</li>
			<li>BEAUTIFUL</li>
			<li>ELECTRIC</li>
		</ul>
		<span>Price: $70,000</span>
		<a href="https://tesla.com">Details</a>
	</li>
</ul>
`

type Car struct {
	Name     string
	Features []string
	Price    float64
	Link     string
}

var grapeExpression = `
li
	h1 (text Name)
	ul
		li (text Features.lower)
	span (text=/Price: \$(.+)/ Price.nocomma)
	a (href Link)
`

func main() {
	// Parse HTML document.
	doc, err := grape.Document(strings.NewReader(document))
	if err != nil {
		log.Fatal(err)
	}
	// Compile Grape expression.
	expr, err := grape.Compile(grapeExpression)
	if err != nil {
		log.Fatal(err)
	}
	// Scrape.
	var cars []Car
	err = expr.Find(doc, &cars)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v", cars)
}
```