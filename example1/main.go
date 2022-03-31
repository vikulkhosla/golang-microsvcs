package main

import (
        "fmt"
        "io/ioutil"
        "net/http"

        "github.com/praetorian-inc/microservices-golang/nicohttp"
)

var (
        argJSONFilePath *string
)

func main() {
        b := nicohttp.GetBuilder()
        b = b.WithDefaults().WithBaseFlags()
        b, argJSONFilePath = b.WithStringFlag("jsonFilePath", "", "Path to JSON regions file", true)
        s, _ := b.Create("regions-service", 8080)
        s.Mux().HandleFunc("/regions", regions).Methods("GET")

        fmt.Printf("%v\n", b.Props())

        s.Start()
}

func regions(w http.ResponseWriter, r *http.Request) {
        dat, _ := ioutil.ReadFile(*argJSONFilePath)
        w.Header().Set("Content-Type", "application/json")
        w.Write(dat)
        fmt.Print(string(dat))
}
