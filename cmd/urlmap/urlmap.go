package main

import (
	"fmt"
	"github.com/larryhou/www/api/util"
	"github.com/larryhou/www/api/util/svr"
	"log"
	"net/http"
	"path"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle(`/url/app/`, svr.Handle(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf(`https://rapidsir.com/url/%s`, path.Base(r.URL.Path)), http.StatusMovedPermanently)
	}))
	mux.Handle(`/urlmap/`, svr.Handle(func(w http.ResponseWriter, r *http.Request) {
		util.Request(nil, `https://rapidsir.com/url`, nil, util.Attachment{
			Method: http.MethodPut,
			Data:   r.Body,
		}, nil)
	}))
	mux.Handle(`/`, svr.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	log.Fatal(http.ListenAndServe(`:8080`, mux))
}
