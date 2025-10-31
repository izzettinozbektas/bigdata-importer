package worker

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"bigdataimporter/internal/executor"
	"bigdataimporter/internal/generator"
	"bigdataimporter/internal/parser"
)

type Job struct {
	ID       string
	FilePath string
	Target   string
}

var jobQueue chan Job

func StartPool(workerCount int) {
	jobQueue = make(chan Job, 100)
	for i := 0; i < workerCount; i++ {
		go workerLoop(i)
	}
	log.Printf("Worker pool started with %d workers", workerCount)
}

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

	parsedTables, err := parser.ParseSQLFile(job.FilePath)
	if err != nil {
		log.Printf("Parse error in %s: %v", job.FilePath, err)
		return
	}
	if len(parsedTables) == 0 {
		log.Printf("No tables found in %s", job.FilePath)
		return
	}
	log.Printf("Parsed %d tables from %s", len(parsedTables), job.FilePath)

	if err := os.MkdirAll("results", 0777); err != nil {
		log.Printf("results folder error: %v", err)
		return
	}

	var genTables []generator.Table
	for _, t := range parsedTables {
		genTable := generator.Table{
			TableName:  t.TableName,
			Fields:     make([]generator.Field, len(t.Fields)),
			Engine:     t.Engine,
			Charset:    t.Charset,
			PrimaryKey: t.PrimaryKeys,
		}
		for fi, f := range t.Fields {
			genField := generator.Field{
				Name:          f.Name,
				Type:          f.Type,
				Nullable:      f.Nullable,
				PrimaryKey:    f.PrimaryKey,
				AutoIncrement: f.AutoIncrement,
				Default:       f.Default,
				Index:         f.Index,
			}
			if f.ForeignKey != nil {
				genField.ForeignKey = &generator.ForeignKey{
					ReferencedTable: f.ForeignKey.ReferencedTable,
					ReferencedField: f.ForeignKey.ReferencedField,
				}
			}
			genTable.Fields[fi] = genField
		}
		genTables = append(genTables, genTable)
	}

	gen := selectGenerator(job.Target)
	if gen == nil {
		log.Printf("Unsupported target: %s", job.Target)
		return
	}

	output, err := gen.GenerateSchema(genTables)
	if err != nil {
		log.Printf("Schema generation error: %v", err)
		return
	}
	if len(output) == 0 {
		log.Printf("Empty schema generated for %s", job.Target)
		return
	}

	mergedPath := filepath.Join("results", fmt.Sprintf("merged_%s.sql", job.Target))
	if err := os.WriteFile(mergedPath, []byte(output), 0644); err != nil {
		log.Printf("Failed to write merged file: %v", err)
	} else {
		log.Printf("Schema exported: %s", mergedPath)
	}

	if err := gen.ImportData(genTables); err != nil {
		log.Printf("Data import failed: %v", err)
	}

	log.Printf("Job %s completed successfully.", job.ID)

	go func() {
		log.Printf("Import başlatılıyor: %s (%s)", mergedPath, job.Target)
		executor.Run(executor.Job{
			ID:       job.ID + "-import",
			FilePath: mergedPath,
			Target:   job.Target,
		}, parsedTables)
	}()
}

func selectGenerator(target string) generator.Generator {
	switch target {
	case "postgres", "postgresql":
		return &generator.PostgreGenerator{}
	case "mongo", "mongodb":
		return &generator.MongoGenerator{}
	case "sqlite":
		return &generator.SQLiteGenerator{}
	default:
		return nil
	}
}
