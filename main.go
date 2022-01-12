package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("./blockchain_log.txt")
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, "This is the current blockchain in the PoK network:\n\n\n")
		fmt.Fprintf(w, (string(data)))
	})

	http.HandleFunc("/greet/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/greet/"):]
		fmt.Fprintf(w, "Hello %s\n", name)
	})

	http.HandleFunc("/chain/", handleUploading)

	http.ListenAndServe(":9090", nil)
}

func handleUploading(w http.ResponseWriter, r *http.Request) {
	// ParseMultipartForm parses a request body as multipart/form-data
	r.ParseMultipartForm(32 << 20)

	file, _, err := r.FormFile("file") // Retrieve the file from form data

	if err != nil {
		return
	}
	defer file.Close() // Close the file when we finish

	// This is path which we want to store the file
	f, err := os.OpenFile("./blockchain.txt", os.O_RDWR|os.O_CREATE, 0777)

	if err != nil {
		return
	}

	// Copy the file to the destination path
	io.Copy(f, file)
}
