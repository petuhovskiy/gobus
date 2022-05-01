package cyprusbus

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/net/html"
)

func debug(sel *goquery.Selection) {
	b := &bytes.Buffer{}
	_ = html.Render(b, sel.Nodes[0])
	spew.Dump(b)
}
