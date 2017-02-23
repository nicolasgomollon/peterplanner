package main

import (
	// "encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/beevik/etree"
	"github.com/nicolasgomollon/peterplanner/helpers"
	"github.com/nicolasgomollon/peterplanner/parsers"
	"github.com/nicolasgomollon/peterplanner/types"
	"golang.org/x/net/html/charset"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const DegreeWorksURL = "https://www.reg.uci.edu/dgw/IRISLink.cgi"

func fetchStudentID(cookie string) (string, error) {
	body := "SERVICE=SCRIPTER&SCRIPT=SD2STUCON"
	statusCode, _, responseHTML, err := helpers.Post(DegreeWorksURL, cookie, body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch DegreeWorks HTML file. `%v`.", err.Error()))
	} else if statusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch DegreeWorks HTML file. HTTP status code: %v.", statusCode))
	}
	r, _ := regexp.Compile(`(?s)<input type="hidden" name="STUID" value="(\d*)">`)
	matches := r.FindStringSubmatch(responseHTML)
	if len(matches) == 0 {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch DegreeWorks HTML file. Invalid cookies."))
	}
	studentID := matches[1]
	return studentID, nil
}

func fetchStudentDetails(studentID, cookie string) (string, string, string, string, string, error) {
	body := fmt.Sprintf("SERVICE=SCRIPTER&SCRIPT=SD2STUGID&STUID=%s&DEBUG=OFF", studentID)
	statusCode, _, responseHTML, err := helpers.Post(DegreeWorksURL, cookie, body)
	if err != nil {
		return "", "", "", "", "", errors.New(fmt.Sprintf("ERROR: Unable to fetch DegreeWorks HTML file. `%v`.", err.Error()))
	} else if statusCode != http.StatusOK {
		return "", "", "", "", "", errors.New(fmt.Sprintf("ERROR: Unable to fetch DegreeWorks HTML file. HTTP status code: %v.", statusCode))
	}
	
	r, _ := regexp.Compile(`(?s)<StudentData>(.*)</StudentData>`)
	studentData := r.FindStringSubmatch(responseHTML)[1]
	
	r, _ = regexp.Compile(`(?s)<GoalDtl.*School="(?P<school>.*?)".*Degree="(?P<degree_code>.*?)".*StuLevel="(?P<student_level_code>.*?)".*</GoalDtl>.*<GoalDataDtl.*?GoalCode="MAJOR".*?GoalValue="(?P<major_code>.*?)".*?</GoalDataDtl>`)
	matches := r.FindStringSubmatch(studentData)
	groups := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 {
			groups[name] = matches[i]
		}
	}
	
	r, _ = regexp.Compile(fmt.Sprintf(`(?s)sMajorPicklist\[sMajorPicklist\.length\] = new DataItem\("%s *", "(.*?) *"\);`, groups["major_code"]))
	studentMajor := r.FindStringSubmatch(responseHTML)[1]
	
	r, _ = regexp.Compile(fmt.Sprintf(`(?s)sLevelPicklist\[sLevelPicklist\.length\] = new DataItem\("%s *", "(.*?) *"\);`, groups["student_level_code"]))
	studentLevel := r.FindStringSubmatch(responseHTML)[1]
	
	r, _ = regexp.Compile(fmt.Sprintf(`(?s)sDegreePicklist\[sDegreePicklist\.length\] = new DataItem\("%s *", "(.*?) *"\);`, groups["degree_code"]))
	degreeName := r.FindStringSubmatch(responseHTML)[1]
	
	return groups["school"], groups["degree_code"], degreeName, studentLevel, studentMajor, nil
}

func fetchXML(studentID, school, degree, degreeName, studentLevel, studentMajor, cookie string) (string, error) {
	studentLevel = html.UnescapeString(studentLevel)
	studentLevel = url.QueryEscape(studentLevel)
	studentMajor = html.UnescapeString(studentMajor)
	studentMajor = url.QueryEscape(studentMajor)
	body := fmt.Sprintf("SERVICE=SCRIPTER&REPORT=WEB31&SCRIPT=SD2GETAUD%%26ContentType%%3Dxml&USERID=%s&USERCLASS=STU&BROWSER=NOT-NAV4&ACTION=REVAUDIT&AUDITTYPE&DEGREETERM=ACTV&INTNOTES&INPROGRESS=N&CUTOFFTERM=ACTV&REFRESHBRDG=N&AUDITID&JSERRORCALL=SetError&NOTENUM&NOTETEXT&NOTEMODE&PENDING&INTERNAL&RELOADSEP=TRUE&PRELOADEDPLAN&ContentType=xml&STUID=%s&SCHOOL=%s&STUSCH=%s&DEGREE=%s&STUDEG=%s&STUDEGLIT=%s&STUDI&STULVL=%s&STUMAJLIT=%s&STUCATYEAR&CLASSES&DEBUG=OFF", studentID, studentID, school, school, degree, degree, degreeName, studentLevel, studentMajor)
	statusCode, _, responseXML, err := helpers.Post(DegreeWorksURL, cookie, body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch DegreeWorks XML file. `%v`.", err.Error()))
	} else if statusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch DegreeWorks XML file. HTTP status code: %v.", statusCode))
	}
	return responseXML, nil
}

func fetchBasicXML(studentID string, cookie string) (string, error) {
	body := fmt.Sprintf("SERVICE=SCRIPTER&REPORT=WEB31&SCRIPT=SD2GETAUD%%26ContentType%%3Dxml&ACTION=REVAUDIT&ContentType=xml&STUID=%v&DEBUG=OFF", studentID)
	statusCode, _, responseXML, err := helpers.Post(DegreeWorksURL, cookie, body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch DegreeWorks XML file. `%v`.", err.Error()))
	} else if statusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("ERROR: Unable to fetch DegreeWorks XML file. HTTP status code: %v.", statusCode))
	}
	return responseXML, nil
}

func readFromString(contentsXML string) {
	doc := etree.NewDocument()
	doc.ReadSettings.CharsetReader = charset.NewReaderLabel
	if err := doc.ReadFromString(contentsXML); err != nil {
		panic(err)
	}
	parse(doc)
}

func readFromFile(fileName string) {
	doc := etree.NewDocument()
	doc.ReadSettings.CharsetReader = charset.NewReaderLabel
	if err := doc.ReadFromFile(fileName); err != nil {
		panic(err)
	}
	parse(doc)
}

type File struct {
	Name string
	Term string
}

func schedulesFor(deptPath string, termsMap *map[string]bool) []File {
	r, _ := regexp.Compile(`soc_(\d{4}-\d{2})\.txt`)
	files := make([]File, 0)
	filepath.Walk(deptPath, func(path string, f os.FileInfo, _ error) error {
		filename := f.Name()
		if strings.HasPrefix(filename, "soc_") {
			term := r.FindStringSubmatch(filename)[1]
			(*termsMap)[term] = true
			files = append(files, File{Name: filename, Term: term})
		}
		return nil
	})
	return files
}

func parse(doc *etree.Document) {
	student, prereqDepts := parsers.Parse(doc)
	for dept, _ := range prereqDepts {
		dir := strings.Replace(dept, "/", "_", -1)
		b, err := ioutil.ReadFile(fmt.Sprintf("/var/www/registrar/%v/prereqs.html", dir))
		if err != nil {
			panic(err)
		}
		responseHTML := string(b)
		parsers.ParsePrerequisites(responseHTML, &student.Courses)
	}
	
	// fmt.Println("Cleared Courses:")
	canTake := make([]types.Course, 0)
	toCheck := make(map[string][]string, 0)
	for _, block := range student.Blocks {
		// fmt.Printf("%v: %v\n", block.ReqType, block.Title)
		for _, req := range block.Requirements {
			// fmt.Printf("- %d classes remaining in:\n", req.Required)
			for _, option := range req.Options {
				course := student.Courses[option]
				if course.ClearedPrereqs(student) {
					// fmt.Printf("  %v %v\n", course.Department, course.Number)
					
					courseNums := toCheck[course.Department]
					if courseNums == nil {
						courseNums = make([]string, 0)
					}
					courseNums = append(courseNums, course.Number)
					toCheck[course.Department] = courseNums
					
					canTake = append(canTake, course)
				}
				// fmt.Printf("    %v: %v\n", course.Title)
			}
		}
	}
	
	termsMap := make(map[string]bool, 0)
	// TODO: What to do with `toCheck->courseNums`?
	for dept, _ := range toCheck {
		dir := strings.Replace(dept, "/", "_", -1)
		deptPath := fmt.Sprintf("/var/www/registrar/%v/", dir)
		files := schedulesFor(deptPath, &termsMap)
		for _, file := range files {
			b, err := ioutil.ReadFile(deptPath + file.Name)
			if err != nil {
				panic(err)
			}
			responseTXT := string(b)
			parsers.ParseWebSOC(file.Term, responseTXT, &student.Courses)
		}
	}
	
	terms := make([]string, len(termsMap))
	i := 0
	for t := range termsMap {
		terms[i] = t
		i++
	}
	sort.Sort(sort.Reverse(sort.StringSlice(terms)))
	student.Terms = terms
	yearTerm := terms[0]
	
	for _, block := range student.Blocks {
		fmt.Printf("%v: %v\n", block.ReqType, block.Title)
		for _, req := range block.Requirements {
			fmt.Printf("- %d classes remaining in:\n", req.Required)
			for _, option := range req.Options {
				course := student.Courses[option]
				termsOffered := make(map[string][]int, 0)
				for k := range course.Classes {
					t := "--"
					switch {
					case parsers.IsFQ(k):
						t = "F"
					case parsers.IsWQ(k):
						t = "W"
					case parsers.IsSQ(k):
						t = "S"
					}
					y, _ := strconv.Atoi(k[0:4])
					years := termsOffered[t]
					if years == nil {
						years = make([]int, 0)
					}
					years = append(years, y)
					termsOffered[t] = years
				}
				cleared := course.ClearedPrereqs(student)
				icon := "✗"
				if cleared {
					icon = "✓"
				}
				fmt.Printf("%v %v %v: %v     offered: %v\n", icon, course.Department, course.Number, course.Title, termsOffered)
				if cleared {
					for _, class := range course.Classes[yearTerm] {
						fmt.Printf("    %v %v %v %v\n", class.Code, class.Type, class.Section, class.Instructor)
					}
				} else {
					printArray(course.Prerequisites)
				}
			}
		}
	}
	
	// exportJSON, err := json.Marshal(student)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(exportJSON))
}

func printArray(prereqs [][]string) {
	for i, prereqInter := range prereqs {
		spaces := "    "
		sep := ","
		if i == (len(prereqs) - 1) {
			sep = ""
		}
		
		fmt.Printf("%s[\"", spaces)
		fmt.Printf(strings.Join(prereqInter, "\", \""))
		fmt.Printf("\"]%s\n", sep)
	}
}

func main() {
	studentIDptr := flag.String("studentID", "", "Fetch DegreeWorks XML file for the specified student ID.")
	cookiePtr := flag.String("cookie", "", "Fetch DegreeWorks XML file using specified cookies.")
	flag.Parse()

	if len(*cookiePtr) > 0 {
		if len(*studentIDptr) == 0 {
			studentID, err := fetchStudentID(*cookiePtr)
			if err != nil {
				panic(err)
			}
			*studentIDptr = studentID
		}
		
		school, degree, degreeName, studentLevel, studentMajor, err := fetchStudentDetails(*studentIDptr, *cookiePtr)
		if err != nil {
			panic(err)
		}
		
		responseXML, err := fetchXML(*studentIDptr, school, degree, degreeName, studentLevel, studentMajor, *cookiePtr)
		if err != nil {
			panic(err)
		}
		
		// responseXML, err := fetchBasicXML(*studentIDptr, *cookiePtr)
		// if err != nil {
		// 	panic(err)
		// }
		
		readFromString(responseXML)
	} else if len(*studentIDptr) > 0 {
		readFromFile(fmt.Sprintf("/var/www/reports/DGW_Report-%v.xsl", *studentIDptr))
	} else {
		fmt.Println("No flags were specified. Use `-h` or `--help` flags to get help.")
	}
}
