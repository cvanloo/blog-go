package main

import (
	"net/http"
	"mime"
	//"log"

	"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/markup/gen"
	. "github.com/cvanloo/blog-go/assert"
)

func init() {
	_ = mime.AddExtensionType(".js", "text/javascript")
	_ = mime.AddExtensionType(".css", "text/css")
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/blog", gen.Handler(markup.BlogTestStruct, PanicIf))
	/*http.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s NOT FOUND", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))*/
	PanicIf(http.ListenAndServe(":8081", nil))
}
