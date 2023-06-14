package main

import (
    "flag"
    "fmt"
    "github.com/larryhou/urlmap/api/module/urlmap"
    "github.com/larryhou/urlmap/api/util/db"
    "github.com/larryhou/urlmap/api/util/svr"
    "log"
    "net/http"
    "os"
    "path"
)

func main() {
    opts := struct {
        key string
        pem string
    }{}
    flag.StringVar(&opts.key, `key`, `res/rapidsir.key`, `private key`)
    flag.StringVar(&opts.pem, `pem`, `res/rapidsir.pem`, `certificate`)
    flag.Parse()

    if f, err := os.OpenFile(path.Base(os.Args[0]) + `.log`, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0644); err == nil {
        log.SetOutput(f)
    }

    defer db.Close()

    mux := http.NewServeMux()
    mux.Handle(`/urlmap/`, svr.Handle(urlmap.Handle))
    mux.Handle(`/url/app/`, svr.Handle(func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, fmt.Sprintf(`https://rapidsir.com/urlmap/%s`, path.Base(r.URL.Path)), http.StatusMovedPermanently)
    }))
    mux.Handle("/", svr.Handle(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
    }))
    go http.ListenAndServe(":80", mux)
    log.Fatal(http.ListenAndServeTLS(":443", opts.pem, opts.key, mux))
}
