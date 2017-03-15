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
	"bitbucket.org/zombiezen/cardcpx/natsort"
	"errors"
	"fmt"
	"github.com/kennygrant/sanitize"
	"github.com/nicolasgomollon/peterplanner/helpers"
	"github.com/nicolasgomollon/peterplanner/types"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

/* WebSOC Prerequisites HTML Parser */

const PrereqsURL = "https://www.reg.uci.edu/cob/prrqcgi"
const PrereqsFormatURL = "https://www.reg.uci.edu/cob/prrqcgi?dept=%v&action=view_all&term=%v"

func PDepartmentOptions() (string, map[string]string, error) {
	statusCode, responseHTML, err := helpers.Get(PrereqsURL)
	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("ERROR: Unable to fetch WebSOC Prerequisites HTML file. `%v`.", err.Error()))
	} else if statusCode != http.StatusOK {
		return "", nil, errors.New(fmt.Sprintf("ERROR: Unable to fetch WebSOC Prerequisites HTML file. HTTP status code: %v.", statusCode))
	}
	r, _ := regexp.Compile(`(?s)<select name="term">(?:.*?)<option value="(.*?)">(?:.*?)</select>`)
	term := r.FindStringSubmatch(responseHTML)[1]
	r, _ = regexp.Compile(`(?s)<select name="dept">(.*?)</select>`)
	departments := r.FindStringSubmatch(responseHTML)[1]
	r, _ = regexp.Compile(`<option>(.*?)\r?\n`)
	options := r.FindAllStringSubmatch(departments, -1)
	deptOptions := make(map[string]string, 0)
	for _, option := range options {
		key := strings.Replace(strings.ToUpper(option[1]), " ", "", -1)
		deptOptions[key] = url.QueryEscape(option[1])
	}
	return term, deptOptions, nil
}

func FetchPrerequisites(term string, option string) (string, error) {
	statusCode, responseHTML, err := helpers.Get(fmt.Sprintf(PrereqsFormatURL, option, term))
	if err != nil {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch WebSOC Prerequisites HTML file. `%v`.", err.Error()))
	} else if statusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch WebSOC Prerequisites HTML file. HTTP status code: %v.", statusCode))
	}
	return responseHTML, nil
}

func ParsePrerequisites(responseHTML string, courses *map[string]types.Course) {
	r, _ := regexp.Compile(`(?s)<table border="0">.*?<tr>.*?<td align="center">.*?<h3>(.*?)</h3>.*?</td>`)
	dept := r.FindStringSubmatch(responseHTML)[1]
	dept = strings.Replace(strings.ToUpper(dept), " ", "", -1)
	
	r, _ = regexp.Compile(`(?s)<table width="800"(?:[^>]*)>(.*?)<\/table>`)
	rawPrereqs := r.FindStringSubmatch(responseHTML)[1]
	
	r, _ = regexp.Compile(`(?s)<tr>(.*?)<\/tr>`)
	prereqs := r.FindAllStringSubmatch(rawPrereqs, -1)
	
	r, _ = regexp.Compile(`(?s)<td.*?>(.*?)<\/td>`)
	
	p, _ := regexp.Compile(`(?s)(.*)<span.*?>\* (.*?) since .*?<\/span>`)
	
	for _, prereqSlice := range prereqs {
		prereq := prereqSlice[1]
		elements := r.FindAllStringSubmatch(prereq, -1)
		k := ""
		kp := ""
		t := ""
		for i, elementMatch := range elements {
			element := elementMatch[1]
			switch i {
			case 0:
				matches := p.FindStringSubmatch(element)
				if len(matches) > 0 {
					k = strings.ToUpper(Clean(matches[1]))
					k = strings.Replace(k, " ", "", -1)
					kp = strings.ToUpper(Clean(matches[2]))
					kp = strings.Replace(kp, " ", "", -1)
				} else {
					k = strings.ToUpper(Clean(element))
					k = strings.Replace(k, " ", "", -1)
				}
				break
			case 1:
				t = Clean(element)
				break
			case 2:
				if course, ok := (*courses)[k]; ok {
					course.ShortTitle = t
					course.Prerequisites = parsedPrerequisites(Clean(element))
					(*courses)[k] = course
				} else if len(dept) < len(k) {
					course := types.Course{Department: dept, Number: k[len(dept):], ShortTitle: t}
					course.Prerequisites = parsedPrerequisites(Clean(element))
					(*courses)[k] = course
				}
				if len(kp) > 0 {
					if course, ok := (*courses)[kp]; ok {
						course.ShortTitle = t
						course.Prerequisites = parsedPrerequisites(Clean(element))
						(*courses)[kp] = course
					} else if len(dept) < len(kp) {
						course := types.Course{Department: dept, Number: kp[len(dept):], ShortTitle: t}
						course.Prerequisites = parsedPrerequisites(Clean(element))
						(*courses)[kp] = course
					}
				}
				break
			default:
				break
			}
			if len(k) == 0 {
				break
			}
		}
	}
}

func Clean(element string) string {
	s1, _ := regexp.Compile(`(&#160;)`)
	s2, _ := regexp.Compile(`([\r\n]+[\s\p{Zs}]?|[\s\p{Zs}]{2,})`)
	element = sanitize.HTML(element)			// Remove HTML tags.
	element = s1.ReplaceAllString(element, " ")	// Replace non-breaking space with a space.
	element = strings.TrimSpace(element)		// Trim leading and trailing whitespace.
	element = s2.ReplaceAllString(element, " ")	// Replace consecutive spaces with one.
	element = html.UnescapeString(element)		// Decode HTML entities.
	return element
}

func parsedPrerequisites(rawPrereqs string) [][]string {
	r, _ := regexp.Compile(`(?s) \( (?:coreq|recommended|min score = [\w+-]+) \)`)
	element := r.ReplaceAllString(rawPrereqs, "")
	
	r, _ = regexp.Compile(`(?s) \( min grade = ([\w+-]+) \)`)
	element = r.ReplaceAllString(element, `|$1`)
	
	r, _ = regexp.Compile(`(?s)(\( | \))`)
	element = r.ReplaceAllString(element, "")
	
	prereqs := make([][]string, 0)
	rawArr := strings.Split(element, " AND ")
	for _, rawRow := range rawArr {
		row := strings.Split(rawRow, " OR ")
		cleanRow := make([]string, 0)
		for _, r := range row {
			if !strings.HasPrefix(r, "NO REPEATS ALLOWED") && !strings.HasPrefix(r, "BETTER") && !strings.HasPrefix(r, "SCHOOL OF") && !strings.HasPrefix(r, "PLACEMENT EXAM") {
				cleanRow = append(cleanRow, r)
			}
		}
		if len(cleanRow) > 0 {
			natsort.Strings(cleanRow)
			prereqs = append(prereqs, cleanRow)
		}
	}
	
	return prereqs
}
