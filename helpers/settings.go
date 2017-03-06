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
