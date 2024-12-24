package gen

import (
	"github.com/cvanloo/blog-go/markup/parser"
)

type (
	MakeGenVisitor struct {
		parser.NopVisitor
		TemplateData *Blog
		Errors       error
	}
)

func (v MakeGenVisitor) VisitBlog(b *parser.Blog) {
	if urlPath, ok := b.Meta["url-path"]; ok {
		v.TemplateData.UrlPath = urlPath
	} else {
		v.Errors = errors.Join(v.Errors, errors.New("missing mandatory meta key: url-path"))
	}

	if author, ok := b.Meta["author"]; ok {
		v.TemplateData.Author.Name = author
	} else {
		v.Errors = errors.Join(v.Errors, errors.New("missing mandatory meta key: author"))
	}

	if title, ok := b.Meta["title"]; ok {
		v.TemplateData.Title = title
	} else {
		v.Errors = errors.Join(v.Errors, errors.New("missing mandatory meta key: title"))
	}

	if lang, ok := b.Meta["lang"]; ok {
		v.TemplateData.Lang = lang
	} else {
		v.Errors = errors.Join(v.Errors, errors.New("missing mandatory meta key: lang"))
	}

	if email, ok := b.Meta["email"]; ok {
		v.TemplateData.Author.Email = email
	}
	if relMe, ok := b.Meta["rel-me"]; ok {
		v.TemplateData.Author.RelMe = relMe
	}
	if fediCreator, ok := b.Meta["fedi-creator"]; ok {
		v.TemplateData.Author.FediCreator = fediCreator
	}
	//if template, ok := b.Meta["template"]; ok {
	//	// @todo (this probably has to be handled before / outside the visitor.
	//	// What visitor to use is decided on the template, since the visitor
	//	// has to construct a data structure specific to the template being
	//	// used.
	//}
	if description, ok := b.Meta["description"]; ok {
		v.TemplateData.Description = description
	}
	if altTitle, ok := b.Meta["alt-title"]; ok {
		v.TemplateData.AltTitle = altTitle
	}
	if published, ok := b.Meta["published"]; ok {
		v.TemplateData.Published.Published = published
	}
	if revised, ok := b.Meta["revised"]; ok {
		v.TemplateData.Published.Revised = published
	}
	if estReading, ok := b.Meta["est-reading"]; ok {
		v.TemplateData.EstReading = estReading
	}
	if series, ok := b.Meta["series"]; ok {
		// @todo
	}
	if enableRevisionWarning, ok := b.Meta["enable-revision-warning"]; ok {
		v.TemplateData.EnableRevisionWarning = enableRevisionWarning
	}
	if tags, ok := b.Meta["tags"]; ok {
		for _, tag := range strings.Split(tags, " ") {
			v.TemplateData.Tags = append(v.TemplateData.Tags, Tag(tag))
		}
	}
}
