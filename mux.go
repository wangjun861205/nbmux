package nbmux

import (
	"fmt"
	"net/http"
	"strings"
)

var methodMap = map[string]Method{
	"GET":     GET,
	"HEAD":    HEAD,
	"POST":    POST,
	"PUT":     PUT,
	"DELETE":  DELETE,
	"CONNECT": CONNECT,
	"OPTIONS": OPTIONS,
	"TRACE":   TRACE,
}

type NBMux struct {
	root  *nbNode
	cache map[string]http.Handler
}

func NewMux(notFoundHandler http.Handler) *NBMux {
	return &NBMux{
		root:  newRoot(notFoundHandler),
		cache: make(map[string]http.Handler),
	}
}

func (mux *NBMux) AddHandler(exp string, method Method, handler http.Handler) error {
	if strings.Index(exp, "/") != 0 {
		return fmt.Errorf("regexp path must begin with '/' (%s)", exp)
	}
	expList := strings.Split(exp, "/")[1:]
	return mux.root.addChildren(expList, method, handler)
}

func (mux *NBMux) Search(path string, method string) (http.Handler, error) {
	if strings.Index(path, "/") != 0 {
		return nil, fmt.Errorf("request path must begin with '/' (%s)", path)
	}
	if h, ok := mux.cache[path]; ok {
		return h, nil
	}
	pathList := strings.Split(path, "/")[1:]
	h := mux.root.search(pathList, methodMap[method])
	if h == nil {
		return mux.root.handler, nil
	}
	mux.cache[path] = h
	return h, nil
}

func (mux *NBMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	method := r.Method
	handler, err := mux.Search(url, method)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	handler.ServeHTTP(w, r)
}
