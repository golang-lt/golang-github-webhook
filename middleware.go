package main

import (
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	w http.ResponseWriter
	c int
}

func (m *responseWriter) Header() http.Header {
	return m.w.Header()
}

func (m *responseWriter) Write(b []byte) (int, error) {
	return m.w.Write(b)
}

func (m *responseWriter) WriteHeader(c int) {
	m.c = c
	m.w.WriteHeader(c)
}

func timing(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		started := time.Now()
		ex := &responseWriter{w, 200}
		h.ServeHTTP(ex, req)
		log.Printf("%d - %s %s in %s\n", ex.c, req.Method, req.URL.String(), time.Since(started))
	})
}

func recoverable(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("cought panic during handler:", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		h.ServeHTTP(w, req)
	})
}
