package svr

import (
    "github.com/google/uuid"
    "io"
    "log"
    "net/http"
)

type Message struct {
    Ret  int         `json:"ret"`
    Msg  string      `json:"msg"`
    Data interface{} `json:"data,omitempty"`
}

func NewReader(r interface{}) io.Reader {
    x := &reader{}
    switch r := r.(type) {
    case io.Reader     : x.Reader = r
    case *http.Response: x.Reader = r.Body
    case *http.Request : x.Reader = r.Body
    default: return nil
    }

    return x
}

type reader struct {
    io.Reader
}

func (x *reader) Read(p []byte) (n int, err error) {
    for m := 0; n < len(p) && err == nil; {
        m, err = x.Reader.Read(p[n:])
        n += m
    }
    return
}

type Response interface {
    Error() error
}

type Handle func(w http.ResponseWriter, r *http.Request)
func (x Handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    id, _ := uuid.NewUUID()
    log.Printf("%v %v %v #%v %v %+v %s\n", r.Proto, r.Method, r.URL, r.ContentLength, r.Header.Get("Content-Type"), r.RemoteAddr, id)
    w.Header().Set("H-Engine", "rapidci")
    w.Header().Set("H-Author", "larryhou")
    w.Header().Set("H-Uuid", id.String())
    x(w, r)
}

func FileHandle(dir string) http.Handler {
    return Handle(http.FileServer(http.Dir(dir)).ServeHTTP)
}