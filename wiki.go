package main

import (
	"github.com/gorilla/mux"
	"github.com/jonhanks/middleware"
	"github.com/justinas/alice"
	"net/http"
	"os"
)

var wiki DB

func main() {
	var err error

	endpoint := ":3000"

	wiki, err = newFileDB("wiki_db")
	//wiki, err = newMemDB()
	if err != nil {
		panic(err.Error())
	}

	stdMw := alice.New(middleware.MustGet("middleware.LoggingStdOut"), middleware.MustGet("middleware.Panic"))

	r := mux.NewRouter()

	r.Handle("/", stdMw.Then(adapt(wiki, ListPagesHandler))).Methods("GET")
	r.Handle("/About/", stdMw.Then(adapt(wiki, AboutPageHandler))).Methods("GET")
	r.Handle("/static/{path:.*}", http.FileServer(http.Dir("public/")))
	r.Handle("/edit/{name}/", stdMw.Then(adapt(wiki, ShowEditPageHandler))).Methods("GET")
	r.Handle("/edit/{name}/", stdMw.Then(adapt(wiki, EditPageHandler))).Methods("POST")
	r.Handle("/edit/:name/attachment/", stdMw.Then(adapt(wiki, AddAttachmentHandler))).Methods("POST")
	r.Handle("/{name}/", stdMw.Then(adapt(wiki, PageHandler))).Methods("GET")
	r.Handle("/{name}/{attachment}", stdMw.Then(adapt(wiki, AttachmentHandler))).Methods("GET")

	os.Stdout.WriteString("Staring wiki at " + endpoint + "\n")
	http.ListenAndServe(endpoint, r)
}
