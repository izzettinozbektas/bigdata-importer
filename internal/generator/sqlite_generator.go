package generator

type SQLiteGenerator struct{}

func (s *SQLiteGenerator) GenerateSchema(tables []Table) (string, error) {
	return "// SQLite schema generation not implemented yet\n", nil
}

func (s *SQLiteGenerator) ImportData(tables []Table) error {
	return nil
}
