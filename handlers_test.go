package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestListPageHandler(t *testing.T) {
	wiki := newMemDB()
	f := func() DB {
		return wiki
	}

	Convey("First we create a wiki database", t, func() {
		wiki.SavePage("PageOne", []byte(""))
		wiki.SavePage("PageTwo", []byte(""))
		Convey("After adding pages we can test for responses from the handler", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("Unable to create test request")
			}
			ListPagesHandler(f, record, req)
			So(record.Code, ShouldEqual, http.StatusOK)
		})
	})
}

func TestPageHandler(t *testing.T) {
	wiki := newMemDB()
	f := func() DB {
		return wiki
	}

	Convey("First we create a wiki database", t, func() {
		wiki.SavePage("PageOne", []byte(""))
		wiki.SavePage("PageTwo", []byte(""))
		Convey("After adding pages we can test for responses from the handler", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/PageOne/", nil)
			if err != nil {
				t.Fatalf("Unable to create test request")
			}
			PageHandler(map[string]string{"name": "PageOne"}, f, record, req)
			So(record.Code, ShouldEqual, http.StatusOK)

			Convey("Testing for a page with an invalid page name should give an error", func() {
				record := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/Invalid/", nil)
				if err != nil {
					t.Fatalf("Unable to create test request")
				}
				PageHandler(map[string]string{"name": "Invalid"}, f, record, req)
				So(record.Code, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}

func TestShowEditPageHandler(t *testing.T) {
	wiki := newMemDB()
	f := func() DB {
		return wiki
	}

	Convey("First we create a wiki database", t, func() {
		wiki.SavePage("PageOne", []byte(""))
		wiki.SavePage("PageTwo", []byte(""))
		Convey("After adding pages we can test for responses from the handler", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/edit/PageOne/", nil)
			if err != nil {
				t.Fatalf("Unable to create test request")
			}
			ShowEditPageHandler(map[string]string{"name": "PageOne"}, f, record, req)
			So(record.Code, ShouldEqual, http.StatusOK)

			Convey("Testing for a page with an invalid page name should give an error", func() {
				record := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/edit/Invalid/", nil)
				if err != nil {
					t.Fatalf("Unable to create test request")
				}
				ShowEditPageHandler(map[string]string{"name": "Invalid"}, f, record, req)
				So(record.Code, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}

func TestEditPageHandler(t *testing.T) {
	wiki := newMemDB()
	f := func() DB {
		return wiki
	}

	Convey("We start with a empty database and add a page", t, func() {
		record := httptest.NewRecorder()
		form := url.Values{}
		form.Add("entry", "hello\n=====\n")
		req, err := http.NewRequest("POST", "/edit/PageOne/", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("Unable to create test request")
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		EditPageHandler(map[string]string{"name": "PageOne"}, f, record, req)
		So(record.Code, ShouldEqual, http.StatusFound)

		Convey("The wiki page count should be one", func() {
			count, err := wiki.CountPages()

			So(err, ShouldBeNil)
			So(count, ShouldEqual, 1)
		})
	})
}
