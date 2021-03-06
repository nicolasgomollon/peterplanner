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

package parsers

import (
	"errors"
	"fmt"
	"github.com/nicolasgomollon/peterplanner/helpers"
	"github.com/nicolasgomollon/peterplanner/types"
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

func ParseCatalogue(responseHTML string, courses *map[string]types.Course) {
	s, _ := regexp.Compile(`([\r\n]+[\s\p{Zs}]?|[\s\p{Zs}])`)
	
	r, _ := regexp.Compile(`<h1>.*? \(([^\(\)]*?)\)</h1>`)
	dept := r.FindStringSubmatch(responseHTML)[1]
	dept = strings.Replace(strings.ToUpper(Clean(dept)), " ", "", -1)
	
	r, _ = regexp.Compile(`(?s)<div class="courses">(.*)</div></div>`)
	coursesBlock := r.FindStringSubmatch(responseHTML)[1]
	
	r, _ = regexp.Compile(`(?s)<div class="courseblock">.*?<p class="courseblocktitle"><strong>(.*?)\.\s*(.*?)\..*?</strong></p>.*?<div class="courseblockdesc">.*?<p>(.*?)</p>.*?</div>`)
	cs := r.FindAllStringSubmatch(coursesBlock, -1)
	
	for _, c := range cs {
		number := s.ReplaceAllString(strings.ToUpper(Clean(c[1])), "")[len(dept):]
		title := c[2]
		description := c[3]
		course := types.Course{Department: dept, Number: number, Title: title, Description: description}
		(*courses)[course.Key()] = course
	}
}
