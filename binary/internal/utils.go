package internal

import (
	"encoding/json"

    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark/extension"
    "github.com/yuin/goldmark/parser"
    "github.com/yuin/goldmark/renderer/html"

	highlighting "github.com/yuin/goldmark-highlighting/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	img64 "github.com/tenkoh/goldmark-img64"

	"bytes"
	"fmt"
)

func ToMarkdown(input string) string {
    md := goldmark.New(goldmark.WithExtensions(

            // extension list
            img64.Img64,
            extension.GFM,

            // syntax highlighting extension config
            highlighting.NewHighlighting(

                highlighting.WithStyle("monokai"),
                highlighting.WithFormatOptions( chromahtml.WithLineNumbers(true) ),

            ),
        ),

		goldmark.WithParserOptions(parser.WithAutoHeadingID()),

		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),  // allow rendering of raw HTML in md file
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		panic(err)
	}
	return buf.String()
}

// JSON Object sent over TCP. Events described in ../networking/utils
type Obj struct {
	Event   string `json:"event"`
	Content string `json:"content"`
}

func EventsJSON(jsonObj []byte) Obj {
	var obj Obj

	err := json.Unmarshal(jsonObj, &obj)
	if err != nil {
		fmt.Printf("error unmarshalling json: %s", err)
	}

	return obj
}

type Page struct {
	content string
}

func newPage() *Page {
	base := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Live Preview</title>
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  `
	return &Page{content: base}
}

func scripts() string {
  return `<script>
  const evtSource = new EventSource("/events");

	  evtSource.addEventListener("reload", (event) => {
		  const data = event.data.replace(/<br>/g, "\n");
		  let view = document.getElementById("container");
		  view.innerHTML = data;
		});

	  evtSource.addEventListener("scroll", (event) => {
		  let percentAmount = parseFloat(event.data);
		  var body = document.body,
		  html = document.documentElement;

		  var height = Math.max( body.scrollHeight, body.offsetHeight, 
			  html.clientHeight, html.scrollHeight, html.offsetHeight );
		  
		  let yLoc = percentAmount * height;
		  window.scrollTo({
			  top: yLoc,
			  left: 0,
		  });

	  });
  </script>
  </head>
<body>
<div id="container">`
}

// helper function for page layout
func inject(p *Page, payload string) {
	p.content += "\n" + payload
}

func LayoutPage(html string) string {
	page := newPage()
	inject(page, Css())
	inject(page, scripts())
	inject(page, html)
	ending := `</div></body>
</html>`	
	inject(page, ending)
	return page.content
}
