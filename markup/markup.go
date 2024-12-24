package markup

import (
	"io"
	"errors"
	"os"
	"fmt"

	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/parser"
	"github.com/cvanloo/blog-go/markup/gen"
	. "github.com/cvanloo/blog-go/assert"
)

type (
	Markup struct{
		Sources []source
		StaticSources []string
		OutDir string
	}
	MarkupOption func(*Markup)
	source struct {
		Name string
		In io.Reader
	}
)

func New(opts ...MarkupOption) (m Markup) {
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

func Source(name string, in io.Reader) MarkupOption {
	return func(m *Markup) {
		m.Sources = append(m.Sources, source{
			Name: name,
			In: in,
		})
	}
}

func FileSources(names []string, fds []*os.File) MarkupOption {
	Assert(len(names) == len(fds), "")
	return func(m *Markup) {
		for i := range names {
			m.Sources = append(m.Sources, source{Name: names[i], In: fds[i]})
		}
	}
}

func OutDir(path string) MarkupOption {
	return func(m *Markup) {
		m.OutDir = path
	}
}

func (m Markup) Run() (err error) {
	for _, src := range m.Sources {
		lex := lexer.New()
		bs, rerr := io.ReadAll(src.In)
		if rerr != nil {
			err = errors.Join(err, rerr)
			continue
		}
		lex.LexSource(src.Name, string(bs))
		if len(lex.Errors) > 0 {
			err = errors.Join(err, errors.Join(lex.Errors...))
			continue
		}
		blog, perr := parser.Parse(lex)
		if perr != nil {
			err = errors.Join(err, perr)
			continue
		}
		refFixer := &parser.FixReferencesVisitor{}
		blog.Accept(refFixer)
		if refFixer.Errors != nil {
			err = errors.Join(refFixer.Errors)
			continue
		}
		templateData := gen.Blog{}
		makeGen := &gen.MakeGenVisitor{
			TemplateData: &templateData,
		}
		blog.Accept(makeGen)
		if makeGen.Errors != nil {
			err = errors.Join(err, makeGen.Errors)
			continue
		}
		out, cerr := os.Create(fmt.Sprintf("%s/%s.html", m.OutDir, templateData.UrlPath))
		if cerr != nil {
			err = errors.Join(err, cerr)
			continue
		}
		werr := gen.WriteBlog(out, &templateData)
		err = errors.Join(err, werr)
		err = errors.Join(err, out.Close())
	}
	return err
}
