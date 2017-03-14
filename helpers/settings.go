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

package helpers

import (
	"encoding/json"
	"io/ioutil"
)

type Database struct {
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

func (db Database) Credentials() string {
	return db.Username + ":" + db.Password + "@/" + db.Name
}

func DatabaseFromFile(filepath string) (Database, error) {
	fileBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return Database{}, err
	}
	var settings map[string]*json.RawMessage
	err = json.Unmarshal(fileBytes, &settings)
	if err != nil {
		return Database{}, err
	}
	var database Database
	err = json.Unmarshal(*settings["db"], &database)
	return database, nil
}
