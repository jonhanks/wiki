package main

import (
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func NewPageLookupMiddleware(db DB, next http.Handler) http.Handler {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		params := CurParams(r)
		PageName, ok := params["name"]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		page, err := db.GetPage(PageName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		context.Set(r, keyPage, page)
		next.ServeHTTP(w, r)
	}
	return f
}

func NewRevMiddleware(next http.Handler) http.Handler {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		revision := CURRENT_REVISION
		var err error
		if revision, err = strconv.Atoi(r.FormValue("rev")); err != nil {
			revision = CURRENT_REVISION
		}
		context.Set(r, keyRev, revision)
		next.ServeHTTP(w, r)
	}
	return f
}

// The MuxVarMiddleware is used to decouble the rest of the middleware/handlers from the mux router
// it simply copies mux positional values into a known location in the context
func NewMuxVarMiddleware(next http.Handler) http.Handler {
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		context.Set(r, keyParams, vars)
		next.ServeHTTP(w, r)
	}
	return f
}
