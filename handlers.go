package main

import (
	"bytes"
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
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

func ListPagesHandler(params map[string]string, wiki func() DB, w http.ResponseWriter, r *http.Request) {
	var details struct {
		Pages []string
	}
	details.Pages, _ = wiki().ListPages()
	templates["list_pages"].Execute(w, &details)
}

func AboutPageHandler(params map[string]string, wiki func() DB, w http.ResponseWriter, r *http.Request) {
	templates["about_page"].Execute(w, nil)
}

func PageHandler(params map[string]string, wiki func() DB, w http.ResponseWriter, r *http.Request) {
	PageName := params["name"]

	fmt.Println("PageHandler ", PageName)

	var rawPage []byte
	var details struct {
		PageName        string
		Content         template.HTML
		CurrentRevision int
		AttachmentList  []string
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
	details.AttachmentList, err = page.ListAttachments()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	details.CurrentRevision, minRevision, maxRevision = generateRevisionSplit(revision, revisionCount-1)
	details.RevisionList = generateInt(minRevision, maxRevision)

	if rawPage, err = page.GetData(revision); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rawPage = ExpandWikiWords(rawPage)

	// inject attachment information here
	buf := &bytes.Buffer{}

	for _, attachment := range details.AttachmentList {
		buf.WriteString("[" + attachment + "]: " + r.URL.String() + attachment + "\n")
	}
	buf.Write(rawPage)

	details.Content = template.HTML(string(blackfriday.MarkdownCommon(buf.Bytes())))
	templates["wiki_page"].Execute(w, &details)
}

func AttachmentHandler(params map[string]string, wiki func() DB, w http.ResponseWriter, r *http.Request) {
	PageName := params["name"]
	AttachmentName := params["attachment"]

	page, err := wiki().GetPage(PageName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	attachment, err := page.GetAttachment(AttachmentName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	stream, err := attachment.Open()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer stream.Close()
	io.Copy(w, stream)
}

func AddAttachmentHandler(params map[string]string, wiki func() DB, w http.ResponseWriter, r *http.Request) {
	PageName := params["name"]

	page, err := wiki().GetPage(PageName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_ = page
	f, _, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer f.Close()
	name := r.FormValue("name")
	fmt.Println("Adding attachment '" + name + "'")
	err = page.AddAttachment(f, name)
	if err != nil {
		fmt.Println("Error adding attachment", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/edit/"+PageName+"/", http.StatusFound)
}

func ShowEditPageHandler(params map[string]string, wiki func() DB, w http.ResponseWriter, r *http.Request) {
	PageName := params["name"]

	fmt.Println("ShowEditPageHandler ", PageName)

	var details struct {
		PageName       string
		PageSrc        string
		AttachmentList []string
	}
	details.PageName = PageName

	page, err := wiki().GetPage(PageName)
	if err != nil {
		fmt.Println(err)
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
			details.AttachmentList, err = page.ListAttachments()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
	templates["edit_page"].Execute(w, &details)
}

func EditPageHandler(params map[string]string, wiki func() DB, w http.ResponseWriter, r *http.Request) {
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
