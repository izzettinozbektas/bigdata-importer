package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

type ForeignKeyMeta struct {
	ReferencedTable string `json:"referenced_table"`
	ReferencedField string `json:"referenced_field"`
}

type Field struct {
	Name          string          `json:"name"`
	Type          string          `json:"type"`
	Nullable      bool            `json:"nullable"`
	Default       string          `json:"default,omitempty"`
	PrimaryKey    bool            `json:"primary_key,omitempty"`
	Unique        bool            `json:"unique,omitempty"`
	AutoIncrement bool            `json:"auto_increment,omitempty"`
	Index         bool            `json:"index,omitempty"`
	ForeignKey    *ForeignKeyMeta `json:"foreign_key,omitempty"`
}

type ParsedTable struct {
	TableName   string   `json:"table_name"`
	Fields      []Field  `json:"fields"`
	UniqueKeys  []string `json:"unique_keys,omitempty"`
	Engine      string   `json:"engine,omitempty"`
	Charset     string   `json:"charset,omitempty"`
	PrimaryKeys []string `json:"primary_keys,omitempty"`
	Inserts     []string `json:"inserts,omitempty"` // eklendi
}

func ParseSQLFile(filePath string) ([]ParsedTable, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tables []ParsedTable
	var inserts []string

	scanner := bufio.NewScanner(file)

	var insideCreate bool
	var insideAlter bool
	var createLines []string
	var alterLines []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		upper := strings.ToUpper(line)

		// CREATE TABLE yakala
		if strings.HasPrefix(upper, "CREATE TABLE") {
			insideCreate = true
			createLines = []string{line}
			continue
		}

		// CREATE TABLE bloğu
		if insideCreate {
			createLines = append(createLines, line)
			if strings.HasSuffix(line, ");") {
				table, err := parseCreateBlock(createLines)
				if err == nil {
					tables = append(tables, table)
				}
				insideCreate = false
			}
			continue
		}

		// ALTER TABLE yakala
		if strings.HasPrefix(upper, "ALTER TABLE") {
			insideAlter = true
			alterLines = []string{line}
			continue
		}

		// ALTER TABLE bloğu
		if insideAlter {
			alterLines = append(alterLines, line)
			if strings.HasSuffix(line, ";") {
				joined := strings.Join(alterLines, " ")
				parseAlterStatement(joined, &tables)
				insideAlter = false
			}
			continue
		}

		// INSERT INTO yakala (yeni eklenen kısım)
		if strings.HasPrefix(upper, "INSERT INTO") {
			currentInsert := strings.TrimSpace(line)

			// Eğer satır ; ile bitmiyorsa devam satırlarını birleştir
			for !strings.HasSuffix(currentInsert, ";") {
				if !scanner.Scan() {
					break
				}
				nextLine := strings.TrimSpace(scanner.Text())
				currentInsert += " " + nextLine
			}

			inserts = append(inserts, currentInsert)
		}
	}

	// INSERT ifadelerini ilgili tablolara dağıt
	for _, insert := range inserts {
		parts := strings.SplitN(insert, " ", 4)
		if len(parts) > 2 {
			tableName := strings.Trim(parts[2], "`")
			for i := range tables {
				if tables[i].TableName == tableName {
					tables[i].Inserts = append(tables[i].Inserts, insert)
				}
			}
		}
	}

	return tables, scanner.Err()
}

func parseCreateBlock(lines []string) (ParsedTable, error) {
	var table ParsedTable
	var fields []Field
	var primaryKeys []string
	var uniqueKeys []string

	stmt := strings.Join(lines, " ")
	reTable := regexp.MustCompile("(?i)CREATE TABLE\\s+`([^`]+)`")
	if m := reTable.FindStringSubmatch(stmt); len(m) >= 2 {
		table.TableName = m[1]
	}

	pkRe := regexp.MustCompile(`(?i)PRIMARY KEY\s*\(([^)]+)\)`)
	if m := pkRe.FindStringSubmatch(stmt); len(m) >= 2 {
		keys := strings.Split(m[1], ",")
		for _, key := range keys {
			primaryKeys = append(primaryKeys, strings.Trim(key, "` "))
		}
		table.PrimaryKeys = primaryKeys
	}

	uqRe := regexp.MustCompile("(?i)UNIQUE KEY\\s+`[^`]+`\\s*\\(`([^`]+)`\\)")
	uqMatches := uqRe.FindAllStringSubmatch(stmt, -1)
	for _, m := range uqMatches {
		if len(m) >= 2 {
			uniqueKeys = append(uniqueKeys, m[1])
		}
	}
	table.UniqueKeys = uniqueKeys

	engineRe := regexp.MustCompile(`ENGINE=([a-zA-Z0-9]+)`)
	charsetRe := regexp.MustCompile(`CHARSET=([a-zA-Z0-9_]+)`)

	if m := engineRe.FindStringSubmatch(stmt); len(m) >= 2 {
		table.Engine = m[1]
	}
	if m := charsetRe.FindStringSubmatch(stmt); len(m) >= 2 {
		table.Charset = m[1]
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		colRe := regexp.MustCompile("^`([^`]+)`\\s+([a-zA-Z0-9()]+)(.*)")
		if matches := colRe.FindStringSubmatch(line); len(matches) >= 3 {
			name := matches[1]
			typeStr := matches[2]
			extra := matches[3]

			field := Field{
				Name:          name,
				Type:          typeStr,
				Nullable:      !strings.Contains(strings.ToUpper(extra), "NOT NULL"),
				Default:       extractDefault(extra),
				AutoIncrement: strings.Contains(strings.ToUpper(extra), "AUTO_INCREMENT"),
			}

			for _, pk := range primaryKeys {
				if strings.EqualFold(field.Name, pk) {
					field.PrimaryKey = true
				}
			}
			for _, uq := range uniqueKeys {
				if strings.EqualFold(field.Name, uq) {
					field.Unique = true
				}
			}

			fields = append(fields, field)
		}
	}

	table.Fields = fields
	return table, nil
}

func parseAlterStatement(line string, tables *[]ParsedTable) {
	pkRe := regexp.MustCompile(`(?i)ALTER TABLE\s+` + "`" + `([^` + "`" + `]+)` + "`" + `.*ADD PRIMARY KEY\s*\(([^)]+)\)`)
	if matches := pkRe.FindStringSubmatch(line); len(matches) >= 3 {
		tableName := matches[1]
		columns := strings.Split(matches[2], ",")
		for i := range *tables {
			if (*tables)[i].TableName == tableName {
				for _, col := range columns {
					col = strings.Trim(col, "` ")
					for j := range (*tables)[i].Fields {
						if strings.EqualFold((*tables)[i].Fields[j].Name, col) {
							(*tables)[i].Fields[j].PrimaryKey = true
							(*tables)[i].PrimaryKeys = appendIfMissing((*tables)[i].PrimaryKeys, col)
						}
					}
				}
			}
		}
	}

	autoIncRe := regexp.MustCompile(`(?i)ALTER TABLE\s+` + "`" + `([^` + "`" + `]+)` + "`" + `.*MODIFY\s+` + "`" + `([^` + "`" + `]+)` + "`" + `.*AUTO_INCREMENT`)
	if matches := autoIncRe.FindStringSubmatch(line); len(matches) >= 3 {
		tableName := matches[1]
		fieldName := matches[2]
		for i := range *tables {
			if (*tables)[i].TableName == tableName {
				for j := range (*tables)[i].Fields {
					if strings.EqualFold((*tables)[i].Fields[j].Name, fieldName) {
						(*tables)[i].Fields[j].AutoIncrement = true
					}
				}
			}
		}
	}

	addKeyRe := regexp.MustCompile(`(?i)ALTER TABLE\s+` + "`" + `([^` + "`" + `]+)` + "`" + `.*ADD KEY\s+` + "`" + `[^` + "`" + `]+` + "`" + `\s*\(([^)]+)\)`)
	if matches := addKeyRe.FindStringSubmatch(line); len(matches) >= 3 {
		tableName := matches[1]
		columns := strings.Split(matches[2], ",")
		for i := range *tables {
			if (*tables)[i].TableName == tableName {
				for _, col := range columns {
					col = strings.Trim(col, "` ")
					for j := range (*tables)[i].Fields {
						if strings.EqualFold((*tables)[i].Fields[j].Name, col) {
							(*tables)[i].Fields[j].Index = true
						}
					}
				}
			}
		}
	}

	fkRe := regexp.MustCompile(`(?i)ADD CONSTRAINT\s+` + "`" + `[^` + "`" + `]+` + "`" + `\s+FOREIGN KEY\s+\(` + "`" + `([^` + "`" + `]+)` + "`" + `\)\s+REFERENCES\s+` + "`" + `([^` + "`" + `]+)` + "`" + `\s+\(` + "`" + `([^` + "`" + `]+)` + "`" + `\)`)
	if matches := fkRe.FindStringSubmatch(line); len(matches) >= 4 {
		sourceField := matches[1]
		targetTable := matches[2]
		targetField := matches[3]

		for i := range *tables {
			for j := range (*tables)[i].Fields {
				if strings.EqualFold((*tables)[i].Fields[j].Name, sourceField) {
					(*tables)[i].Fields[j].ForeignKey = &ForeignKeyMeta{
						ReferencedTable: targetTable,
						ReferencedField: targetField,
					}
				}
			}
		}
	}
}

func extractDefault(extra string) string {
	re := regexp.MustCompile(`(?i)DEFAULT\s+('?[^',\s]+'?)`)
	matches := re.FindStringSubmatch(extra)
	if len(matches) >= 2 {
		return strings.Trim(matches[1], "'\",")
	}
	return ""
}

func appendIfMissing(slice []string, val string) []string {
	for _, item := range slice {
		if strings.EqualFold(item, val) {
			return slice
		}
	}
	return append(slice, val)
}
