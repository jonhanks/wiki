package main

import (
	"bytes"
	"github.com/codegangsta/martini"
	"github.com/russross/blackfriday"
	"html/template"
	"net/http"
)

var templates map[string]*template.Template = make(map[string]*template.Template)

func init() {
	file_list := []string{"list_pages", "about_page", "not_found", "edit_page", "wiki_page"}
	for _, page_name := range file_list {
		templates[page_name] = template.Must(template.ParseFiles("./templates/" + page_name + ".tmpl"))
	}
}

func ListPagesHandler(wiki func() DB, w http.ResponseWriter, r *http.Request) {
	var details struct {
		Pages []string
	}
	details.Pages, _ = wiki().ListPages()
	templates["list_pages"].Execute(w, &details)
}

func AboutPageHandler(w http.ResponseWriter, r *http.Request) {
	templates["about_page"].Execute(w, nil)
}

func PageHandler(params martini.Params, wiki func() DB, w http.ResponseWriter, r *http.Request) {
	PageName := params["name"]

	var details struct {
		PageName string
		Content  template.HTML
	}
	details.PageName = PageName

	rawPage, err := wiki().GetPage(PageName)
	if err != nil {
		if err != NOT_FOUND {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// page not found
		templates["not_found"].Execute(w, &details)
		return
	}
	for _, wikiWord := range ExtractWikiWords(rawPage) {
		word := string(wikiWord)
		rawPage = bytes.Replace(rawPage, wikiWord, []byte("["+word+"](/"+word+"/)"), -1)
	}
	details.Content = template.HTML(string(blackfriday.MarkdownCommon(rawPage)))
	templates["wiki_page"].Execute(w, &details)
}

func ShowEditPageHandler(params martini.Params, wiki func() DB, w http.ResponseWriter, r *http.Request) {
	PageName := params["name"]

	var details struct {
		PageName string
		PageSrc  string
	}
	details.PageName = PageName

	rawPage, err := wiki().GetPage(PageName)
	if err != nil {
		if err != NOT_FOUND {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	} else {
		details.PageSrc = string(rawPage)
	}
	templates["edit_page"].Execute(w, &details)
}

func EditPageHandler(params martini.Params, wiki func() DB, w http.ResponseWriter, r *http.Request) {
	PageName := params["name"]

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	src := r.FormValue("entry")

	wiki().SavePage(PageName, []byte(src))
	http.Redirect(w, r, "/"+PageName+"/", 302)
}
