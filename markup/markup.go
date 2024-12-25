package markup

import (
	"io"
	"errors"
	"os"
	"fmt"
	"path/filepath"
	"io/fs"
	"log"

	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/parser"
	"github.com/cvanloo/blog-go/page"
	//. "github.com/cvanloo/blog-go/assert"
)

type (
	Markup struct{
		SourcePaths []string
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

func SourcePaths(paths []string) MarkupOption {
	return func(m *Markup) {
		m.SourcePaths = paths
	}
}

func OutDir(path string) MarkupOption {
	return func(m *Markup) {
		m.OutDir = path
	}
}

func (m Markup) Run() (runErr error) {
	p := processor{
		lex: lexer.New(),
		outDir: m.OutDir,
	}
	for _, src := range m.Sources {
		runErr = errors.Join(runErr, p.process(src))
	}
	for _, path := range m.SourcePaths {
		fi, statErr := os.Stat(path)
		if statErr != nil {
			runErr = errors.Join(runErr, statErr)
			continue
		}
		if fi.IsDir() {
			filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					runErr = errors.Join(runErr, err)
					return nil
				}
				if !d.IsDir() {
					if !(filepath.Ext(path) == ".md" || filepath.Ext(path) == ".ᗢ") { // @todo: make configurable
						return nil
					}
					fd, err := os.Open(path)
					if err != nil {
						runErr = errors.Join(runErr, err)
					} else {
						runErr = errors.Join(runErr, p.process(source{
							Name: path,
							In: fd,
						}))
						runErr = errors.Join(runErr, fd.Close())
					}
				}
				return nil
			})
		} else {
			fd, openErr := os.Open(path)
			if openErr != nil {
				runErr = errors.Join(runErr, openErr)
			} else {
				runErr = errors.Join(runErr, p.process(source{
					Name: path,
					In: fd,
				}))
				runErr = errors.Join(runErr, fd.Close())
			}
		}
	}
	return runErr
}

type processor struct {
	lex *lexer.Lexer
	outDir string
}

func (p processor) process(src source) error {
	log.Printf("processing: %s", src.Name)
	p.lex.Clear()
	bs, err := io.ReadAll(src.In)
	if err != nil {
		return fmt.Errorf("processing %s failed while reading: %w", src.Name, err)
	}
	p.lex.LexSource(src.Name, string(bs))
	if len(p.lex.Errors) > 0 {
		return fmt.Errorf("processing %s failed while lexing: %w", src.Name, errors.Join(p.lex.Errors...))
	}
	blog, err := parser.Parse(p.lex)
	if err != nil {
		return fmt.Errorf("processing %s failed while parsing: %w", src.Name, err)
	}
	refFixer := &parser.FixReferencesVisitor{}
	blog.Accept(refFixer)
	if refFixer.Errors != nil {
		return fmt.Errorf("processing %s failed while resolving references: %w", src.Name, refFixer.Errors)
	}
	templateData := page.Post{}
	makeGen := &page.MakeGenVisitor{
		TemplateData: &templateData,
	}
	blog.Accept(makeGen)
	if makeGen.Errors != nil {
		return fmt.Errorf("processing %s failed while producing template data: %w", src.Name, makeGen.Errors)
	}
	out, err := os.Create(fmt.Sprintf("%s/%s.html", p.outDir, templateData.UrlPath))
	if err != nil {
		return fmt.Errorf("processing %s failed while generating static html: %w", src.Name, err)
	}
	if err := page.WritePost(out, templateData); err != nil {
		return fmt.Errorf("processing %s failed while writing out static html: %w", src.Name, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("processing %s failed while closing output file: %w", src.Name, err)
	}
	return nil
}
