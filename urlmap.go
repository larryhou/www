package urlmap

import (
    "encoding/base64"
    "encoding/binary"
    "encoding/json"
    "fmt"
    "github.com/larryhou/urlmap/database"
    "github.com/larryhou/urlmap/devops"
    "github.com/larryhou/urlmap/model"
    "io/ioutil"
    "log"
    "net/http"
    "time"
    "unsafe"
)

type Handle func(w http.ResponseWriter, r *http.Request)
func (f Handle) ServeHTTP(w http.ResponseWriter, r *http.Request)  {
    log.Printf("%v %v %v %+v %+v\n", r.Proto, r.Method, r.URL, r.Header, r.RemoteAddr)
    w.Header().Set("#-Author", "larryhou")
    w.Header().Set("#-Engine", "urlmap")
    f(w, r)
}

type Response struct {
    Ret  int         `json:"ret"`
    Msg  string      `json:"msg"`
    Data interface{} `json:"data,omitempty"`
}

type Client struct {

}

func (c *Client) Handle(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    rsp := &Response{Ret: 0, Msg: "success"}
    var err error
    defer func() {
        if err != nil {
            rsp.Msg = fmt.Sprintf("%v", err)
            rsp.Ret = http.StatusBadRequest
        }
        j := json.NewEncoder(w)
        j.SetIndent("", "    ")
        j.SetEscapeHTML(false)
        j.Encode(rsp)
    }()

    data, err := ioutil.ReadAll(r.Body)
    if err != nil {return}

    m := &model.Mapping{}
    if err = json.Unmarshal(data, m); err != nil {return} else {
        log.Printf("%+v", m)
        if len(m.Url) == 0 {
            err = fmt.Errorf("missing url field")
            return
        }
        t := time.Now().UnixNano()
        b := make([]byte, unsafe.Sizeof(t))
        binary.BigEndian.PutUint64(b, uint64(t))
        m.ID = base64.URLEncoding.EncodeToString(b)
        data := ""
        if m.Data != nil { if d, err := json.Marshal(m.Data); err == nil {data = string(d)} }
        err = database.Execute(`REPLACE INTO urlmap(id,url,data) VALUES(?,?,?)`, m.ID, m.Url, data)
        if err == nil { rsp.Data = m }
    }
}

func (c *Client) Listen(port int16) {
    mux := http.NewServeMux()
    mux.Handle("/url/devops/", Handle(devops.Handle))
    mux.Handle("/urlmap", Handle(c.Handle))
    mux.Handle("/", Handle(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
    }))
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
