package page

import (
	"log"
	"html/template"
	"embed"
	"io"
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
)

func WriteIndex(w io.Writer, d IndexData) error {
	return index.Execute(w, "index.gohtml", d)
}
