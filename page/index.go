package page

import (
	"log"
	"html/template"
	"embed"
	"io"
	"fmt"
	"sort"
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
		"UrlEscapeLower": UrlEscapeLower,
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
	sort.Slice(d.Listing, func(i, j int) bool {
		p1 := d.Listing[i].Published.Published
		p2 := d.Listing[j].Published.Published
		return p1.Compare(p2) > 0 // reverse chronological listing
	})
	return index.Execute(w, "index.gohtml", d)
}

func (w Weird) Render() (template.HTML, error) {
	return template.HTML(w.Text()), nil
}

func (w Weird) Text() string {
	return fmt.Sprintf(`<span class="weird">%s</span>`, string(w))
}

func (i IndexData) ObfuscatedAuthorCredit() (template.HTML, error) {
	authorName, err := i.Site.Owner.Render()
	if err != nil {
		return "", err
	}
	return template.HTML(fmt.Sprintf(`<a href="mailto:%s">%s</a>`, ObfuscateText(i.Site.Email), authorName)), nil
}
