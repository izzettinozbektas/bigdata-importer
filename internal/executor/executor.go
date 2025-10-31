package executor

import (
	"bigdataimporter/internal/config"
	"bigdataimporter/internal/db"
	"bigdataimporter/internal/parser"
	"log"
	"os"
	"path/filepath"
)

type Job struct {
	ID       string
	FilePath string
	Target   string
}

func Run(job Job, tables []parser.ParsedTable) {
	wd, _ := os.Getwd()
	log.Printf("Current working directory: %s", wd)
	log.Printf("Executor started: %s -> %s", job.FilePath, job.Target)

	cfg, err := config.LoadConfig("/app/config.yaml")
	if err != nil {
		log.Printf("Config load error: %v", err)
		return
	}

	connector := db.SelectConnector(job.Target, cfg)
	if connector == nil {
		log.Printf("Unsupported target: %s", job.Target)
		return
	}

	conn, err := connector.Connect()
	if err != nil {
		log.Printf("DB connection failed: %v", err)
		return
	}
	defer conn.Close()

	_, _ = conn.Exec(`SET session_replication_role = replica;`)
	log.Println("Foreign key checks disabled temporarily")

	defer func() {
		_, _ = conn.Exec(`SET session_replication_role = DEFAULT;`)
		log.Println("Foreign key checks re-enabled (deferred)")
	}()

	content, err := os.ReadFile(filepath.Clean(job.FilePath))
	if err != nil {
		log.Printf("SQL file read error: %v", err)
		return
	}

	if err := connector.ApplySchema(conn, string(content)); err != nil {
		log.Printf("Schema apply error: %v", err)
		return
	}

	log.Printf("Schema successfully applied: %s", job.FilePath)

	if err := connector.ImportData(conn, tables); err != nil {
		log.Printf("Data import error: %v", err)
	} else {
		log.Printf("Data import completed successfully.")
	}
}
