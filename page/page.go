package page

import (
	"io"
	"fmt"
	"time"
	"sync"
	"html/template"
	"net/url"
)

var SiteInfo Site

type (
	Site struct {
		Address *url.URL // https://blog.vanloo.ch/
		Name string // save-lisp-and-die
		DefaultTagline StringRenderable // A blog about programming <weird> computers using <weird> languages.
		RelMe string // https://tech.lgbt/@attaboy
		FediCreator string // @attaboy@tech.lgbt
		Owner StringRenderable // Colin van~Loo
		Email string
		Birthday time.Time // 2024
	}
	Revision struct {
		Published time.Time
		Revised *time.Time
	}
	PostListing []PostItem
	PostItem struct {
		Title, AltTitle StringRenderable
		UrlPath string
		Tags []Tag
		Description string
		Abstract []Renderable
		EstReading int
		Published Revision
	}
	Tag string
)

type (
	Template struct {
		*template.Template
	}
	Renderable interface {
		Render() (template.HTML, error)
	}
	StringRenderable interface {
		Renderable
		Text() string
	}
	StringSanitizedRenderable interface {
		StringRenderable
		SanitizedText() string
	}
	Identifiable interface {
		ID() string
	}
)

func Render(element Renderable) (template.HTML, error) {
	return element.Render()
}

// @todo: unique per page that we're generating
func MakeUniqueID(element any) (string, error) {
	mu.Lock()
	defer mu.Unlock()
	switch i := element.(type) {
	default:
		for {
			id := fmt.Sprintf("%d", NextID())
			if _, idExists := seenIDs[id]; !idExists {
				seenIDs[id] = struct{}{}
				return id, nil
			}
		}
	case Identifiable:
		id := i.ID()
		if _, alreadySeen := seenIDs[id]; alreadySeen {
			return id, fmt.Errorf("duplicate id: %s", id)
		}
		seenIDs[id] = struct{}{}
		return id, nil
	}
}

var (
	mu sync.Mutex
	currentID int
	seenIDs = map[string]struct{}{}
)

func NextID() int {
	currentID++
	return currentID
}
// @todo: end

func (t *Template) Execute(w io.Writer, name string, data any) error {
	return t.Template.ExecuteTemplate(w, name, data)
}

func (s Site) CanonicalAddress() string {
	return fmt.Sprintf("%s://%s/", s.Address.Scheme, s.Address.Host)
}

func (r Revision) HasRevision() bool {
	return r.Revised != nil
}

func (r Revision) String() string {
	const timeFormat = "Mon, 2 Jan 2006"
	if r.Revised != nil {
		return fmt.Sprintf("%s (revised %s)", r.Published.Format(timeFormat), r.Revised.Format(timeFormat))
	}
	return fmt.Sprintf("%s", r.Published.Format(timeFormat))
}

func ObfuscateText(s string) template.HTML {
	const (
		janetStart = `janet -e '(print (string/from-bytes (splice (map (fn [c] (if (<= 97 c 122) (+ 97 (mod (+ c -97 13) 26)) c)) &quot;`
		janetEnd = `&quot;))))'`
	)
	rot13 := func(text string) string {
		out := []rune(text)
		for i, r := range out {
			if r >= 'a' && r <= 'z' {
				out[i] = ((r - 'a' + 13) % 26) + 'a'
			}
		}
		return string(out)
	}
	return template.HTML(janetStart + rot13(s) + janetEnd)
}

func CopyrightYear() string {
	return time.Now().Format("2006")
}

func CopyrightYears(start time.Time) template.HTML {
	now := time.Now()
	if now.Year() != start.Year() {
		return template.HTML(fmt.Sprintf("%s&ndash;%s", start.Format("2006"), now.Format("2006")))
	}
	return template.HTML(start.Format("2006"))
}

func (t Tag) String() string {
	return string(t)
}
