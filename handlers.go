package main

import (
	"bytes"
	"github.com/codegangsta/martini"
	"github.com/russross/blackfriday"
	"html/template"
	"net/http"
	"strconv"
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

	var rawPage []byte
	var details struct {
		PageName        string
		Content         template.HTML
		CurrentRevision int
		RevisionList    <-chan int
	}
	var err error
	var minRevision, maxRevision int

	revision := CURRENT_REVISION
	if revision, err = strconv.Atoi(r.FormValue("rev")); err != nil {
		revision = CURRENT_REVISION
	}

	details.PageName = PageName

	page, err := wiki().GetPage(PageName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	revisionCount := page.Revisions()
	if revisionCount == 0 {
		// page not found
		templates["not_found"].Execute(w, &details)
		return
	}
	if rawPage, err = page.GetData(revision); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	details.CurrentRevision, minRevision, maxRevision = generateRevisionSplit(revision, revisionCount-1)
	details.RevisionList = generateInt(minRevision, maxRevision)

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

	page, err := wiki().GetPage(PageName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		if page.Revisions() > 0 {
			if rawPage, err := page.GetData(CURRENT_REVISION); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				details.PageSrc = string(rawPage)
			}
		}
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

	page, err := wiki().GetPage(PageName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	page.AddRevision([]byte(src))
	http.Redirect(w, r, "/"+PageName+"/", 302)
}