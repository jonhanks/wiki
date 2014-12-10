package main

import (
	"github.com/gorilla/context"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestAdapter(t *testing.T) {
	Convey("The adapt function makes the custom handler type into a http.HandlerFunc type", t, func() {
		f := adapt(wiki, AboutPageHandler)
		So(f, ShouldNotBeNil)

		Convey("The returned handler will pass onto the final handler if it can get all the pices", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/about/", nil)
			So(err, ShouldBeNil)
			So(func() { f.ServeHTTP(record, req) }, ShouldNotPanic)
		})
	})
}

func TestAboutPageHandler(t *testing.T) {

	Convey("Test the about page handler, checking the responses", t, func() {
		record := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/about/", nil)
		if err != nil {
			t.Fatalf("Unable to create test request")
		}
		AboutPageHandler(&RequestInfo{Params: make(map[string]string)}, record, req)
		So(record.Code, ShouldEqual, http.StatusOK)
	})
}

func TestListPageHandler(t *testing.T) {
	wiki, _ := newMemDB()

	Convey("First we create a wiki database", t, func() {
		page, _ := wiki.GetPage("PageOne")
		page.AddRevision([]byte(""))
		page, _ = wiki.GetPage("PageOne")
		page.AddRevision([]byte(""))

		Convey("After adding pages we can test for responses from the handler", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("Unable to create test request")
			}
			ListPagesHandler(&RequestInfo{Params: make(map[string]string), DB: wiki}, record, req)
			So(record.Code, ShouldEqual, http.StatusOK)
		})
	})
}

func TestPageHandler(t *testing.T) {
	wiki, _ := newMemDB()
	Convey("First we create a wiki database", t, func() {
		page, _ := wiki.GetPage("PageOne")
		page.AddRevision([]byte("AbcDef"))
		page.AddRevision([]byte("abc"))
		newPage, _ := wiki.GetPage("WhichPage")

		Convey("After adding pages we can test for responses from the handler", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/PageOne/", nil)
			if err != nil {
				t.Fatalf("Unable to create test request")
			}
			context.Set(req, keyPage, page)
			PageHandler(&RequestInfo{Params: map[string]string{"name": "PageOne"}, DB: wiki}, record, req)
			context.Clear(req)
			So(record.Code, ShouldEqual, http.StatusOK)

			Convey("Testing various revisions", func() {
				for i := -1; i <= 2; i++ {
					record := httptest.NewRecorder()
					req, err := http.NewRequest("GET", "/PageOne/?rev="+strconv.Itoa(i), nil)
					if err != nil {
						t.Fatalf("Unable to create test request")
					}
					context.Set(req, keyPage, page)
					context.Set(req, keyRev, i)
					PageHandler(&RequestInfo{Params: map[string]string{"name": "PageOne"}, DB: wiki}, record, req)
					context.Clear(req)
					if i <= 1 {
						So(record.Code, ShouldEqual, http.StatusOK)
					} else {
						So(record.Code, ShouldEqual, http.StatusNotFound)
					}
				}
			})

			// should never happen
			// Convey("Testing for a page with an invalid page name should give an error", func() {
			// 	record := httptest.NewRecorder()
			// 	req, err := http.NewRequest("GET", "/Invalid/", nil)
			// 	if err != nil {
			// 		t.Fatalf("Unable to create test request")
			// 	}
			// 	context.Set(req, keyPage, page)
			// 	PageHandler(&RequestInfo{Params: map[string]string{"name": "Invalid"}, DB: wiki}, record, req)
			// 	context.Clear(req)
			// 	So(record.Code, ShouldEqual, http.StatusNotFound)
			// })

			Convey("Testing for a page with a page that does not exist should be ok", func() {
				record := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/WhichPage/", nil)
				if err != nil {
					t.Fatalf("Unable to create test request")
				}
				context.Set(req, keyPage, newPage)
				PageHandler(&RequestInfo{Params: map[string]string{"name": "WhichPage"}, DB: wiki}, record, req)
				context.Clear(req)
				So(record.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}

func TestAttachmentHandler(t *testing.T) {
	wiki, _ := newMemDB()

	Convey("First we create a wiki database", t, func() {
		page, _ := wiki.GetPage("PageOne")
		page.AddRevision([]byte("AbcDef"))
		page.AddAttachment(strings.NewReader("test"), "test.txt")

		Convey("After adding pages we can test for responses from the handler", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/PageOne/test.txt", nil)
			if err != nil {
				t.Fatalf("Unable to create test request")
			}
			AttachmentHandler(&RequestInfo{Params: map[string]string{"name": "PageOne", "attachment": "test.txt"}, DB: wiki}, record, req)
			So(record.Code, ShouldEqual, http.StatusOK)

			Convey("Testing for a non-existant attachment should give an error", func() {
				record := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/PageOne/missing.txt", nil)
				if err != nil {
					t.Fatalf("Unable to create test request")
				}
				AttachmentHandler(&RequestInfo{Params: map[string]string{"name": "PageOne", "attachment": "missing.txt"}, DB: wiki}, record, req)
				So(record.Code, ShouldEqual, http.StatusNotFound)

				Convey("Testing for a non-existant page should give an error", func() {
					record := httptest.NewRecorder()
					req, err := http.NewRequest("GET", "/Invalid/test.txt", nil)
					if err != nil {
						t.Fatalf("Unable to create test request")
					}
					AttachmentHandler(&RequestInfo{Params: map[string]string{"name": "Invalid", "attachment": "test.txt"}, DB: wiki}, record, req)
					So(record.Code, ShouldEqual, http.StatusNotFound)
				})
			})
		})
	})
}

func TestShowEditPageHandler(t *testing.T) {
	wiki, _ := newMemDB()

	Convey("First we create a wiki database", t, func() {
		page, _ := wiki.GetPage("PageOne")
		page.AddRevision([]byte(""))
		page, _ = wiki.GetPage("PageOne")
		page.AddRevision([]byte(""))

		Convey("After adding pages we can test for responses from the handler", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/edit/PageOne/", nil)
			if err != nil {
				t.Fatalf("Unable to create test request")
			}
			ShowEditPageHandler(&RequestInfo{Params: map[string]string{"name": "PageOne"}, DB: wiki}, record, req)
			So(record.Code, ShouldEqual, http.StatusOK)

			Convey("Testing for a page with an invalid page name should give an error", func() {
				record := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/edit/Invalid/", nil)
				if err != nil {
					t.Fatalf("Unable to create test request")
				}
				ShowEditPageHandler(&RequestInfo{Params: map[string]string{"name": "Invalid"}, DB: wiki}, record, req)
				So(record.Code, ShouldEqual, http.StatusNotFound)

				Convey("Editing a non-existant page (with a valid name) should work fine", func() {
					record := httptest.NewRecorder()
					req, err := http.NewRequest("GET", "/edit/NotYetCreated/", nil)
					if err != nil {
						t.Fatalf("Unable to create test request")
					}
					ShowEditPageHandler(&RequestInfo{Params: map[string]string{"name": "NotYetCreated"}, DB: wiki}, record, req)
					So(record.Code, ShouldEqual, http.StatusOK)
				})
			})
		})
	})
}

func TestEditPageHandler(t *testing.T) {
	wiki, _ := newMemDB()

	Convey("We start with a empty database and add a page", t, func() {
		record := httptest.NewRecorder()
		form := url.Values{}
		form.Add("entry", "hello\n=====\n")
		req, err := http.NewRequest("POST", "/edit/PageOne/", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("Unable to create test request")
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		EditPageHandler(&RequestInfo{Params: map[string]string{"name": "PageOne"}, DB: wiki}, record, req)
		So(record.Code, ShouldEqual, http.StatusFound)

		Convey("The wiki page count should be one", func() {
			count, err := wiki.CountPages()

			So(err, ShouldBeNil)
			So(count, ShouldEqual, 1)
		})
	})
}
