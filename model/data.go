package model

type Mapping struct {
    ID   string      `json:"id"`
    Url  string      `json:"url"`
    Data interface{} `json:"data,omitempty"`
}
