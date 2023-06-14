package main

import (
    "encoding/json"
    _ "github.com/larryhou/www/api/module/urlmap"
    "github.com/larryhou/www/api/util/db"
    "os"
    "strings"
    "time"
)

func main() {
    if f, err := os.Open(os.Args[1]); err != nil {panic(err)} else {
        var data [][4]string
        j := json.NewDecoder(f)
        if err = j.Decode(&data); err != nil {panic(err)}
        var records [][]interface{}
        for n := range data {
            it := data[n]
            u := it[2]
            var i, a *string
            switch {
            case strings.Contains(it[1],`.apk`): a = &it[1]
            case strings.Contains(it[1],`.ipa`): i = &it[1]
            }
            t, _ := time.ParseInLocation(`2006-01-02 15:04:05.000000`, it[3], time.Local)
            records = append(records, []interface{} {
                it[0], u, i, a, t,
            })
        }

        err = db.ExecuteMany(`REPLACE INTO urlmap(id,url,ios,android,ts) VALUES(?,?,?,?,?)`, records...)
        if err != nil {panic(err)}
    }
}