package page

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"sort"
)

//go:embed listing.gohtml
var listingTemplates embed.FS

var (
	listing = Template{Template: template.New("")}
)

func init() {
	listing.Funcs(template.FuncMap{
		"Render":         Render,
		"MakeUniqueID":   MakeUniqueID,
		"ObfuscateText":  ObfuscateText,
		"CopyrightYear":  CopyrightYear,
		"CopyrightYears": CopyrightYears,
		"UrlEscapeLower": UrlEscapeLower,
	})
	template.Must(listing.ParseFS(listingTemplates, "*.gohtml"))
	log.Printf("listing: %s", listing.DefinedTemplates())
}

type (
	ListingData struct {
		Site     Site
		UrlPath  string
		Title    StringRenderable
		Abstract []Renderable
		Listing  PostListing
	}
)

func WriteListing(w io.Writer, d ListingData) error {
	d.Site = SiteInfo // @todo
	sort.Slice(d.Listing, func(i, j int) bool {
		p1 := d.Listing[i].Published.Published
		p2 := d.Listing[j].Published.Published
		return p1.Compare(p2) < 0 // chronological listing
	})
	return listing.Execute(w, "listing.gohtml", d)
}

func (l ListingData) Canonical() string {
	return fmt.Sprintf("%s://%s/%s", l.Site.Address.Scheme, l.Site.Address.Host, l.UrlPath)
}

func (l ListingData) ObfuscatedAuthorCredit() (template.HTML, error) {
	authorName, err := l.Site.Owner.Render()
	if err != nil {
		return "", err
	}
	return template.HTML(fmt.Sprintf(`<a href="mailto:%s">%s</a>`, ObfuscateText(l.Site.Email), authorName)), nil
}
