package migration

type Migration struct {
	Num      int
	Title    string
	Up       bool
	Migrated bool
	Content  string
}
