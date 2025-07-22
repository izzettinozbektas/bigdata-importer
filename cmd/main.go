package main

import (
	"bigdataimporter/internal/httpserver"
	"bigdataimporter/setup"
	"fmt"
	"net/http"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Starting bigdata-importer server...")
	})

	mux.HandleFunc("/upload-sql", httpserver.UploadSQLHandler)

	srv := setup.NewServer(mux)
	setup.StartServer(srv)
}
