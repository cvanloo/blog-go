package main

import (
	"net/http"
	"time"
	"fmt"
	"mime"
	//"log"

	. "github.com/cvanloo/blog-go/gen"
)

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("must: %s", err))
	}
	return t
}

func asRef[T any](t T) *T {
	return &t
}

func init() {
	_ = mime.AddExtensionType(".js", "text/javascript")
	_ = mime.AddExtensionType(".css", "text/css")
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/blog", Handler(blog, panicIf))
	/*http.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s NOT FOUND", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))*/
	panicIf(http.ListenAndServe(":8081", nil))
}

var blog = &Blog{
		Author: Author{
			Name: "Colin van Loo",
			Email: "colin@vanloo.ch",
		},
		Lang: "en",
		Title: "Lorem Ipsum",
		AltTitle: "Neque porro quisquam est qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit...",
		Published: Revision{
			Published: must(time.Parse("2006-01-02", "2019-11-15")),
			Revised: asRef(must(time.Parse("2006-01-02", "2020-12-13"))),
		},
		EstReading: 20,
		Tags: []Tag{"meta", "test", "mock"},
		Series: &Series{
			Prev: &SeriesItem{
				Title: "Lorem Epsum",
				Link: "/not-found/lorem-epsum",
			},
			Next: &SeriesItem{
				Title: "Lorem Ipsum The Sequel",
				Link: "/not-found/lorem-sequel",
			},
		},
		EnableRevisionWarning: true,
		TOC: TableOfContents{
			Sections: []TOCSection{
				TOCSection{
					ID: "section-1",
					Heading: "Section 1",
					NextLevel: []TOCSection{
						TOCSection{
							ID: "section-1-1",
							Heading: "Section 1-1",
						},
					},
				},
				TOCSection{
					ID: "section-2",
					Heading: "Section 2",
				},
			},
		},
		Abstract: []Renderable{
			Paragraph{
				Content: []Renderable{
					Text("Lorem ipsum dolor sit amet, consectetur "),
					Bold("adipiscing elit"),
					Text(", sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."),
				},
			},
		},
		Sections: []Section{
			Section{
				Level: 1,
				Heading: "Section 1",
				Content: []Renderable{
					Paragraph{
						Content: []Renderable{
							Text("Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem. Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam, nisi ut aliquid ex ea commodi consequatur? Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur, vel illum qui dolorem eum fugiat quo voluptas nulla pariatur?"),
						},
					},
					Section{
						Level: 2,
						Heading: "Section 1-1",
						Content: []Renderable{
							Paragraph{
								Content: []Renderable{
									Text("But I must explain to you how all this mistaken idea of denouncing pleasure and praising pain was born and I will give you a complete account of the system, and expound the actual teachings of the great explorer of the truth, the master-builder of human happiness. No one rejects, dislikes, or avoids pleasure itself, because it is pleasure, but because those who do not know how to pursue pleasure rationally encounter consequences that are extremely painful. Nor again is there anyone who loves or pursues or desires to obtain pain of itself, because it is pain, but because occasionally circumstances occur in which toil and pain can procure him some great pleasure. To take a trivial example, which of us ever undertakes laborious physical exercise, except to obtain some advantage from it? But who has any right to find fault with a man who chooses to enjoy a pleasure that has no annoying consequences, or one who avoids a pain that produces no resultant pleasure?"),
								},
							},
							CodeBlock{
								Lines: []string{
									"func pointOfNoReturn(n int) (r int) {",
									"\tdefer func() {",
									"\t\tp := recover()",
									"\t\tr = p + 1",
									"\t}()",
									"\tpanic(n - 1)",
									"}",
								},
							},
						},
					},
				},
			},
			Section{
				Level: 1,
				Heading: "Section 2",
				Content: []Renderable{
					Paragraph{
						Content: []Renderable{
							Text("At vero eos et accusamus et iusto odio dignissimos "),
							Italic("ducimus qui blanditiis praesentium"),
							Text(" voluptatum deleniti atque corrupti quos dolores et "),
							Mono("quas molestias excepturi sint occaecati"),
							Text(" cupiditate non provident, similique sunt in culpa qui officia deserunt mollitia animi, id est laborum et dolorum fuga."),
						},
					},
					Paragraph{
						Content: []Renderable{
							Text("On the other hand, we denounce with righteous indignation and dislike men who are so beguiled and demoralized by the charms of pleasure of the moment, so blinded by desire, that they cannot foresee the pain and trouble that are bound to ensue; and equal blame belongs to those who fail in their duty through weakness of will, which is the same as saying through shrinking from toil and pain."),
						},
					},
				},
			},
		},
		Relevant: &RelevantBox{
			Heading: "Articles from blogs I read",
			Articles: []ReadingItem{
				ReadingItem{
					Link: "#",
					Title: "Lorem Ipsum",
					AuthorLink: "#",
					Author: "Anonymous",
					Abstract: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
				},
				ReadingItem{
					Link: "#",
					Title: "Lorem Epsum",
					AuthorLink: "#",
					Author: "Anonymous",
					Abstract: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
				},
				ReadingItem{
					Link: "#",
					Title: "Lorem Ipsum The Sequel",
					AuthorLink: "#",
					Author: "Anonymous",
					Abstract: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.",
				},
			},
		},
	}
