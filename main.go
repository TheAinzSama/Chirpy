package main

import (
	//"fmt"
	"net/http"
)

type reqHandler struct {
	Addr string
}

func (reqHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/ServeMux", &reqHandler{Addr: "8080"})
	http.ListenAndServe(":8080", mux)
}
