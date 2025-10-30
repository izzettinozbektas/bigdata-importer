package generator

import (
	"fmt"
	"strings"
)

type Field struct {
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	Nullable      bool        `json:"nullable"`
	PrimaryKey    bool        `json:"primary_key"`
	AutoIncrement bool        `json:"auto_increment"`
	Default       string      `json:"default"`
	Index         bool        `json:"index"`
	ForeignKey    *ForeignKey `json:"foreign_key"`
}

type ForeignKey struct {
	ReferencedTable string `json:"referenced_table"`
	ReferencedField string `json:"referenced_field"`
}

type Table struct {
	TableName  string   `json:"table_name"`
	Fields     []Field  `json:"fields"`
	Engine     string   `json:"engine,omitempty"`
	Charset    string   `json:"charset,omitempty"`
	PrimaryKey []string `json:"primary_keys,omitempty"`
}

func MySQLToPostgreType(mysqlType string, autoIncrement bool) string {
	t := strings.ToLower(mysqlType)
	switch {
	case autoIncrement:
		return "SERIAL"
	case strings.Contains(t, "tinyint"):
		return "SMALLINT"
	case strings.Contains(t, "bigint"):
		return "BIGINT"
	case strings.Contains(t, "int"):
		return "INTEGER"
	case strings.Contains(t, "varchar"):
		return t
	case strings.Contains(t, "text"):
		return "TEXT"
	case strings.Contains(t, "datetime"), strings.Contains(t, "timestamp"):
		return "TIMESTAMP"
	case strings.Contains(t, "date"):
		return "DATE"
	case strings.Contains(t, "decimal"):
		return "NUMERIC"
	case strings.Contains(t, "float"), strings.Contains(t, "double"):
		return "DOUBLE PRECISION"
	default:
		return "TEXT"
	}
}

// ✅ Tüm tabloları işler; ALTER ve INDEX'leri her zaman en alta yazar
func GeneratePostgreSQLSchema(tables []Table) (string, error) {
	var sb strings.Builder
	var allAlters []string
	var allIndexes []string

	// 1️⃣ CREATE TABLE'lar
	for _, table := range tables {
		if table.TableName == "" {
			continue
		}

		sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", table.TableName))

		for i, f := range table.Fields {
			pgType := MySQLToPostgreType(f.Type, f.AutoIncrement)
			col := fmt.Sprintf("  %s %s", f.Name, pgType)

			if !f.Nullable {
				col += " NOT NULL"
			}

			// DEFAULT
			defRaw := strings.TrimSpace(f.Default)
			def := strings.ToLower(defRaw)
			if defRaw != "" {
				switch {
				case def == "current_timestamp()" || def == "current_timestamp":
					col += " DEFAULT current_timestamp"
				case def == "null" || def == "NULL":
					col += " DEFAULT NULL"
				case strings.HasPrefix(pgType, "DATE") || strings.HasPrefix(pgType, "TIMESTAMP"):
					if defRaw != "0000-00-00" && defRaw != "'0000-00-00'" && defRaw != "''" {
						col += fmt.Sprintf(" DEFAULT '%s'", strings.Trim(defRaw, "'"))
					}
				case strings.HasPrefix(pgType, "INT") || strings.HasPrefix(pgType, "NUMERIC") ||
					strings.HasPrefix(pgType, "SMALLINT") || strings.HasPrefix(pgType, "BIGINT") ||
					strings.HasPrefix(pgType, "DOUBLE"):
					col += fmt.Sprintf(" DEFAULT %s", strings.Trim(defRaw, "'"))
				case def == "true" || def == "false" || def == "1" || def == "0":
					col += fmt.Sprintf(" DEFAULT %s", def)
				default:
					col += fmt.Sprintf(" DEFAULT '%s'", strings.Trim(defRaw, "'"))
				}
			}

			if f.PrimaryKey {
				col += " PRIMARY KEY"
			}

			if i < len(table.Fields)-1 {
				col += ","
			}
			sb.WriteString(col + "\n")

			// 🔹 ALTER ve INDEX sadece toplanıyor, hemen yazılmıyor
			if f.ForeignKey != nil && f.ForeignKey.ReferencedTable != "" && f.ForeignKey.ReferencedField != "" {
				fkName := fmt.Sprintf("fk_%s_%s", table.TableName, f.Name)
				allAlters = append(allAlters,
					fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s);",
						table.TableName, fkName, f.Name, f.ForeignKey.ReferencedTable, f.ForeignKey.ReferencedField))
			}
			if f.Index {
				allIndexes = append(allIndexes,
					fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_%s ON %s(%s);",
						table.TableName, f.Name, table.TableName, f.Name))
			}
		}

		// çoklu primary key desteği
		if len(table.PrimaryKey) > 1 {
			sb.WriteString(fmt.Sprintf(",  PRIMARY KEY (%s)\n", strings.Join(table.PrimaryKey, ", ")))
		}
		sb.WriteString(");\n\n")
	}

	// 2️⃣ ALTER’lar (tüm tablolar bittikten sonra)
	if len(allAlters) > 0 {
		sb.WriteString("-- Foreign Keys\n")
		for _, a := range allAlters {
			sb.WriteString(a + "\n")
		}
		sb.WriteString("\n")
	}

	// 3️⃣ INDEX’ler (en en sonda)
	if len(allIndexes) > 0 {
		sb.WriteString("-- Indexes\n")
		for _, i := range allIndexes {
			sb.WriteString(i + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// 💥 Eski kodlar bozulmasın diye ekledik
func GeneratePostgreSQL(table Table) (string, error) {
	return GeneratePostgreSQLSchema([]Table{table})
}
