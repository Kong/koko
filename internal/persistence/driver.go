package persistence

type Driver int

const (
	SQLite3 Driver = 0
	Postgres
)

func (d Driver) String() string {
	return [...]string{"sqlite3", "postgres"}[d]
}
