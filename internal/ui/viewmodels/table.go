package viewmodels

type ViewData struct {
	Error					*Error
	Tables				[]*Table
	Stats					*Stats
	Result				*Result
}

type Table struct {
	Id				int
	Name 			string
	RowCount	int16
	Columns		*ColumnList
}

type Stats struct {
	Version	string
	SqliteVersion	string
	Size		float64
}
