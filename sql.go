package urlmap

import (
    "github.com/larryhou/urlmap/database"
    "log"
)

func init() {
    err := database.Execute(`CREATE TABLE IF NOT EXISTS urlmap (
	id VARCHAR(64) NOT NULL PRIMARY KEY,
	url TEXT NOT NULL,
	data JSON,
	ts TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`)
    log.Printf("mysql init rapidci.urlmap %v", err)
}
