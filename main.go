package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var Port *int
var Listen string
var Target *string

type Files struct {
	Filename []string
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "")
}

func ListFiles(w http.ResponseWriter, r *http.Request) {
	filesfound := Files{}
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".mvd2") {
			filesfound.Filename = append(filesfound.Filename, f.Name())
		}
	}

	if len(filesfound.Filename) > 0 {
		data, _ := json.Marshal(filesfound)
		fmt.Fprintf(w, "%s", string(data))
	}
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/GetMVDFiles", ListFiles)
	log.Fatal(http.ListenAndServe(Listen, nil))
}

func main() {
	handleRequests()
}

func init() {
	Port = flag.Int("p", 27999, "The TCP port to listen on")
	Target = flag.String("d", ".", "The directory to look in")
	flag.Parse()

	Listen = fmt.Sprintf(":%d", *Port)
}
