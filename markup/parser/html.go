package parser

import (
	"log"
	"fmt"
	"errors"

	"github.com/cvanloo/blog-go/markup/gen"
)

var RegisteredHtmlTags = map[string]HtmlTagHandler{
	"Abstract": abstractHtmlTagHandler,
	"Code": codeHtmlTagHandler,
	"Sidenote": sidenoteHtmlTagHandler,
}

type (
	HtmlTagHandler func(*gen.Blog, HtmlTag) (gen.Renderable, error)
	HtmlTag struct {
		Name string
		Args map[string]string
		Strings []string
		Text []gen.StringRenderable
		Children []HtmlTag
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

func sidenoteHtmlTagHandler(blog *gen.Blog, htmlTag HtmlTag) (gen.Renderable, error) {
	var word, content gen.StringRenderable
	wordStr, hasWordArg := htmlTag.Args["Word"]
	word = gen.StringOnlyContent{gen.Text(wordStr)}
	content = gen.StringOnlyContent(htmlTag.Text)
	if !hasWordArg {
		merr := errors.New("sidenote needs two children <Word> and <Content> (in this order)")
		if len(htmlTag.Children) != 2 {
			return nil, merr
		}
		c0 := htmlTag.Children[0]
		if c0.Name != "Word" {
			return nil, merr
		}
		c1 := htmlTag.Children[1]
		if c1.Name != "Content" {
			return nil, merr
		}
		word = gen.StringOnlyContent(c0.Text)
		content = gen.StringOnlyContent(c1.Text)
	}
	return gen.Sidenote{
		ID: fmt.Sprintf("%d", gen.NextID()),
		Word: word,
		Content: content,
	}, nil
}
