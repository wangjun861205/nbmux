package nbmux

import (
	"fmt"
	"net/http"
	"testing"
)

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("not found"))
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test"))
}

func TestMux(t *testing.T) {
	nfh := http.HandlerFunc(notFoundHandler)
	h := http.HandlerFunc(handler)
	mux := NewMux(nfh)
	err := mux.AddHandler(`/`, ALL, h)
	if err != nil {
		panic(err)
	}
	err = mux.AddHandler(`/html`, ALL, h)
	if err != nil {
		panic(err)
	}
	hdr, err := mux.Search("/hello", "POST")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%p\n", nfh)
	fmt.Printf("%p\n", hdr)
}
