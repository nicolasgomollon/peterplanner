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
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

/* Exported Constants & Variables */
const Timeout = time.Duration(210 * time.Second)

/* Unexported Variables */
var client = http.Client{Timeout: Timeout}

func Get(url string) (statusCode int, responseBody string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	statusCode = 0
	response, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	statusCode = response.StatusCode

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	return statusCode, string(contents), nil
}

func Post(url string, cookie string, data string) (statusCode int, contentType string, responseBody string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	statusCode = 0
	request, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if len(cookie) > 0 {
		request.Header.Set("Cookie", cookie)
	}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	statusCode = response.StatusCode
	contentType = strings.Split(response.Header["Content-Type"][0], ";")[0]

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	return statusCode, contentType, string(contents), nil
}
