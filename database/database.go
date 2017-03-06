package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nicolasgomollon/peterplanner/helpers"
)

/* Database Functions */

func Connect(database helpers.Database) (*sql.DB, error) {
	db, err := sql.Open(database.Kind, database.Credentials())
	if err != nil {
		return db, err
	}
	err = db.Ping()
	return db, err
}

func Execute(db *sql.DB, query string, args ...interface{}) (int64, error) {
	statement, err := db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer statement.Close()
	result, err := statement.Exec(args...)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func RowExists(db *sql.DB, query string, args ...interface{}) (bool, string, error) {
	exists := false
	var result string = ""
	err := db.QueryRow(query, args...).Scan(&result)
	switch {
	case err == sql.ErrNoRows:
		{
			exists = false
			err = nil
		}
	case err != nil:
		{
			exists = false
		}
	default:
		{
			exists = true
		}
	}
	return exists, result, err
}
