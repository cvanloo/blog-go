package parser

import (
	"log"

	"github.com/cvanloo/blog-go/markup/gen"
)

var RegisteredHtmlTags = map[string]HtmlTagHandler{
	"Abstract": abstractHtmlTagHandler,
	"Code": codeHtmlTagHandler,
}

type (
	HtmlTagHandler func(*gen.Blog, HtmlTag) (gen.Renderable, error)
	HtmlTag struct {
		Name string
		Args map[string]string
		Strings []string
		Text []gen.StringRenderable
	}
)

func evaluateHtmlTag(blog *gen.Blog, htmlTag HtmlTag) (gen.Renderable, error) {
	handler, hasHandler := RegisteredHtmlTags[htmlTag.Name]
	if hasHandler {
		return handler(blog, htmlTag)
	}
	return defaultHtmlTagHandler(blog, htmlTag)
}

func defaultHtmlTagHandler(blog *gen.Blog, htmlTag HtmlTag) (gen.Renderable, error) {
	log.Printf("no handler registered for html tag: %s", htmlTag.Name)
	// @todo: output verbatim?
	return nil, nil
}

func abstractHtmlTagHandler(blog *gen.Blog, htmlTag HtmlTag) (gen.Renderable, error) {
	blog.Abstract = gen.StringOnlyContent(htmlTag.Text)
	return nil, nil
}

func codeHtmlTagHandler(blog *gen.Blog, htmlTag HtmlTag) (gen.Renderable, error) {
	return gen.CodeBlock{
		Lines: htmlTag.Strings,
	}, nil
}
