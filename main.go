package main

import (
    "flag"
    "github.com/larryhou/www/api/module/urlmap"
    "github.com/larryhou/www/api/util/db"
    "github.com/larryhou/www/api/util/svr"
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

    defer db.Close()
    if f, err := os.OpenFile(path.Base(os.Args[0]) + `.log`, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0644); err == nil {
        log.SetOutput(f)
    }

    mux := http.NewServeMux()
    mux.Handle(`/url` , svr.Handle(urlmap.Handle))
    mux.Handle(`/url/`, svr.Handle(urlmap.Handle))
    mux.Handle("/", svr.Handle(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
    }))

    go http.ListenAndServe(":80", mux)
    log.Fatal(http.ListenAndServeTLS(":443", opts.pem, opts.key, mux))
}
