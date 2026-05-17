package repository

import "database/sql"

type Stor struct {
	*Queries
	db *sql.DB
}

// func NewStore(db *sql.DB) *Stor {
// 	return &Stor{
// 		Queries: New(db),
// 		db:      db,
// 	}
// }
