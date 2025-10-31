package generator

import (
	"fmt"
	"regexp"
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

func GeneratePostgreSQLSchema(tables []Table) (string, error) {
	var sb strings.Builder
	var allAlters []string
	var allIndexes []string

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

		if len(table.PrimaryKey) > 1 {
			sb.WriteString(fmt.Sprintf(",  PRIMARY KEY (%s)\n", strings.Join(table.PrimaryKey, ", ")))
		}
		sb.WriteString(");\n\n")
	}

	if len(allAlters) > 0 {
		sb.WriteString("-- Foreign Keys\n")
		for _, a := range allAlters {
			sb.WriteString(a + "\n")
		}
		sb.WriteString("\n")
	}

	if len(allIndexes) > 0 {
		sb.WriteString("-- Indexes\n")
		for _, i := range allIndexes {
			sb.WriteString(i + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func GeneratePostgreSQL(table Table) (string, error) {
	return GeneratePostgreSQLSchema([]Table{table})
}

type PostgreGenerator struct{}

func (p *PostgreGenerator) GenerateSchema(tables []Table) (string, error) {
	return GeneratePostgreSQLSchema(tables)
}

func (p *PostgreGenerator) ImportData(tables []Table) error {
	return nil
}

func NormalizePostgresSyntax(sql string) string {
	sql = strings.ReplaceAll(sql, "`", "\"")
	sql = strings.ReplaceAll(sql, "\\'", "''")
	sql = strings.ReplaceAll(sql, "’", "''")
	sql = strings.ReplaceAll(sql, "‘", "''")
	sql = strings.ReplaceAll(sql, "´", "''")
	sql = strings.ReplaceAll(sql, "\\\\", "\\")
	sql = strings.ReplaceAll(sql, "TRUE", "true")
	sql = strings.ReplaceAll(sql, "FALSE", "false")
	sql = strings.ReplaceAll(sql, "ENGINE=InnoDB", "")
	sql = strings.ReplaceAll(sql, "CHARSET=utf8mb4", "")
	sql = strings.ReplaceAll(sql, "\r", " ")
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "'0000-00-00'", "'1970-01-01'")
	sql = strings.ReplaceAll(sql, "'0000-00-00 00:00:00'", "'1970-01-01 00:00:00'")
	sql = strings.ReplaceAll(sql, " DEFAULT NULL", "")
	sql = strings.ReplaceAll(sql, "b'0'", "false")
	sql = strings.ReplaceAll(sql, "b'1'", "true")
	sql = strings.ReplaceAll(sql, " ,", ",")
	sql = strings.ReplaceAll(sql, ", ", ", ")
	return strings.TrimSpace(sql)
}

func SafeNormalize(sql string) string {
	re := regexp.MustCompile(`VALUES\s*\(([^)]+)\)`)
	return re.ReplaceAllStringFunc(sql, func(match string) string {
		values := match[len("VALUES (") : len(match)-1]
		parts := strings.Split(values, ",")
		for i, p := range parts {
			p = strings.TrimSpace(p)
			// Boş veya NULL geç
			if strings.EqualFold(p, "NULL") || p == "" {
				continue
			}
			// Eğer tırnak yoksa ve harf içermiyorsa tırnakla
			if !strings.Contains(p, "'") && regexp.MustCompile(`[A-Za-z]`).MatchString(p) {
				// içeri harf varsa olduğu gibi bırak
			} else if !strings.Contains(p, "'") {
				parts[i] = "'" + p + "'"
			}
		}
		return "VALUES (" + strings.Join(parts, ", ") + ")"
	})
}
