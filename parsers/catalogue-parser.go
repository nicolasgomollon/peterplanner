package parsers

import (
	"errors"
	"fmt"
	"github.com/nicolasgomollon/peterplanner/helpers"
	"net/http"
	"regexp"
	"strings"
)

/* Course Catalogue HTML Parser */

const CatalogueFormatURL = "http://catalogue.uci.edu/allcourses/%v"

func AllDepartments() (map[string]string, error) {
	statusCode, responseHTML, err := helpers.Get(fmt.Sprintf(CatalogueFormatURL, ""))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ERROR: Unable to fetch Course Catalogue HTML file. `%v`.", err.Error()))
	} else if statusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("ERROR: Unable to fetch Course Catalogue HTML file. HTTP status code: %v.", statusCode))
	}
	r, _ := regexp.Compile(`(?s)<div id="atozindex">(.*?)</div>`)
	index := r.FindStringSubmatch(responseHTML)[1]
	r, _ = regexp.Compile(`(?s)<li><a href="/allcourses/(.*?)">.*? \(([^\(\)]*?)\)</a></li>`)
	departments := r.FindAllStringSubmatch(index, -1)
	depts := make(map[string]string, 0)
	for _, element := range departments {
		path := element[1]
		dept := element[2]
		key := strings.Replace(strings.ToUpper(dept), " ", "", -1)
		depts[key] = fmt.Sprintf(CatalogueFormatURL, path)
	}
	return depts, nil
}

func FetchCatalogue(deptURL string) (string, error) {
	statusCode, responseHTML, err := helpers.Get(deptURL)
	if err != nil {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch Course Catalogue HTML file. `%v`.", err.Error()))
	} else if statusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch Course Catalogue HTML file. HTTP status code: %v.", statusCode))
	}
	return responseHTML, nil
}
