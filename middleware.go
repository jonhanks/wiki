package main

import (
	"io"
	"net/http"
	"strconv"
)

type statusResponseWriter struct {
	wrapped http.ResponseWriter
	status  int
}

func (l *statusResponseWriter) Header() http.Header {
	return l.wrapped.Header()
}

func (l *statusResponseWriter) Write(data []byte) (int, error) {
	l.setStatus(http.StatusOK)
	return l.wrapped.Write(data)
}

func (l *statusResponseWriter) WriteHeader(statusValue int) {
	l.setStatus(statusValue)
	l.wrapped.WriteHeader(statusValue)
}

func (l *statusResponseWriter) setStatus(statusValue int) {
	if l.status == 0 {
		l.status = statusValue
	}
}

func (l *statusResponseWriter) GetStatus() int {
	if l.status == 0 {
		return http.StatusOK
	}
	return l.status
}

func NewLoggingMiddleware(out io.Writer, chain http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statusWriter := &statusResponseWriter{wrapped: w}
		chain(statusWriter, r)
		code := statusWriter.GetStatus()
		out.Write([]byte(strconv.FormatInt(int64(code), 10) + " " + r.URL.Path + "\n"))
	}
}
