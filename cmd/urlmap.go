package main

import (
    "flag"
)

func main() {
    port := 80
    flag.IntVar(&port, "port", 8080, "server port")
    flag.Parse()

    c := &urlmap.Client{}
    c.Listen(int16(port))
}