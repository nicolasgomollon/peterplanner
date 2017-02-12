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

func Post(url string, cookie string, data string) (statusCode int, responseBody string, err error) {
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
	request.Header.Set("Cookie", cookie)
	response, err := client.Do(request)
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
