package main

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	called := false
	okHandler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}

	Convey("The logging middleware should log the request path", t, func() {
		buf := bytes.NewBuffer(make([]byte, 0, 100))
		m := NewLoggingMiddleware(buf, okHandler)

		record := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/about/", nil)
		if err != nil {
			t.Fatalf("Unable to create test request")
		}
		m(record, req)
		So(record.Code, ShouldEqual, http.StatusOK)
		So(strings.Contains(buf.String(), "/about/"), ShouldBeTrue)
		Convey("The middleware should also have called the next function in the chain", func() {
			So(called, ShouldBeTrue)
		})
		Convey("The middleware should have caught the response code as well", func() {
			So(strings.Contains(buf.String(), "200"), ShouldBeTrue)
		})
		Convey("The last character in the log entry should be a newline", func() {
			tmp := buf.Bytes()
			So(tmp[len(tmp)-1], ShouldEqual, '\n')
		})
	})
}

func TestLoggingResponseWriter(t *testing.T) {
	Convey("The statusResponseWriter caches the http status value on a request", t, func() {
		record := httptest.NewRecorder()
		statusWriter := &statusResponseWriter{wrapped: record}

		Convey("The default status returned is a 200 (http.StatusOK)", func() {
			So(statusWriter.GetStatus(), ShouldEqual, http.StatusOK)
		})

		Convey("If you call WriteHeader with a value you should be able to read it back", func() {
			statusWriter.WriteHeader(http.StatusForbidden)
			So(statusWriter.GetStatus(), ShouldEqual, http.StatusForbidden)

			Convey("Calling WriteHeader a second time does not make sense, so you should only read back the original value", func() {
				statusWriter.WriteHeader(http.StatusConflict)
				So(statusWriter.GetStatus(), ShouldEqual, http.StatusForbidden)
			})
		})

		record = httptest.NewRecorder()
		record.Body = bytes.NewBuffer(make([]byte, 0, 100))
		statusWriter = &statusResponseWriter{wrapped: record}
		Convey("Calling Write on a StatusWriter should set the status value to StatusOK (the default anyways)", func() {
			statusWriter.Write([]byte("Hello World!"))

			So(statusWriter.GetStatus(), ShouldEqual, http.StatusOK)
			Convey("The value written should have been written through to the wrapped ResponseWriter", func() {
				So(record.Body.String(), ShouldEqual, "Hello World!")
			})
		})
		Convey("You should be able to call Headers() on the statusResponseWriter", func() {
			header := statusWriter.Header()
			So(header, ShouldNotBeNil)
		})
	})
}
