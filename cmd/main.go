package main

import (
	"bigdataimporter/internal/httpserver"
	"bigdataimporter/internal/worker"
	"bigdataimporter/setup"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		fmt.Println("logs klasörü oluşturulamadı:", err)
	}

	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
		log.Println("Loglama başlatıldı -> logs/app.log")
	} else {
		log.Println("Log dosyası oluşturulamadı, terminale yazılıyor:", err)
	}

	worker.StartPool(4)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Starting bigdata-importer server...")
	})

	mux.HandleFunc("/upload-sql", httpserver.UploadSQLHandler)

	srv := setup.NewServer(mux)
	setup.StartServer(srv)
}
