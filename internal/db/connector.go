package db

import (
	"bigdataimporter/internal/config"
	"bigdataimporter/internal/parser"
	"database/sql"
)

type Connector interface {
	Connect() (*sql.DB, error)
	ApplySchema(conn *sql.DB, schema string) error
	ImportData(conn *sql.DB, tables []parser.ParsedTable) error
}

func SelectConnector(target string, cfg *config.Config) Connector {
	switch target {
	case "postgres", "postgresql":
		return &PostgresConnector{Cfg: cfg}
	// case "mongo":
	// 	return &MongoConnector{Cfg: cfg}
	// case "sqlite":
	// 	return &SQLiteConnector{Cfg: cfg}
	default:
		return nil
	}
}
