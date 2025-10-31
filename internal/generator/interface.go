package generator

type Generator interface {
	GenerateSchema(tables []Table) (string, error)
	ImportData(tables []Table) error
}
