package generator

type MongoGenerator struct{}

func (m *MongoGenerator) GenerateSchema(tables []Table) (string, error) {
	return "// Mongo schema generation not implemented yet\n", nil
}

func (m *MongoGenerator) ImportData(tables []Table) error {
	return nil
}
