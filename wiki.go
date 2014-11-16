package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

var wiki DB

func adapt(r *mux.Router, wikiF func() DB, f func(params map[string]string, wiki func() DB, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return NewLoggingMiddleware(os.Stdout, func(w http.ResponseWriter, r *http.Request) {
		f(mux.Vars(r), wikiF, w, r)
	})
}

func main() {
	var err error

	endpoint := ":3000"

	wiki, err = newFileDB("wiki_db")
	//wiki, err = newMemDB()
	if err != nil {
		panic(err.Error())
	}

	wikiF := func() DB {
		return wiki
	}

	r := mux.NewRouter()

	r.HandleFunc("/", adapt(r, wikiF, ListPagesHandler)).Methods("GET")
	r.HandleFunc("/About/", adapt(r, wikiF, AboutPageHandler)).Methods("GET")
	r.Handle("/static/{path:.*}", http.FileServer(http.Dir("public/")))
	r.HandleFunc("/edit/{name}/", adapt(r, wikiF, ShowEditPageHandler)).Methods("GET")
	r.HandleFunc("/edit/{name}/", adapt(r, wikiF, EditPageHandler)).Methods("POST")
	r.HandleFunc("/edit/:name/attachment/", adapt(r, wikiF, AddAttachmentHandler)).Methods("POST")
	r.HandleFunc("/{name}/", adapt(r, wikiF, PageHandler)).Methods("GET")
	r.HandleFunc("/{name}/{attachment}", adapt(r, wikiF, AttachmentHandler)).Methods("GET")

	os.Stdout.WriteString("Staring wiki at " + endpoint + "\n")
	http.ListenAndServe(endpoint, r)
}
