package parsers

import (
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
const PrereqsFormatURL = "https://www.reg.uci.edu/cob/prrqcgi?dept=%v&action=view_by_term&term=%v"

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
	r, _ := regexp.Compile(`(?s)<table width="800"(?:[^>]*)>(.*?)<\/table>`)
	rawPrereqs := r.FindStringSubmatch(responseHTML)[1]
	
	r, _ = regexp.Compile(`(?s)<tr>(.*?)<\/tr>`)
	prereqs := r.FindAllStringSubmatch(rawPrereqs, -1)
	
	r, _ = regexp.Compile(`(?s)<td.*?>(.*?)<\/td>`)
	
	for _, prereqSlice := range prereqs {
		prereq := prereqSlice[1]
		elements := r.FindAllStringSubmatch(prereq, -1)
		c := ""
		k := ""
		t := ""
		for i, elementMatch := range elements {
			element := elementMatch[1]
			element = sanitize.HTML(element)			// Remove HTML tags.
			element = strings.TrimSpace(element)		// Trim leading and trailing whitespace.
			s, _ := regexp.Compile(`([\r\n]+[\s\p{Zs}]?|[\s\p{Zs}]{2,})`)
			element = s.ReplaceAllString(element, " ")	// Replace consecutive spaces with one.
			element = html.UnescapeString(element)		// Decode HTML entities.
			switch i {
			case 0:
				c = strings.ToUpper(element)
				k = strings.Replace(c, " ", "", -1)
				break
			case 1:
				t = element
				break
			case 2:
				if course, ok := (*courses)[k]; ok {
					if len(course.Title) == 0 {
						course.Title = t
					}
					course.Prerequisites = parsedPrerequisites(element)
					(*courses)[k] = course
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
			if !strings.HasPrefix(r, "NO REPEATS ALLOWED") {
				cleanRow = append(cleanRow, r)
			}
		}
		if len(cleanRow) > 0 {
			prereqs = append(prereqs, cleanRow)
		}
	}
	
	return prereqs
}
