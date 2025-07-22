package httpserver

import (
	"bigdataimporter/internal/parser"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func UploadSQLHandler(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 1 << 30 // 1GB max file size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Dosya alınamadı: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = os.MkdirAll("uploads", os.ModePerm)
	if err != nil {
		http.Error(w, "uploads klasörü oluşturulamadı", http.StatusInternalServerError)
		return
	}

	dstPath := filepath.Join("uploads", header.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Dosya oluşturulamadı: %v", err), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Dosya kopyalanamadı: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse process
	tables, err := parser.ParseSQLFile(dstPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Parse hatası: %v", err), http.StatusInternalServerError)
		return
	}

	// JSON return data type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tables)
}
