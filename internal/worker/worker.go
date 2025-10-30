package worker

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"bigdataimporter/internal/generator"
	"bigdataimporter/internal/parser"
)

type Job struct {
	ID       string
	FilePath string
	Target   string
}

var jobQueue chan Job

// StartPool - worker havuzunu başlatır
func StartPool(workerCount int) {
	jobQueue = make(chan Job, 100)
	for i := 0; i < workerCount; i++ {
		go workerLoop(i)
	}
	log.Printf("Worker pool started with %d workers", workerCount)
}

// Enqueue - yeni bir işi kuyruğa ekler
func Enqueue(job Job) {
	if jobQueue == nil {
		log.Println("Worker pool not started")
		return
	}
	jobQueue <- job
	log.Printf("Job queued: %s (%s -> %s)", job.ID, job.FilePath, job.Target)
}

func workerLoop(id int) {
	for job := range jobQueue {
		log.Printf("[Worker %d] started job: %s", id, job.ID)
		processJob(job)
		log.Printf("[Worker %d] finished job: %s", id, job.ID)
	}
}

func processJob(job Job) {
	log.Printf("Processing job %s ...", job.ID)

	if _, err := os.Stat(job.FilePath); os.IsNotExist(err) {
		log.Printf("File not found: %s", job.FilePath)
		return
	}

	tables, err := parser.ParseSQLFile(job.FilePath)
	if err != nil {
		log.Printf("Parse error in %s: %v", job.FilePath, err)
		return
	}

	if len(tables) == 0 {
		log.Printf("No tables found in %s", job.FilePath)
		return
	}
	log.Printf("Parsed %d tables from %s", len(tables), job.FilePath)

	if err := os.MkdirAll("results", 0777); err != nil {
		log.Printf("results folder error: %v", err)
		return
	}

	// ✅ 1️⃣ Tüm tabloları generator formatına dönüştür
	var genTables []generator.Table
	for _, t := range tables {
		genTable := generator.Table{
			TableName:  t.TableName,
			Fields:     make([]generator.Field, len(t.Fields)),
			Engine:     t.Engine,
			Charset:    t.Charset,
			PrimaryKey: t.PrimaryKeys,
		}

		for fi, f := range t.Fields {
			if f.Name == "" {
				continue
			}
			genTable.Fields[fi] = generator.Field{
				Name:          f.Name,
				Type:          f.Type,
				Nullable:      f.Nullable,
				PrimaryKey:    f.PrimaryKey,
				AutoIncrement: f.AutoIncrement,
				Default:       f.Default,
				Index:         f.Index,
			}

			if f.ForeignKey != nil &&
				f.ForeignKey.ReferencedTable != "" &&
				f.ForeignKey.ReferencedField != "" {
				genTable.Fields[fi].ForeignKey = &generator.ForeignKey{
					ReferencedTable: f.ForeignKey.ReferencedTable,
					ReferencedField: f.ForeignKey.ReferencedField,
				}
			}
		}

		genTables = append(genTables, genTable)
	}

	// ✅ 2️⃣ Tüm tabloları tek seferde PostgreSQL'e çevir
	output, err := generator.GeneratePostgreSQLSchema(genTables)
	if err != nil {
		log.Printf("Generator error: %v", err)
		return
	}
	if len(output) == 0 {
		log.Printf("Empty SQL generated.")
		return
	}

	// ✅ 3️⃣ Sonuç dosyasını oluştur
	mergedPath := filepath.Join("results", fmt.Sprintf("merged_%s.sql", job.Target))
	if err := os.WriteFile(mergedPath, []byte(output), 0644); err != nil {
		log.Printf("Failed to write merged file: %v", err)
	} else {
		log.Printf("All converted tables merged into: %s", mergedPath)
	}

	log.Printf("Job %s completed successfully.", job.ID)
}
