package page

import (
	"log"
	"html/template"
	"embed"
	"io"
	"fmt"
)

//go:embed index.gohtml
var indexTemplates embed.FS

var (
	index = Template{Template: template.New("")}
)

func init() {
	index.Funcs(template.FuncMap{
		"Render": Render,
		"MakeUniqueID": MakeUniqueID,
		"ObfuscateText": ObfuscateText,
		"CopyrightYear": CopyrightYear,
		"CopyrightYears": CopyrightYears,
	})
	template.Must(index.ParseFS(indexTemplates, "*.gohtml"))
	log.Printf("index: %s", index.DefinedTemplates())
}

type (
	IndexData struct {
		Site Site
		Listing PostListing
	}
	Weird string
)

func WriteIndex(w io.Writer, d IndexData) error {
	d.Site = SiteInfo // @todo
	return index.Execute(w, "index.gohtml", d)
}

func (w Weird) Render() (template.HTML, error) {
	return template.HTML(w.Text()), nil
}

func (w Weird) Text() string {
	return fmt.Sprintf(`<span class="weird">%s</span>`, string(w))
}
