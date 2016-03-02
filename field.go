package csv

type (
	Field struct {
		Name  string
		Value string
		Tag   Tag
	}
	Fields []Field
)

func (a Fields) Len() int           { return len(a) }
func (a Fields) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Fields) Less(i, j int) bool { return a[i].Name < a[j].Name }

func (a Fields) Names() []string {
	names := make([]string, len(a))
	for i := range names {
		names[i] = a[i].Name
	}
	return names
}

func (a Fields) Values() []string {
	values := make([]string, len(a))
	for i := range values {
		values[i] = a[i].Value
	}
	return values
}
