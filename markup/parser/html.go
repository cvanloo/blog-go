package parser

import (
	"log"
	"fmt"
	"errors"

	"github.com/cvanloo/blog-go/markup/gen"
)

var RegisteredHtmlTags = map[string]HtmlHandler{
	"Todo": MH(todoHtmlTagHandler),
	"Abstract": MH(abstractHtmlTagHandler),
	"Code": CH(codeHtmlTagHandler),
	"Sidenote": TH(sidenoteHtmlTagHandler),
}

type (
	MetaHandler func(*gen.Blog, HtmlTag) error
	ContentHandler func(*gen.Blog, HtmlTag) (gen.Renderable, error)
	TextHandler func(*gen.Blog, HtmlTag) (gen.StringRenderable, error)
	HtmlHandler struct {
		Type HtmlHandlerType
		MH MetaHandler
		CH ContentHandler
		TH TextHandler
	}
	HtmlTag struct {
		Name string
		Args map[string]string
		Strings []string
		Text []gen.StringRenderable
		Children []HtmlTag
	}
)

//go:generate stringer -type HtmlHandlerType -trimprefix HtmlType
type HtmlHandlerType int

const (
	HtmlTypeMeta HtmlHandlerType = iota
	HtmlTypeContent
	HtmlTypeText
)

func MH(h MetaHandler) HtmlHandler {
	return HtmlHandler{
		Type: HtmlTypeMeta,
		MH: h,
	}
}

func CH(h ContentHandler) HtmlHandler {
	return HtmlHandler{
		Type: HtmlTypeContent,
		CH: h,
	}
}

func TH(h TextHandler) HtmlHandler {
	return HtmlHandler{
		Type: HtmlTypeText,
		TH: h,
	}
}

func evaluateHtmlTag(blog *gen.Blog, htmlTag HtmlTag) (gen.Renderable, error) {
	handler, hasHandler := RegisteredHtmlTags[htmlTag.Name]
	if hasHandler {
		switch handler.Type {
		default:
			panic("invalid handler type")
		case HtmlTypeMeta:
			merr := handler.MH(blog, htmlTag)
			return nil, merr
		case HtmlTypeContent:
			return handler.CH(blog, htmlTag)
		case HtmlTypeText:
			return handler.TH(blog, htmlTag)
		}
	}
	return defaultHtmlTagHandler(blog, htmlTag)
}

func defaultHtmlTagHandler(blog *gen.Blog, htmlTag HtmlTag) (gen.Renderable, error) {
	return nil, fmt.Errorf("no handler registered for html tag: %s", htmlTag.Name)
}

func todoHtmlTagHandler(blog *gen.Blog, htmlTag HtmlTag) error {
	log.Println(htmlTag.Strings)
	return nil
}

func abstractHtmlTagHandler(blog *gen.Blog, htmlTag HtmlTag) error {
	blog.Abstract = gen.StringOnlyContent(htmlTag.Text)
	return nil
}

func codeHtmlTagHandler(blog *gen.Blog, htmlTag HtmlTag) (gen.Renderable, error) {
	return gen.CodeBlock{
		Lines: htmlTag.Strings,
	}, nil
}

func sidenoteHtmlTagHandler(blog *gen.Blog, htmlTag HtmlTag) (gen.StringRenderable, error) {
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
