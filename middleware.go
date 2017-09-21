package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Middleware func(http.Handler) http.Handler
type Middlewares []Middleware
type Routes []Route

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
	Middlewares Middlewares
}

var routes = Routes{
	Route{"Notification", "POST", "/aaa", notificationGet, chain(logMw, tokenMw, recordMw)},
	Route{"Notification", "DELETE", "/aaa", notificationGet, chain(logMw, tokenMw, archiveMw)},
	Route{"Notification", "GET", "/bbb", notificationGet, chain(tokenMw)},
	Route{"Notification", "GET", "/ccc", notificationGet, chain()},
	Route{"Notification", "GET", "/ddd", notificationGet, nil},
}

func reverse(m Middlewares) Middlewares {
	for i := 0; i < len(m)/2; i++ {
		j := len(m) - i - 1
		m[i], m[j] = m[j], m[i]
	}
	return m
}

func NewRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		r.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(use(http.HandlerFunc(route.HandlerFunc), route.Middlewares))
	}
	return r
}

func chain(middleware ...Middleware) Middlewares {
	return middleware
}

func use(h http.Handler, middleware Middlewares) http.Handler {
	sortedMws := reverse(middleware)
	for _, m := range sortedMws {
		h = m(h)
	}
	return h
}

func readReqContent(r *http.Request) string {
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
	}
	// Restore the io.ReadCloser to its original state
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	// Use the content
	return string(bodyBytes)
}

func notificationGet(w http.ResponseWriter, r *http.Request) {
	log.Println("Notification received GET")
}

func tokenMw(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("TokenMiddleware", r.URL)
		body := readReqContent(r)
		log.Printf("Body : content :%s \n", body)
		h.ServeHTTP(w, r)
	})
}

func logMw(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("LogMiddleware", r.URL)
		body := readReqContent(r)
		log.Println("request", r.URL, r.Method)
		log.Printf("Body : content :%s \n", body)
		h.ServeHTTP(w, r)
	})
}

func archiveMw(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("ArchiveMiddleware", r.URL)
		body := readReqContent(r)
		log.Printf("Body : content :%s \n", body)
		h.ServeHTTP(w, r)
	})
}

func recordMw(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("RecordMiddleware", r.URL)
		body := readReqContent(r)
		log.Printf("Body : content :%s \n", body)
		h.ServeHTTP(w, r)
	})
}

func main() {
	r := NewRouter()
	log.Fatal(http.ListenAndServe(":9999", r))
}
