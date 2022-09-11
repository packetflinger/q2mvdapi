package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

var Port *int
var Listen string
var Target *string
var ACLDomain *string
var AllowedIPs []string

type Files struct {
	NumFiles int
	Filename []string
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "")
}

// enumerate all the files in the specific directory.
// If we find MVD files, add to the structure then
// marshal the structure into JSON and send it
func ListFiles(w http.ResponseWriter, r *http.Request) {
	// have to split, looks like [::1]:xxxxx otherwise
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if !CheckACL(ip, AllowedIPs) {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Auth: %s not allowed\n", ip)
		return
	}

	filesfound := Files{}
	files, err := ioutil.ReadDir(*Target)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	total := 0
	for _, f := range files {
		// autorecorded demos in mid-recording start with _
		if strings.HasSuffix(f.Name(), ".mvd2") && (f.Name()[0] != '_') {
			filesfound.Filename = append(filesfound.Filename, f.Name())
			total++
		}
	}

	filesfound.NumFiles = total

	if len(filesfound.Filename) > 0 {
		data, _ := json.Marshal(filesfound)
		fmt.Fprintf(w, "%s", string(data))
	}
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if !CheckACL(ip, AllowedIPs) {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Auth: %s not allowed\n", ip)
		return
	}

	fname := r.URL.Query().Get("demo")
	path := fmt.Sprintf("%s/%s", *Target, fname)
	http.ServeFile(w, r, path)
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if !CheckACL(ip, AllowedIPs) {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Auth: %s not allowed\n", ip)
		return
	}

	fname := r.URL.Query().Get("demo")
	path := fmt.Sprintf("%s/%s", *Target, fname)
	err := os.Remove(path)
	if err != nil {
		fmt.Fprintln(w, err)
	}
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/GetMVDFiles", ListFiles)
	http.HandleFunc("/GrabFile", DownloadFile)
	http.HandleFunc("/NukeFile", DeleteFile)

	log.Fatal(http.ListenAndServe(Listen, nil))
}

// Request is allowed or not
func CheckACL(needle string, haystack []string) bool {
	for _, ip := range haystack {
		if needle == ip {
			return true
		}
	}

	return false
}

func main() {
	handleRequests()
}

// Parse the flags and setup the IPs allowed
func init() {
	Port = flag.Int("p", 27999, "The TCP port to listen on")
	Target = flag.String("d", ".", "The directory to look in")
	ACLDomain = flag.String("acl", "_acl.pfl.gr", "DNS name of allowed IPs")
	flag.Parse()

	Listen = fmt.Sprintf(":%d", *Port)

	// We use a simple DNS name for an ACL.
	// Look up the name, any IP address that resolves (v4 and v6) will be allowed
	ips, err := net.LookupTXT(*ACLDomain)
	if err != nil {
		log.Fatal(err)
	}

	for _, ip := range ips {
		current := strings.Split(ip, ",")
		AllowedIPs = append(AllowedIPs, current...)
	}

	log.Println("Allowed IPs:")
	log.Println(AllowedIPs)
}
