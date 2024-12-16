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

func OutDir(path string) MarkupOption {
	return func(m *Markup) {
		m.OutDir = path
	}
}

func (m Markup) Run() (err error) {
	{
		_, err := os.Stat(m.OutDir)
		if err != nil {
			return err
		}
	}
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
		Assert(len(blog.UrlPath) > 0, "urlpath must be set")
		out, cerr := os.Create(fmt.Sprintf("%s/%s.html", m.OutDir, blog.UrlPath))
		if cerr != nil {
			err = errors.Join(err, cerr)
			continue
		}
		werr := gen.WriteBlog(out, &blog)
		err = errors.Join(err, werr)
		err = errors.Join(err, out.Close())
	}
	return err
}
