package markup

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	readingtime "github.com/begmaroman/reading-time"
	"github.com/gorilla/feeds"

	"github.com/cvanloo/blog-go/markup/lexer"
	"github.com/cvanloo/blog-go/markup/parser"
	"github.com/cvanloo/blog-go/page"
	//. "github.com/cvanloo/blog-go/assert"
)

type (
	Markup struct {
		SiteInfo      page.Site
		IncludeExt    []string
		ExcludeExt    []string
		SourcePaths   []string
		Sources       []source
		StaticSources []string
		OutDir        string
	}
	MarkupOption func(*Markup)
	source       struct {
		Name string
		In   io.Reader
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
			In:   in,
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

	// @todo: only continue processing error free sources
	//if runErr != nil {
	//	return runErr
	//}
	// @todo: make sure url paths don't collide!

	tp := newTemplatePreProcessor(mp.results)
	runErr = errors.Join(runErr, tp.Run())

	ap := newAssetsProcessor(m.OutDir, mp.results)
	runErr = errors.Join(runErr, ap.Run())

	gp := newTemplateGenProcessor(m.OutDir, tp)
	runErr = errors.Join(runErr, gp.Run())

	fp := newFeedProcessor(m.SiteInfo, m.OutDir, gp.posts)
	runErr = errors.Join(runErr, fp.Run())

	return runErr
}

type (
	ProcessingStep interface {
		Run() error
	}

	markupProcessor struct {
		c           chan markupResult
		includeExt  []string
		excludeExt  []string
		sourcePaths []string
		sources     []source
		results     []markupResult
		err         error
	}
	markupResult struct {
		src source
		err error
		lex *lexer.Lexer
		par *parser.Blog
		est *readingtime.Result
	}

	templatePreProcessor struct {
		markups []markupResult
		tags    map[string]page.ListingData
		series  map[string]page.ListingData
		posts   map[string]*page.Post
		quotes  map[string]*page.Post
		index   page.IndexData
	}

	templateGenProcessor struct {
		outDir string
		tags   []page.ListingData
		series []page.ListingData
		posts  []page.Post
		quotes []page.Post
		index  page.IndexData
	}

	feedProcessor struct {
		siteInfo page.Site
		outDir string
		posts  []page.Post
	}

	assetsProcessor struct {
		outDir string
		markups []markupResult
	}
)

func newMarkupProcessor(includeExt []string, excludeExt []string, sourcePaths []string, sources []source) markupProcessor {
	return markupProcessor{
		c:           make(chan markupResult, 16),
		includeExt:  includeExt,
		excludeExt:  excludeExt,
		sourcePaths: sourcePaths,
		sources:     sources,
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
								In:   fd,
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
						In:   fd,
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
	lex, par, est, err := lexAndParse(src)
	p.c <- markupResult{
		src: src,
		err: err,
		lex: lex,
		par: par,
		est: est,
	}
	wg.Done()
}

func lexAndParse(src source) (*lexer.Lexer, *parser.Blog, *readingtime.Result, error) {
	lex := lexer.New()
	bs, err := io.ReadAll(src.In)
	est := readingtime.Estimate(string(bs))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("processing %s failed while reading: %w", src.Name, err)
	}
	lex.LexSource(src.Name, string(bs))
	if len(lex.Errors) > 0 {
		return lex, nil, est, fmt.Errorf("processing %s failed while lexing: %w", src.Name, errors.Join(lex.Errors...))
	}
	blog, err := parser.Parse(lex)
	if err != nil {
		return lex, blog, est, fmt.Errorf("processing %s failed while parsing: %w", src.Name, err)
	}
	refFixer := &parser.FixReferencesVisitor{}
	blog.Accept(refFixer)
	if refFixer.Errors != nil {
		return lex, blog, est, fmt.Errorf("processing %s failed while resolving references: %w", src.Name, refFixer.Errors)
	}
	if rc, ok := src.In.(io.ReadCloser); ok {
		if err := rc.Close(); err != nil {
			return lex, blog, est, err
		}
	}
	return lex, blog, est, nil
}

func newTemplatePreProcessor(markups []markupResult) templatePreProcessor {
	return templatePreProcessor{
		markups: markups,
		series:  map[string]page.ListingData{},
		tags:    map[string]page.ListingData{},
		posts:   map[string]*page.Post{},
		quotes:  map[string]*page.Post{},
		index:   page.IndexData{},
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
			case "post-quotes":
				runErr = errors.Join(runErr, p.processQuotes(m))
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
	templateData.EstReading = int(m.est.Duration.Minutes())
	p.posts[templateData.UrlPath] = &templateData
	if templateData.MakePublish {
		p.index.Listing = append(p.index.Listing, page.PostItem{
			Title:       templateData.Title,
			AltTitle:    templateData.AltTitle,
			UrlPath:     templateData.UrlPath,
			Tags:        templateData.Tags,
			Description: templateData.Description,
			Abstract:    templateData.Abstract,
			EstReading:  templateData.EstReading,
			Published:   templateData.Published,
		})
		for _, tag := range templateData.Tags {
			ti := p.tags[string(tag)]
			ti.Title = page.StringOnlyContent{
				page.Text("Posts tagged with :"),
				page.Text(tag),
			}
			ti.UrlPath = fmt.Sprintf(":%s", tag)
			ti.Listing = append(ti.Listing, page.PostItem{
				Title:       templateData.Title,
				AltTitle:    templateData.AltTitle,
				UrlPath:     templateData.UrlPath,
				Tags:        templateData.Tags,
				Description: templateData.Description,
				Abstract:    templateData.Abstract,
				EstReading:  templateData.EstReading,
				Published:   templateData.Published,
			})
			p.tags[string(tag)] = ti
		}
		if templateData.Series != nil {
			seriesName := templateData.Series.Name
			si := p.series[seriesName.Text()]
			si.Title = seriesName
			si.UrlPath = seriesName.Text()
			si.Listing = append(si.Listing, page.PostItem{
				Title:       templateData.Title,
				AltTitle:    templateData.AltTitle,
				UrlPath:     templateData.UrlPath,
				Tags:        templateData.Tags,
				Description: templateData.Description,
				Abstract:    templateData.Abstract,
				EstReading:  templateData.EstReading,
				Published:   templateData.Published,
			})
			p.series[seriesName.Text()] = si
		}
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
					Link:  prev.UrlPath,
				}
			}
			if i < len(seriesListing.Listing)-1 { // has next item
				next := seriesListing.Listing[i+1]
				post.Series.Next = &page.SeriesItem{
					Title: next.Title,
					Link:  next.UrlPath,
				}
			}
		}
	}
	return nil
}

func (p *templatePreProcessor) processQuotes(m markupResult) error {
	templateData := page.Post{}
	makeGen := &page.MakeQuotesVisitor{
		page.MakeGenVisitor{
			TemplateData: &templateData,
		},
	}
	m.par.Accept(makeGen)
	if makeGen.Errors != nil {
		return fmt.Errorf("processing %s failed while producing template data: %w", m.src.Name, makeGen.Errors)
	}
	templateData.EstReading = int(m.est.Duration.Minutes())
	p.quotes[templateData.UrlPath] = &templateData
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
	var quotes []page.Post
	for _, quote := range t.quotes {
		quotes = append(quotes, *quote)
	}
	return templateGenProcessor{
		outDir: outDir,
		tags:   tags,
		series: series,
		posts:  posts,
		quotes: quotes,
		index:  t.index,
	}
}

func (p templateGenProcessor) Run() (runErr error) {
	for _, post := range p.posts {
		if post.MakePublish || os.Getenv("TESTING") == "1" {
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
		} else {
			// @todo: maybe still generate the html, but save it into a draft/ folder?
			runErr = errors.Join(runErr, fmt.Errorf("not publishing post, because meta key draft != false: %s", post.UrlPath))
		}
	}
	for _, quote := range p.quotes {
		if quote.MakePublish {
			out, err := os.Create(filepath.Join(p.outDir, fmt.Sprintf("%s.html", quote.UrlPath))) // @todo: make UrlPath custom type
			if err != nil {
				runErr = errors.Join(runErr, err)
				continue
			}
			if err := page.WritePost(out, quote); err != nil {
				runErr = errors.Join(runErr, err)
				continue
			}
			runErr = errors.Join(runErr, out.Close())
		} else {
			// @todo: maybe still generate the html, but save it into a draft/ folder?
			runErr = errors.Join(runErr, fmt.Errorf("not publishing post, because meta key draft != false: %s", quote.UrlPath))
		}
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

func newFeedProcessor(siteInfo page.Site, outDir string, posts []page.Post) feedProcessor {
	return feedProcessor{
		siteInfo: siteInfo,
		outDir: outDir,
		posts:  posts,
	}
}

func (p feedProcessor) Run() (runErr error) {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       p.siteInfo.Name,
		Link:        &feeds.Link{Href: p.siteInfo.Address.String()},
		Description: p.siteInfo.DefaultTagline.Text(), // @todo: DefaultTagline.TextOnly()
		Author:      &feeds.Author{Name: p.siteInfo.Owner.Text()}, // @todo: Owner.TextOnly()
		Created:     now,
	}

	for _, post := range p.posts {
		if !post.MakePublish {
			continue
		}
		title := post.Title.Text()
		desc := post.Description
		author := post.Author.Name.Text()
		published := post.Published.Published
		var revised time.Time
		if post.Published.Revised != nil {
			revised = *post.Published.Revised
		}
		feed.Items = append(feed.Items, &feeds.Item{
			Id:          post.Canonical(),
			Title:       title,
			Link:        &feeds.Link{Href: post.Canonical()},
			Description: desc,
			Author:      &feeds.Author{Name: author},
			Created:     published,
			Updated:     revised,
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

func newAssetsProcessor(outDir string, markups []markupResult) assetsProcessor {
	return assetsProcessor{
		outDir: outDir,
		markups: markups,
	}
}

type AssetFinderVisitor struct {
	parser.NopVisitor
	Images []string
	Videos []string
	Err error
}

func (v *AssetFinderVisitor) VisitImage(i *parser.Image) {
	ext := filepath.Ext(i.Name)
	switch ext {
	default:
		v.Err = errors.Join(v.Err, fmt.Errorf("unrecognized file extension: %s", ext))
	case ".jpg", ".jpeg", ".jxl", ".avif", ".webp", ".png":
		v.Images = append(v.Images, i.Name)
	case ".mp4", ".mkv", ".webm":
		v.Videos = append(v.Videos, i.Name)
	}
}

func (p assetsProcessor) Run() (runErr error) {
	if os.Getenv("MAKE_ASSETS") != "1" {
		log.Println("not generating assets, because MAKE_ASSETS=1 is not set")
		return p.verifyAssets()
	}
	return p.generateAssets()
}

var ExtensionsImage = []string{".jxl", ".avif", ".jpg"}
var ExtensionsVideo = []string{".mp4", ".webm"}

func (p assetsProcessor) verifyAssets() (runErr error) {
	v := &AssetFinderVisitor{}
	for _, markup := range p.markups {
		post := markup.par
		post.Accept(v)
		runErr = errors.Join(runErr, v.Err)
		v.Err = nil
	}
	if len(v.Images) + len(v.Videos) > 0 {
		outDir := filepath.Join(p.outDir, "/assets/")
		fi, err := os.Stat(outDir)
		if err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("asset directory does not exist: %w", err))
		} else if !fi.IsDir() {
			runErr = errors.Join(runErr, errors.New("asset directory is unexpectedly a file"))
		} else {
			neededAssets := map[string]bool{}
			for _, asset := range v.Images {
				for _, ext := range ExtensionsImage {
					targetBase := strings.ReplaceAll(filepath.Base(asset), filepath.Ext(asset), ext)
					dst := filepath.Join(p.outDir, "/assets/", targetBase)
					neededAssets[dst] = false
				}
			}
			for _, asset := range v.Videos {
				for _, ext := range ExtensionsVideo {
					targetBase := strings.ReplaceAll(filepath.Base(asset), filepath.Ext(asset), ext)
					dst := filepath.Join(p.outDir, "/assets/", targetBase)
					neededAssets[dst] = false
				}
			}
			filepath.WalkDir(outDir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					runErr = errors.Join(runErr, err)
					return nil
				}
				neededAssets[filepath.Clean(path)] = true
				return nil
			})
			for assetPath, foundInAssetDir := range neededAssets {
				if !foundInAssetDir {
					log.Printf("missing asset: %s", assetPath)
				}
			}
		}
	}
	return runErr
}

func (p assetsProcessor) generateAssets() (runErr error) {
	v := &AssetFinderVisitor{}
	converter, err := convertUtility()
	if err != nil {
		return err
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return err
	}
	//if _, err := exec.LookPath("exiftool"); err != nil {
	//	return err
	//}
	outDir := filepath.Join(p.outDir, "/assets/")
	if err := os.Mkdir(outDir, 0777); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	for _, markup := range p.markups {
		// @todo: can do this in parallel
		post := markup.par
		post.Accept(v)
		for _, asset := range v.Images {
			src := filepath.Join(filepath.Dir(markup.src.Name), asset)
			fi, err := os.Stat(src)
			if err != nil {
				runErr = errors.Join(runErr, fmt.Errorf("%s: missing asset: %v", markup.src.Name, err))
				continue
			}
			if fi.IsDir() {
				runErr = errors.Join(runErr, fmt.Errorf("%s: asset is a directory, expected an image: %v", markup.src.Name, err))
				continue
			}
			extensions := ExtensionsImage
			switch converter {
			case "magick":
				for _, ext := range extensions {
					targetBase := strings.ReplaceAll(filepath.Base(asset), filepath.Ext(asset), ext)
					dst := filepath.Join(p.outDir, "/assets/", targetBase)
					log.Printf("processing asset: %s -> %s", src, dst)
					cmd := exec.Command("magick", src, "-strip", dst)
					out, err := cmd.CombinedOutput()
					if err != nil {
						if err, ok := err.(*exec.ExitError); ok {
							runErr = errors.Join(runErr, fmt.Errorf("magick exited with status: %d", err.ExitCode()))
						} else {
							runErr = errors.Join(runErr, err)
						}
					}
					log.Println(string(out))
				}
			case "convert":
				for _, ext := range extensions {
					targetBase := strings.ReplaceAll(filepath.Base(asset), filepath.Ext(asset), ext)
					dst := filepath.Join(p.outDir, "/assets/", targetBase)
					log.Printf("processing asset: %s -> %s", src, dst)
					cmd := exec.Command("convert", src, "-strip", dst)
					out, err := cmd.CombinedOutput()
					if err != nil {
						if err, ok := err.(*exec.ExitError); ok {
							runErr = errors.Join(runErr, fmt.Errorf("convert exited with status: %d", err.ExitCode()))
						} else {
							runErr = errors.Join(runErr, err)
						}
					}
					log.Println(string(out))
				}
			case "ffmpeg":
				for _, ext := range extensions {
					targetBase := strings.ReplaceAll(filepath.Base(asset), filepath.Ext(asset), ext)
					dst := filepath.Join(p.outDir, "/assets/", targetBase)
					log.Printf("processing asset: %s -> %s", src, dst)
					cmd := exec.Command("ffmpeg", "-i", src, "-map_metadata", "-1", dst)
					out, err := cmd.CombinedOutput()
					if err != nil {
						if err, ok := err.(*exec.ExitError); ok {
							runErr = errors.Join(runErr, fmt.Errorf("ffmpeg exited with status: %d", err.ExitCode()))
						} else {
							runErr = errors.Join(runErr, err)
						}
					}
					log.Println(string(out))
				}
			default:
				panic("unreachable, unless programmer fucked up")
			}
		}
		for _, asset := range v.Videos {
			extensions := ExtensionsVideo
			src := filepath.Join(filepath.Dir(markup.src.Name), asset)
			fi, err := os.Stat(src)
			if err != nil {
				runErr = errors.Join(runErr, fmt.Errorf("%s: missing asset: %v", markup.src.Name, err))
				continue
			}
			if fi.IsDir() {
				runErr = errors.Join(runErr, fmt.Errorf("%s: asset is a directory, expected a video: %v", markup.src.Name, err))
				continue
			}
			for _, ext := range extensions {
				targetBase := strings.ReplaceAll(filepath.Base(asset), filepath.Ext(asset), ext)
				dst := filepath.Join(p.outDir, "/assets/", targetBase)
				log.Printf("processing asset: %s -> %s", src, dst)
				cmd := exec.Command("ffmpeg", "-i", src, dst)
				if err := cmd.Start(); err != nil {
					runErr = errors.Join(runErr, err)
				} else if err := cmd.Wait(); err != nil {
					if err, ok := err.(*exec.ExitError); ok {
						runErr = errors.Join(runErr, fmt.Errorf("ffmpeg exited with status: %d", err.ExitCode()))
					} else {
						runErr = errors.Join(runErr, err)
					}
				}
			}
		}
		runErr = errors.Join(runErr, v.Err)
		v.Images = nil
		v.Videos = nil
		v.Err = nil
	}
	return runErr
}

func convertUtility() (string, error) {
	if _, err := exec.LookPath("magick"); err != nil {
		if _, err := exec.LookPath("convert"); err != nil {
			if _, err := exec.LookPath("ffmpeg"); err != nil {
				return "", errors.New("cannot convert images, since neither magick, convert, nor ffmpeg could be found in $PATH")
			}
			return "ffmpeg", nil
		}
		return "convert", nil
	}
	return "magick", nil
}
