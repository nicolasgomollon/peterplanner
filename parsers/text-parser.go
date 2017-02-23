package parsers

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/nicolasgomollon/peterplanner/helpers"
	"github.com/nicolasgomollon/peterplanner/types"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

/* WebSOC Text Parser */

const WebSocURL = "https://www.reg.uci.edu/perl/WebSoc/"

func IsAcademicTerm(term string) bool {
	if len(term) == 7 {
		switch term[5:] {
		case "92", "03", "14":
			return true
		}
	}
	return false
}

func IsFQ(term string) bool {
	if IsAcademicTerm(term) {
		return term[5:] == "92"
	}
	return false
}

func IsWQ(term string) bool {
	if IsAcademicTerm(term) {
		return term[5:] == "03"
	}
	return false
}

func IsSQ(term string) bool {
	if IsAcademicTerm(term) {
		return term[5:] == "14"
	}
	return false
}

func FallQuarter(year int) string {
	return fmt.Sprintf("%v-92", year)
}

func WinterQuarter(year int) string {
	return fmt.Sprintf("%v-03", year)
}

func SpringQuarter(year int) string {
	return fmt.Sprintf("%v-14", year)
}

func YearFQ() int {
	currentDate := time.Now()
	year := currentDate.Year()
	return year
}

func YearWQ() int {
	currentDate := time.Now()
	year := currentDate.Year()
	month := currentDate.Month()
	if month > time.June {
		year++
	}
	return year
}

func YearSQ() int {
	currentDate := time.Now()
	year := currentDate.Year()
	month := currentDate.Month()
	if month > time.June {
		year++
	}
	return year
}

func AcademicYear() int {
	currentDate := time.Now()
	year := currentDate.Year()
	month := currentDate.Month()
	if month <= time.June {
		year--
	}
	return year
}

func SDepartmentOptions() (string, map[string]string, error) {
	statusCode, responseHTML, err := helpers.Get(WebSocURL)
	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("ERROR: Unable to fetch WebSOC HTML file. `%v`.", err.Error()))
	} else if statusCode != http.StatusOK {
		return "", nil, errors.New(fmt.Sprintf("ERROR: Unable to fetch WebSOC HTML file. HTTP status code: %v.", statusCode))
	}
	r, _ := regexp.Compile(`<option value="(.*?)".*?>`)
	term := r.FindStringSubmatch(responseHTML)[1]
	r, _ = regexp.Compile(`(?s)<select name="Dept">(.*?)</select>`)
	departments := r.FindStringSubmatch(responseHTML)[1]
	r, _ = regexp.Compile(`<option value="(.*?)">`)
	options := r.FindAllStringSubmatch(departments, -1)
	deptOptions := make(map[string]string, 0)
	for i, option := range options {
		if i == 0 {
			continue
		}
		opt := html.UnescapeString(option[1])
		key := strings.Replace(strings.ToUpper(opt), " ", "", -1)
		deptOptions[key] = opt
	}
	return term, deptOptions, nil
}

func FetchWebSOC(yearTerm, dept string, courseNums []string) (string, error) {
	courseNum := strings.Join(courseNums, ",")
	body := fmt.Sprintf("Submit=Display+Text+Results&YearTerm=%s&ShowFinals=on&Breadth=ANY&Dept=%s&CourseNum=%s&Division=ANY&ClassType=ALL&FullCourses=ANY&CancelledCourses=Exclude", yearTerm, url.QueryEscape(dept), url.QueryEscape(courseNum))
	statusCode, contentType, responseTXT, err := helpers.Post(WebSocURL, "", body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch WebSOC TXT file. `%v`.", err.Error()))
	} else if statusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch WebSOC TXT file. HTTP status code: %v.", statusCode))
	}
	switch contentType {
	case "text/plain":
		break
	case "text/html":
		return "", errors.New("ERROR: Unable to fetch WebSOC TXT file. Quarter has not yet started.")
	default:
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch WebSOC TXT file. Received unexpected Content-Type: `%v`.", contentType))
	}
	return responseTXT, nil
}

type Token struct {
	Start int
	End   int
}

func (token Token) Len() int {
	return token.End + token.Start
}

// func (subscription Subscription) IsValid() bool {
// 	currentDate := time.Now()
// 	return (currentDate.After(subscription.ValidFrom.Time()) || currentDate.Equal(subscription.ValidFrom.Time())) && currentDate.Before(subscription.ExpiresOn.Time())
// }

// func (subscription Subscription) JSON() (string, error) {
// 	subscriptionJSON, err := json.Marshal(subscription)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(subscriptionJSON), nil
// }

func ParseWebSOC(responseTXT string, courses *map[string]types.Course, offered *map[string]bool) error {
	scanner := bufio.NewScanner(strings.NewReader(responseTXT))
	shouldParse := false
	for scanner.Scan() {
		line := scanner.Text()
		if (len(line) > 5) && (line[0:5] == "**** ") {
			return errors.New(line[5:])
		} else if line == "       _________________________________________________________________" {
			if !shouldParse {
				shouldParse = true
			} else {
				scanner.Scan() // Consume the empty line.
				break
			}
		}
	}
	if shouldParse {
		width := 0
		tabTkn := Token{}
		ccodeTkn := Token{}
		typTkn := Token{}
		secTkn := Token{}
		untTkn := Token{}
		instTkn := Token{}
		timeTkn := Token{}
		placeTkn := Token{}
		
		cDept := ""
		cNum := ""
		cTitle := ""
		
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) == 0 {
				// Empty line.
				continue
			} else if (len(line) > 4) && (line[0:4] == "*** ") {
				// Reached the end of the readable file.
				break
			} else if (len(line) > 4) && (len(strings.TrimSpace(line[0:4])) > 0) {
				// Course information.
				width = 0
				cDept = strings.TrimSpace(line[0:8])
				cNum = strings.TrimSpace(line[9:18])
				cTitle = line[19:]
				// fmt.Printf("`%v` `%v` `%v`\n", cDept, cNum, cTitle)
				continue
			}
			if width == 0 {
				// First line of data for course.
				// Recalculate all values.
				width = len(line)
				
				tabTkn.Start = 0
				tabTkn.End = strings.Index(line, "CCode")
				
				ccodeTkn.Start = tabTkn.End
				ccodeTkn.End = strings.Index(line, "Typ") - 1
				
				typTkn.Start = ccodeTkn.End + 1
				typTkn.End = strings.Index(line, "Sec") - 1
				
				secTkn.Start = typTkn.End + 1
				secTkn.End = strings.Index(line, "Unt") - 1
				
				untTkn.Start = secTkn.End + 1
				untTkn.End = strings.Index(line, "Instructor") - 1
				
				instTkn.Start = untTkn.End + 1
				instTkn.End = strings.Index(line, "Time") - 1
				
				timeTkn.Start = instTkn.End + 1
				timeTkn.End = strings.Index(line, "Place") - 1
				
				placeTkn.Start = timeTkn.End + 1
				placeTkn.End = strings.Index(line, "Final") - 1
				continue
			} else if len(line) < width {
				// Handle `~ Same as 34030 (CompSci 113, Lec A).`
				continue
			}
			cCode := line[ccodeTkn.Start:ccodeTkn.End]
			if len(strings.TrimSpace(cCode)) == 0 {
				// Extra instructors.
				continue
			}
			class := types.Class{}
			class.Code = cCode
			class.Type = line[typTkn.Start:typTkn.End]
			class.Section = strings.TrimSpace(line[secTkn.Start:secTkn.End])
			class.Instructor = strings.TrimSpace(line[instTkn.Start:instTkn.End])
			cTimeRaw := line[timeTkn.Start:timeTkn.End]
			r, _ := regexp.Compile(`([A-z]*)\s+((?: \d|\d{2}):\d{2}-(?: \d|\d{2}):\d{2}p?)`)
			cTimeParts := r.FindStringSubmatch(cTimeRaw)
			if len(cTimeParts) == 3 {
				class.Days = types.ParseDays(cTimeParts[1])
				class.Time = types.ParseTime(cTimeParts[2])
			}
			class.Place = strings.TrimSpace(line[placeTkn.Start:placeTkn.End])
			//fmt.Printf("`%v` `%v` `%v` `%v` `%v` `%v` `%v`\n", class.Code, class.Type, class.Section, class.Instructor, class.Days, class.Time, class.Place)
			
			k := strings.Replace(strings.ToUpper(cDept + cNum), " ", "", -1)
			if course, ok := (*courses)[k]; ok {
				if len(course.Title) == 0 {
					course.Title = cTitle
				}
				classes := course.Classes
				if classes == nil {
					classes = make([]types.Class, 0)
				}
				classes = append(classes, class)
				course.Classes = classes
				(*courses)[k] = course
				(*offered)[k] = true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
