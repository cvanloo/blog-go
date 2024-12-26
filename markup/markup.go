package markup

import (
	"io"
	"errors"
	"os"
	"fmt"
	"path/filepath"
	"io/fs"
	"log"
	"sync"
	"time"
	"sort"
	"slices"

	"github.com/gorilla/feeds"

	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/parser"
	"github.com/cvanloo/blog-go/page"
	//. "github.com/cvanloo/blog-go/assert"
)

type (
	Markup struct {
		SiteInfo page.Site
		IncludeExt []string
		ExcludeExt []string
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
	if len(m.IncludeExt) == 0 {
		m.IncludeExt = append(m.IncludeExt, ".md", ".ᗢ")
	}
	return m
}

func SiteInfo(site page.Site) MarkupOption {
	return func(m *Markup) {
		m.SiteInfo = site
	}
}

func IncludeExtensions(ext ...string) MarkupOption {
	return func(m *Markup) {
		m.IncludeExt = append(m.IncludeExt, ext...)
	}
}

func ExcludeExtensions(ext ...string) MarkupOption {
	return func(m *Markup) {
		m.ExcludeExt = append(m.ExcludeExt, ext...)
	}
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
	page.SiteInfo = m.SiteInfo // @todo

	mp := newMarkupProcessor(m.IncludeExt, m.ExcludeExt, m.SourcePaths, m.Sources)
	runErr = errors.Join(runErr, mp.Run())

	// @todo: make sure url paths don't collide!
	
	tp := newTemplatePreProcessor(mp.results)
	runErr = errors.Join(runErr, tp.Run())

	gp := newTemplateGenProcessor(m.OutDir, tp)
	runErr = errors.Join(runErr, gp.Run())

	fp := newFeedProcessor(m.OutDir,gp.posts)
	runErr = errors.Join(runErr, fp.Run())

	return runErr
}

type (
	ProcessingStep interface {
		Run() error
	}

	markupProcessor struct {
		c chan markupResult
		includeExt []string
		excludeExt []string
		sourcePaths []string
		sources []source
		results []markupResult
		err error
	}
	markupResult struct {
		src source
		err error
		lex *lexer.Lexer
		par *parser.Blog
	}

	templatePreProcessor struct {
		markups []markupResult
		tags map[string]page.ListingData
		series map[string]page.ListingData
		posts map[string]*page.Post
		index page.IndexData
	}

	templateGenProcessor struct {
		outDir string
		tags []page.ListingData
		series []page.ListingData
		posts []page.Post
		index page.IndexData
	}

	feedProcessor struct {
		outDir string
		posts []page.Post
	}
)

func newMarkupProcessor(includeExt []string, excludeExt []string, sourcePaths []string, sources []source) markupProcessor {
	return markupProcessor{
		c: make(chan markupResult, 16),
		includeExt: includeExt,
		excludeExt: excludeExt,
		sourcePaths: sourcePaths,
		sources: sources,
	}
}

func (p *markupProcessor) Run() (runErr error) {
	go func() {
		var wg sync.WaitGroup

		wg.Add(len(p.sources))
		for _, src := range p.sources {
			go p.process(src, &wg)
		}

		for _, path := range p.sourcePaths {
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
						if !slices.Contains(p.includeExt, filepath.Ext(path)) {
							return nil
						}
						if slices.Contains(p.excludeExt, filepath.Ext(path)) {
							return nil
						}
						fd, err := os.Open(path)
						if err != nil {
							runErr = errors.Join(runErr, err)
						} else {
							wg.Add(1)
							go p.process(source{
								Name: path,
								In: fd,
							}, &wg)
						}
					}
					return nil
				})
			} else {
				fd, openErr := os.Open(path)
				if openErr != nil {
					runErr = errors.Join(runErr, openErr)
				} else {
					wg.Add(1)
					go p.process(source{
						Name: path,
						In: fd,
					}, &wg)
				}
			}
		}

		wg.Wait()
		close(p.c)
	}()

	p.collect()
	runErr = errors.Join(runErr, p.err)
	return runErr
}

func (p *markupProcessor) collect() {
	for {
		res, ok := <-p.c
		if !ok {
			return
		}
		p.err = errors.Join(p.err, res.err)
		p.results = append(p.results, res)
	}
}

func (p *markupProcessor) process(src source, wg *sync.WaitGroup) {
	log.Printf("processing: %s", src.Name)
	lex, par, err := lexAndParse(src)
	p.c <- markupResult{
		src: src,
		err: err,
		lex: lex,
		par: par,
	}
	wg.Done()
}

func lexAndParse(src source) (*lexer.Lexer, *parser.Blog, error) {
	lex := lexer.New()
	bs, err := io.ReadAll(src.In)
	if err != nil {
		return nil, nil, fmt.Errorf("processing %s failed while reading: %w", src.Name, err)
	}
	lex.LexSource(src.Name, string(bs))
	if len(lex.Errors) > 0 {
		return lex, nil, fmt.Errorf("processing %s failed while lexing: %w", src.Name, errors.Join(lex.Errors...))
	}
	blog, err := parser.Parse(lex)
	if err != nil {
		return lex, blog, fmt.Errorf("processing %s failed while parsing: %w", src.Name, err)
	}
	refFixer := &parser.FixReferencesVisitor{}
	blog.Accept(refFixer)
	if refFixer.Errors != nil {
		return lex, blog, fmt.Errorf("processing %s failed while resolving references: %w", src.Name, refFixer.Errors)
	}
	if rc, ok := src.In.(io.ReadCloser); ok {
		if err := rc.Close(); err != nil {
			return lex, blog, err
		}
	}
	return lex, blog, nil
}

func newTemplatePreProcessor(markups []markupResult) templatePreProcessor {
	return templatePreProcessor{
		markups: markups,
		series: map[string]page.ListingData{},
		tags: map[string]page.ListingData{},
		posts: map[string]*page.Post{},
		index: page.IndexData{},
	}
}

func (p *templatePreProcessor) Run() (runErr error) {
	for _, m := range p.markups {
		template, ok := m.par.Meta.Template()
		if !ok {
			runErr = errors.Join(runErr, fmt.Errorf("missing or invalid template definition for: %s", m.src.Name))
		} else {
			switch template {
			default:
				runErr = errors.Join(runErr, fmt.Errorf("unknown template: %s", template))
			case "post":
				runErr = errors.Join(runErr, p.processPost(m))
			}
		}
	}
	runErr = errors.Join(runErr, p.fixSeriesData())
	return runErr
}

func (p *templatePreProcessor) processPost(m markupResult) error {
	templateData := page.Post{}
	makeGen := &page.MakeGenVisitor{
		TemplateData: &templateData,
	}
	m.par.Accept(makeGen)
	if makeGen.Errors != nil {
		return fmt.Errorf("processing %s failed while producing template data: %w", m.src.Name, makeGen.Errors)
	}
	p.posts[templateData.UrlPath] = &templateData
	p.index.Listing = append(p.index.Listing, page.PostItem{
		Title: templateData.Title,
		AltTitle: templateData.AltTitle,
		UrlPath: templateData.UrlPath,
		Tags: templateData.Tags,
		Abstract: templateData.Abstract,
		EstReading: templateData.EstReading,
		Published: templateData.Published,
	})
	for _, tag := range templateData.Tags {
		ti := p.tags[string(tag)]
		ti.Title = page.Text(tag)
		ti.UrlPath = fmt.Sprintf(":%s", tag)
		ti.Listing = append(ti.Listing, page.PostItem{
			Title: templateData.Title,
			AltTitle: templateData.AltTitle,
			UrlPath: templateData.UrlPath,
			Tags: templateData.Tags,
			Abstract: templateData.Abstract,
			EstReading: templateData.EstReading,
			Published: templateData.Published,
		})
		p.tags[string(tag)] = ti
	}
	if templateData.Series != nil {
		seriesName := templateData.Series.Name
		si := p.series[seriesName.Text()]
		si.Title = seriesName
		si.UrlPath = seriesName.Text()
		si.Listing = append(si.Listing, page.PostItem{
			Title: templateData.Title,
			AltTitle: templateData.AltTitle,
			UrlPath: templateData.UrlPath,
			Tags: templateData.Tags,
			Abstract: templateData.Abstract,
			EstReading: templateData.EstReading,
			Published: templateData.Published,
		})
		p.series[seriesName.Text()] = si
	}
	return nil
}

func (p *templatePreProcessor) fixSeriesData() error {
	for seriesName, seriesListing := range p.series {
		_ = seriesName
		sort.Slice(seriesListing.Listing, func(i, j int) bool {
			p1 := seriesListing.Listing[i].Published.Published
			p2 := seriesListing.Listing[j].Published.Published
			return p1.Compare(p2) < 0
		})
		for i, si := range seriesListing.Listing {
			post := p.posts[si.UrlPath]
			if i > 0 { // has prev item
				prev := seriesListing.Listing[i-1]
				post.Series.Prev = &page.SeriesItem{
					Title: prev.Title,
					Link: prev.UrlPath,
				}
			}
			if i < len(seriesListing.Listing)-1 { // has next item
				next := seriesListing.Listing[i+1]
				post.Series.Next = &page.SeriesItem{
					Title: next.Title,
					Link: next.UrlPath,
				}
			}
		}
	}
	return nil
}

func newTemplateGenProcessor(outDir string, t templatePreProcessor) templateGenProcessor {
	var tags []page.ListingData
	for _, tag := range t.tags {
		tags = append(tags, tag)
	}
	var series []page.ListingData
	for _, s := range t.series {
		series = append(series, s)
	}
	var posts []page.Post
	for _, post := range t.posts {
		posts = append(posts, *post)
	}
	return templateGenProcessor{
		outDir: outDir,
		tags: tags,
		series: series,
		posts: posts,
		index: t.index,
	}
}

func (p templateGenProcessor) Run() (runErr error) {
	for _, post := range p.posts {
		out, err := os.Create(filepath.Join(p.outDir, fmt.Sprintf("%s.html", post.UrlPath))) // @todo: make UrlPath custom type
		if err != nil {
			runErr = errors.Join(runErr, err)
			continue
		}
		if err := page.WritePost(out, post); err != nil {
			runErr = errors.Join(runErr, err)
			continue
		}
		runErr = errors.Join(runErr, out.Close())
	}
	for _, series := range p.series {
		out, err := os.Create(filepath.Join(p.outDir, fmt.Sprintf("%s.html", series.UrlPath)))
		if err != nil {
			runErr = errors.Join(runErr, err)
			continue
		}
		if err := page.WriteListing(out, series); err != nil {
			runErr = errors.Join(runErr, err)
			continue
		}
		runErr = errors.Join(runErr, out.Close())
	}
	for _, tag := range p.tags {
		out, err := os.Create(filepath.Join(p.outDir, fmt.Sprintf("%s.html", tag.UrlPath)))
		if err != nil {
			runErr = errors.Join(runErr, err)
			continue
		}
		if err := page.WriteListing(out, tag); err != nil {
			runErr = errors.Join(runErr, err)
			continue
		}
		runErr = errors.Join(runErr, out.Close())
	}
	{
		out, err := os.Create(filepath.Join(p.outDir, "index.html"))
		if err != nil {
			runErr = errors.Join(runErr, err)
		}
		if err := page.WriteIndex(out, p.index); err != nil {
			runErr = errors.Join(runErr, err)
		}
		runErr = errors.Join(runErr, out.Close())
	}
	return runErr
}

func newFeedProcessor(outDir string, posts []page.Post) feedProcessor {
	return feedProcessor{
		outDir: outDir,
		posts: posts,
	}
}

func (p feedProcessor) Run() (runErr error) {
	now := time.Now()
	feed := &feeds.Feed{
		Title: "",
		Link: &feeds.Link{Href: ""},
		Description: "",
		Author: &feeds.Author{Name: ""},
		Created: now,
	}

	for _, post := range p.posts {
		title := post.Title.Text()
		desc := post.Description
		author := post.Author.Name.Text()
		published := post.Published.Published
		var revised time.Time
		if post.Published.Revised != nil {
			revised = *post.Published.Revised
		}
		feed.Items = append(feed.Items, &feeds.Item{
			Title: title,
			Link: &feeds.Link{Href: post.Canonical()},
			Description: desc,
			Author: &feeds.Author{Name: author},
			Created: published,
			Updated: revised,
		})
	}

	atom, err := feed.ToAtom()
	runErr = errors.Join(runErr, err)
	runErr = errors.Join(runErr, os.WriteFile(filepath.Join(p.outDir, "feed.atom"), []byte(atom), 0666))

	rss, err := feed.ToRss()
	runErr = errors.Join(runErr, err)
	runErr = errors.Join(runErr, os.WriteFile(filepath.Join(p.outDir, "feed.rss"), []byte(rss), 0666))

	json, err := feed.ToJSON()
	runErr = errors.Join(runErr, err)
	runErr = errors.Join(runErr, os.WriteFile(filepath.Join(p.outDir, "feed.json"), []byte(json), 0666))

	return runErr
}
