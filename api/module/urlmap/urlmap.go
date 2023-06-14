package urlmap

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/larryhou/www/api/util/db"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)


func init() {
	db.Execute(`CREATE TABLE IF NOT EXISTS urlmap (
    id VARCHAR(64) NOT NULL PRIMARY KEY,
    url TEXT NOT NULL,
    ios TEXT,
    android TEXT,
    ts TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`)
}

type Entity struct {
	ID      string  `json:"id"`
	Url     string  `json:"url"`
	IOS     *string `json:"ios,omitempty"`
	Android *string `json:"android,omitempty"`
}

func Handle(w http.ResponseWriter, r *http.Request)  {
	switch r.Method {
	case http.MethodGet: (&get{}).Handle(w, r)
	case http.MethodPut: (&put{}).Handle(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

type get struct {}
func (x *get) Handle(w http.ResponseWriter, r *http.Request) {
	apple := regexp.MustCompile(`iPhone|iPad|iPod`)
	agent := r.UserAgent()

	parts := strings.Split(r.URL.Path, "/")
	id := parts[len(parts)-1]
	m := &Entity{}
	if _, err := db.Query(`SELECT * FROM urlmap WHERE id=?`, m, id); err == nil {
		u := m.Url
		switch {
		case strings.Contains(agent, `Android`):
			if m.Android != nil { u = *m.Android }
		case apple.MatchString(agent):
			if m.IOS != nil { u = *m.IOS }
		}

		http.Redirect(w, r, u, http.StatusFound)
	}
}


type Response struct {
	Ret  int         `json:"ret"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

type put struct { }
func (x *put) Handle(w http.ResponseWriter, r *http.Request) {
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

	data, err := io.ReadAll(r.Body)
	if err != nil {return}

	m := &Entity{}
	if err = json.Unmarshal(data, m); err != nil {return} else {
		log.Printf("urlmap put %+v", m)
		if len(m.Url) == 0 {
			err = fmt.Errorf("missing url field")
			return
		}

		sum := sha256.Sum256([]byte(m.Url))
		m.ID = base64.URLEncoding.EncodeToString(sum[:])
		err = db.Execute(`REPLACE INTO urlmap(id,url,ios,android) VALUES(?,?,?,?)`, m.ID, m.Url, m.IOS, m.Android)
		if err == nil { rsp.Data = m }
	}
}