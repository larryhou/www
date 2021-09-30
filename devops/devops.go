package devops

import (
    "encoding/json"
    "github.com/larryhou/urlmap/database"
    "github.com/larryhou/urlmap/model"
    "net/http"
    "regexp"
    "strings"
)

type Custom struct {
    Build string
}

func Handle(w http.ResponseWriter, r *http.Request)  {
    parts := strings.Split(r.URL.Path, "/")
    id := parts[len(parts)-1]
    m := &model.Mapping{}
    if _, err := database.Query(`SELECT * FROM urlmap WHERE id=?`, m, id); err == nil {
        data := m.Data.([]byte)
        c := &Custom{}
        json.Unmarshal(data, c)
        reg := regexp.MustCompile(`(iPad|iPhone|Android)`)
        if reg.MatchString(r.UserAgent()) {
            w.Header().Set("Location", m.Url)
        } else {
            if len(c.Build) > 0 { w.Header().Set("Location", c.Build) }
        }
        w.WriteHeader(http.StatusFound)
    }
}
