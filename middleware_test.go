package main

import (
	"bytes"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVieworCreateMiddleware(t *testing.T) {
	created := false
	viewed := false
	var (
		createHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			created = true
		}
		viewHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			viewed = true
		}
	)

	wiki, _ := newMemDB()
	pageView, _ := wiki.GetPage("PageExists")
	pageView.AddRevision([]byte("AbcDef"))
	pageCreate, _ := wiki.GetPage("PageNew")

	wrapMware := func(m http.Handler) http.Handler {
		return NewMuxVarMiddleware(NewPageLookupMiddleware(wiki, m))
	}
	_ = pageCreate
	_ = createHandler
	_ = viewHandler
	Convey("The VieworCreateMiddleware dispatches to different paths depending on the status of the requested page", t, func() {
		m := NewViewCreateMiddleware(viewHandler, createHandler)
		chain := wrapMware(m)
		Convey("The factory function should return a non-nil value", func() {
			So(m, ShouldNotBeNil)
		})

		Convey("The middleware should call the next function", func() {
			Convey("For existing pages the next function is the view function", func() {
				created = false
				viewed = false
				record := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/page/PageExists/", nil)
				So(err, ShouldBeNil)

				mx := mux.NewRouter()
				mx.Path("/page/{name}/").Handler(chain)

				mx.ServeHTTP(record, req)
				So(record.Code, ShouldEqual, http.StatusOK)
				So(viewed, ShouldBeTrue)
				So(created, ShouldBeFalse)
			})
			Convey("For new pages the next function is the create function", func() {
				created = false
				viewed = false
				record := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/page/PageNew/", nil)
				So(err, ShouldBeNil)

				mx := mux.NewRouter()
				mx.Path("/page/{name}/").Handler(chain)

				mx.ServeHTTP(record, req)
				So(record.Code, ShouldEqual, http.StatusOK)
				So(viewed, ShouldBeFalse)
				So(created, ShouldBeTrue)
			})
		})
		Convey("For a junk page, you return a 404", func() {
			created = false
			viewed = false
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/page/invalid/", nil)
			So(err, ShouldBeNil)

			mx := mux.NewRouter()
			mx.Path("/page/{name}/").Handler(chain)

			mx.ServeHTTP(record, req)
			So(record.Code, ShouldEqual, http.StatusNotFound)
			So(viewed, ShouldBeFalse)
			So(created, ShouldBeFalse)
		})
	})
}

func TestPageLookupMiddleware(t *testing.T) {
	called := false
	var pageData Page
	var okHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		called = true
		pageData, _ = context.Get(r, keyPage).(Page)
	}
	db, _ := newMemDB()
	page, _ := db.GetPage("TestPage")
	page.AddRevision([]byte("Hello World!"))

	Convey("The PageLookupMiddleware retreives a page and sticks it into the request context", t, func() {
		m := NewMuxVarMiddleware(NewPageLookupMiddleware(db, okHandler))
		Convey("The factory function should return a non-nil value", func() {
			So(NewPageLookupMiddleware(db, okHandler), ShouldNotBeNil)
		})
		Convey("The middleware should call the next handler", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/page/TestPage/", nil)
			So(err, ShouldBeNil)

			mx := mux.NewRouter()
			mx.Path("/page/{name}/").Handler(m)

			mx.ServeHTTP(record, req)
			So(called, ShouldBeTrue)
		})
		Convey("The middleware should retreive the requested database page (TestPage) and store it in the context", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/page/TestPage/", nil)
			So(err, ShouldBeNil)

			mx := mux.NewRouter()
			mx.Path("/page/{name}/").Handler(m)

			mx.ServeHTTP(record, req)
			So(pageData, ShouldNotBeNil)

			data, err := pageData.GetData(CURRENT_REVISION)
			So(err, ShouldBeNil)
			So(bytes.Compare(data, []byte("Hello World!")), ShouldEqual, 0)
		})
	})
}

func TestRevMiddleware(t *testing.T) {
	called := false
	var rev int
	var okHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		called = true
		rev = CurRev(r)
	}

	Convey("The RevMiddleware examines the requested URL for revision information and puts the revision request in the context", t, func() {
		m := NewRevMiddleware(okHandler)
		Convey("The factory function should return a non-nil value", func() {
			So(m, ShouldNotBeNil)
		})
		Convey("The middleware should call the next handler", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/page/TestPage/", nil)
			So(err, ShouldBeNil)

			mx := mux.NewRouter()
			mx.Path("/page/{name}/").Handler(m)

			mx.ServeHTTP(record, req)
			So(called, ShouldBeTrue)
		})
		Convey("Requests with no revision specified should set the revision as CURRENT_REVISION", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/page/TestPage/", nil)
			So(err, ShouldBeNil)

			mx := mux.NewRouter()
			mx.Path("/page/{name}/").Handler(m)

			mx.ServeHTTP(record, req)
			So(rev, ShouldEqual, CURRENT_REVISION)
		})
		Convey("Requests with a revision specified should set the revision as the specified revision", func() {
			Convey("rev = 5", func() {
				record := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/page/TestPage/?rev=5", nil)
				So(err, ShouldBeNil)

				mx := mux.NewRouter()
				mx.Path("/page/{name}/").Handler(m)

				mx.ServeHTTP(record, req)
				So(rev, ShouldEqual, int(5))
			})
			Convey("rev = 3", func() {
				record := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/page/TestPage/?rev=3", nil)
				So(err, ShouldBeNil)

				mx := mux.NewRouter()
				mx.Path("/page/{name}/").Handler(m)

				mx.ServeHTTP(record, req)
				So(rev, ShouldEqual, int(3))
			})
		})
	})
}

func TestMuxVarLoadingMiddleware(t *testing.T) {
	called := false
	var arg string
	var okHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		called = true
		arg = ""
		val := context.Get(r, keyParams)
		if val != nil {
			if params, ok := val.(map[string]string); ok {
				arg = params["name"]
			}
		}
	}

	Convey("The MuxVarMiddleware loads the parameters into the context", t, func() {
		m := NewMuxVarMiddleware(okHandler)
		Convey("The factory function should return a non-nil value", func() {
			So(m, ShouldNotBeNil)
		})
		Convey("The middleware should call the next handler", func() {
			called = false
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/about/", nil)
			So(err, ShouldBeNil)

			m.ServeHTTP(record, req)
			So(called, ShouldBeTrue)
		})
		Convey("The middleware should put positional arguments in the context", func() {
			record := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/page/test/", nil)
			So(err, ShouldBeNil)

			mx := mux.NewRouter()
			mx.Path("/page/{name}/").Handler(m)

			mx.ServeHTTP(record, req)
			So(arg, ShouldEqual, "test")
		})
	})
}
