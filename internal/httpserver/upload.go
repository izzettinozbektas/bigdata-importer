package httpserver

import (
	"bigdataimporter/internal/worker"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func UploadSQLHandler(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 1 << 30 // 1GB max file size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	target := r.FormValue("to")
	if target == "" {
		http.Error(w, "Eksik parametre: 'to' (örnek: postgres)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Dosya alınamadı: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err = os.MkdirAll("uploads", os.ModePerm); err != nil {
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

	if _, err = io.Copy(dst, file); err != nil {
		http.Error(w, fmt.Sprintf("Dosya kopyalanamadı: %v", err), http.StatusInternalServerError)
		return
	}

	job := worker.Job{
		ID:       fmt.Sprintf("job-%d", time.Now().UnixNano()),
		FilePath: dstPath,
		Target:   target,
	}

	worker.Enqueue(job)

	resp := map[string]interface{}{
		"message":   "Dosya alındı ve işleme kuyruğa eklendi",
		"job_id":    job.ID,
		"file_path": dstPath,
		"target":    target,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
