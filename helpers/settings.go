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
	"net"
	"strconv"

	"github.com/go-sql-driver/mysql"
)

type Database struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

func (db Database) Credentials() string {
	host := db.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := db.Port
	if port == 0 {
		port = 3306
	}
	config := mysql.NewConfig()
	config.User = db.Username
	config.Passwd = db.Password
	config.Net = "tcp"
	config.Addr = net.JoinHostPort(host, strconv.Itoa(port))
	config.DBName = db.Name
	// Preserve the legacy driver's server-side max_allowed_packet lookup.
	config.MaxAllowedPacket = 0
	return config.FormatDSN()
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
