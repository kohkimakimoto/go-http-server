package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

func main() {
	// parse flags...
	var optAddr, optDocroot string

	flag.StringVar(&optAddr, "a", "127.0.0.1:8080", "")
	flag.StringVar(&optAddr, "addr", "127.0.0.1:8080", "")
	flag.StringVar(&optDocroot, "d", ".", "")
	flag.StringVar(&optDocroot, "doc-root", ".", "")
	flag.Usage = func() {
		fmt.Println(`Usage: go-http-server [OPTIONS...]

Simple http server to deliver static files.

Options:
  -a, --addr <ADDRESS>     Specify listen address (default '127.0.0.0:8080')
  -d, --doc-root <PATH>    Specify the document root.
  -h, -help                Show help
`)
	}

	flag.Parse()

	if !filepath.IsAbs(optDocroot) {
		path, err := filepath.Abs(optDocroot)
		if err != nil {
			panic(err)
		}
		optDocroot = path
	}

	fs := http.FileServer(http.Dir(optDocroot))
	http.Handle("/", fs)

	log.Print("Listening '" + optAddr + "'...")
	log.Print("document root '" + optDocroot + "'")

	if err := http.ListenAndServe(optAddr, NoCache(Log(http.DefaultServeMux))); err != nil {
		panic(err)
	}
}

// The follwoing code is borrowed from https://github.com/zenazn/goji/blob/master/web/middleware/nocache.go

var epoch = time.Unix(0, 0).Format(time.RFC1123)

var noCacheHeaders = map[string]string{
	"Expires":         epoch,
	"Cache-Control":   "no-cache, private, max-age=0",
	"Pragma":          "no-cache",
	"X-Accel-Expires": "0",
}

var etagHeaders = []string{
	"ETag",
	"If-Modified-Since",
	"If-Match",
	"If-None-Match",
	"If-Range",
	"If-Unmodified-Since",
}

func NoCache(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		// Delete any ETag headers that may have been set
		for _, v := range etagHeaders {
			if r.Header.Get(v) != "" {
				r.Header.Del(v)
			}
		}

		// Set our NoCache headers
		for k, v := range noCacheHeaders {
			w.Header().Set(k, v)
		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// The follwoing code is borrowed from https://github.com/ajays20078/go-http-logger/blob/master/httpLogger.go

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	w.length = len(b)
	return w.ResponseWriter.Write(b)
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		writer := statusWriter{w, 0, 0}
		handler.ServeHTTP(&writer, r)
		end := time.Now()
		latency := end.Sub(start)
		statusCode := writer.status

		log.Printf("%s %s %s %d \"%s\" %v", r.RemoteAddr, r.Method, r.URL, statusCode, r.Header.Get("User-Agent"), latency)
	})
}
