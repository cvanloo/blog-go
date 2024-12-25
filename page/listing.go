package page

import (
	"fmt"
	"log"
	"io"
	"html/template"
	"embed"
)

//go:embed listing.gohtml
var listingTemplates embed.FS

var (
	listing = Template{Template: template.New("")}
)

func init() {
	listing.Funcs(template.FuncMap{
		"Render": Render,
		"MakeUniqueID": MakeUniqueID,
		"ObfuscateText": ObfuscateText,
		"CopyrightYear": CopyrightYear,
		"CopyrightYears": CopyrightYears,
	})
	template.Must(listing.ParseFS(listingTemplates, "*.gohtml"))
	log.Printf("listing: %s", listing.DefinedTemplates())
}

type (
	ListingData struct {
		Site Site
		UrlPath string
		Title StringRenderable
		Abstract StringRenderable
		Listing PostListing
	}
)

func WriteListing(w io.Writer, d ListingData) error {
	return listing.Execute(w, "listing.gohtml", d)
}

func (l ListingData) Canonical() string {
	return fmt.Sprintf("%s://%s/%s", l.Site.Address.Scheme, l.Site.Address.Host, l.UrlPath)
}
