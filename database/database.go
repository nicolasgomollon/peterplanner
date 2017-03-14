//
//  peterplanner
//  Copyright (c) 2017 Nicolas Gomollon <nicolas@gomollon.me>
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

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
