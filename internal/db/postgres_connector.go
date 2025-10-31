package db

import (
	"bigdataimporter/internal/config"
	"bigdataimporter/internal/generator"
	"bigdataimporter/internal/parser"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type PostgresConnector struct {
	Cfg *config.Config
}

func (p *PostgresConnector) Connect() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Cfg.Database.Host,
		p.Cfg.Database.Port,
		p.Cfg.Database.User,
		p.Cfg.Database.Password,
		p.Cfg.Database.Name,
		p.Cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping failed: %v", err)
	}

	log.Printf("PostgreSQL bağlantısı başarılı: %s:%d/%s",
		p.Cfg.Database.Host,
		p.Cfg.Database.Port,
		p.Cfg.Database.Name,
	)
	return db, nil
}

func (p *PostgresConnector) ApplySchema(conn *sql.DB, schema string) error {
	_, err := conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("schema apply error: %v", err)
	}
	log.Printf("Schema başarıyla uygulandı (%s)", p.Cfg.Database.Name)
	return nil
}
func (p *PostgresConnector) ImportData(conn *sql.DB, tables []parser.ParsedTable) error {
	logDir := "logs"
	os.MkdirAll(logDir, 0755)
	failedFile := logDir + "/failed_inserts.sql"

	for _, t := range tables {
		if len(t.Inserts) == 0 {
			continue
		}

		log.Printf("Importing %d inserts into %s...", len(t.Inserts), t.TableName)
		for _, insertSQL := range t.Inserts {
			normalized := generator.NormalizePostgresSyntax(insertSQL)
			normalized = generator.SafeNormalize(normalized)

			if _, err := conn.Exec(normalized); err != nil {
				log.Printf("Insert error in %s: %v", t.TableName, err)
				os.WriteFile(failedFile, []byte(normalized+";\n\n"), os.ModeAppend)
			}
		}
		log.Printf("%s data imported successfully", t.TableName)
	}
	return nil
}
