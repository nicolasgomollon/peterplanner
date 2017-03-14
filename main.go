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

package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/beevik/etree"
	"github.com/nicolasgomollon/peterplanner/database"
	"github.com/nicolasgomollon/peterplanner/helpers"
	"github.com/nicolasgomollon/peterplanner/parsers"
	"github.com/nicolasgomollon/peterplanner/types"
	"golang.org/x/net/html/charset"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
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

func readFromString(contentsXML string, outputJSON bool) {
	doc := etree.NewDocument()
	doc.ReadSettings.CharsetReader = charset.NewReaderLabel
	if err := doc.ReadFromString(contentsXML); err != nil {
		panic(err)
	}
	parse(doc, outputJSON)
}

func readFromFile(fileName string, outputJSON bool) {
	doc := etree.NewDocument()
	doc.ReadSettings.CharsetReader = charset.NewReaderLabel
	if err := doc.ReadFromFile(fileName); err != nil {
		panic(err)
	}
	parse(doc, outputJSON)
}

func GetCatalogue() (types.Catalogue, error) {
	b, err := ioutil.ReadFile("/var/www/registrar/catalogue.json")
	if err != nil {
		panic(err)
	}
	var catalogue types.Catalogue
	err = json.Unmarshal(b, &catalogue)
	if err != nil {
		return types.Catalogue{}, err
	}
	return catalogue, nil
}

func parse(doc *etree.Document, outputJSON bool) {
	catalogue, err := GetCatalogue()
	if err != nil {
		panic(err)
	}
	
	student := parsers.Parse(doc, &catalogue)
	student.Terms = catalogue.Terms
	yearTerm := student.Terms[0]
	
	if !outputJSON {
		for _, block := range student.Blocks {
			fmt.Printf("%v: %v\n", block.ReqType, block.Title)
			for _, rule := range block.Rules {
				if rule.IsCompleted(&student) {
					fmt.Printf("✓ %v\n", rule.Label)
					continue
				}
				if (rule.Required == 1) && (len(rule.Requirements) == 1) {
					fmt.Printf("- %v\n", rule.Label)
				} else {
					fmt.Printf("- %v (%v)\n", rule.Label, rule.Required)
				}
				for _, req := range rule.Requirements {
					if req.IsCompleted() {
						continue
					}
					fmt.Printf("    - %d classes remaining in:\n", req.Required)
					for _, option := range req.Options {
						course := student.Courses[option]
						termsOffered := make(map[string][]int, 0)
						for k := range course.Classes {
							t := "--"
							switch {
							case types.IsFQ(k):
								t = "F"
							case types.IsWQ(k):
								t = "W"
							case types.IsSQ(k):
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
						cleared := course.ClearedPrereqs(&student)
						icon := "✗"
						if cleared {
							icon = "✓"
						}
						
						fmt.Printf("        %-35s   offered: %v\n", fmt.Sprintf("%v %v %v: %v", icon, course.Department, course.Number, course.Title), termsOffered)
						if cleared {
							for _, class := range course.Classes[yearTerm] {
								fmt.Printf("            %v %v %v %v\n", class.Code, class.Type, class.Section, class.Instructor)
							}
						} else {
							printArray(course.Prerequisites)
						}
					}
				}
			}
		}
	} else {
		exportJSON, err := json.Marshal(student)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(exportJSON))
	}
}

func printArray(prereqs [][]string) {
	for i, prereqInter := range prereqs {
		spaces := "            "
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
	uidPtr := flag.String("uid", "", "Fetch DegreeWorks XML file for the specified uid.")
	studentIDptr := flag.String("studentID", "", "Fetch DegreeWorks XML file for the specified student ID.")
	cookiePtr := flag.String("cookie", "", "Fetch DegreeWorks XML file using specified cookies.")
	jsonPtr := flag.Bool("json", false, "Output the result in JSON format.")
	flag.Parse()

	if len(*cookiePtr) > 0 {
		if len(*studentIDptr) == 0 {
			studentID, err := fetchStudentID(*cookiePtr)
			if err != nil {
				panic(err)
			}
			*studentIDptr = studentID
		}
		
		db, err := helpers.DatabaseFromFile("/var/www/config/settings.json")
		dbConn, err := database.Connect(db)
		if err != nil {
			panic(errors.New(fmt.Sprintf("ERROR: Could not create new database connection. `%v`.", err.Error())))
		}
		
		studentExists, uid, err := database.RowExists(dbConn, "SELECT `uid` FROM `accounts` WHERE `studentID`=? LIMIT 1", *studentIDptr)
		if !studentExists && (err == nil) {
			timestamp := time.Now().Format(time.RFC3339)
			uid = fmt.Sprintf("%v|%v", *studentIDptr, timestamp)
			h := sha1.New()
			h.Write([]byte(uid))
			bs := h.Sum(nil)
			uid = fmt.Sprintf("%x", bs)
			_, err = database.Execute(dbConn, "INSERT INTO `accounts` SET `uid`=?, `studentID`=?", uid, *studentIDptr)
			if err != nil {
				panic(err)
			}
		}
		
		school, degree, degreeName, studentLevel, studentMajor, err := fetchStudentDetails(*studentIDptr, *cookiePtr)
		if err != nil {
			panic(err)
		}
		
		responseXML, err := fetchXML(*studentIDptr, school, degree, degreeName, studentLevel, studentMajor, *cookiePtr)
		if err != nil {
			panic(err)
		}
		
		filepath := fmt.Sprintf("/var/www/reports/DGW_Report-%v.xsl", *studentIDptr)
		err = ioutil.WriteFile(filepath, []byte(responseXML), 0644)
		if err != nil {
			panic(err)
		}
		
		output := make(map[string]string, 0)
		output["uid"] = uid
		
		exportJSON, err := json.Marshal(output)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(exportJSON))
	} else if len(*uidPtr) > 0 {
		db, err := helpers.DatabaseFromFile("/var/www/config/settings.json")
		dbConn, err := database.Connect(db)
		if err != nil {
			panic(errors.New(fmt.Sprintf("ERROR: Could not create new database connection. `%v`.", err.Error())))
		}
		
		studentExists, studentID, _ := database.RowExists(dbConn, "SELECT `studentID` FROM `accounts` WHERE `uid`=? LIMIT 1", *uidPtr)
		if studentExists {
			readFromFile(fmt.Sprintf("/var/www/reports/DGW_Report-%v.xsl", studentID), *jsonPtr)
		} else {
			fmt.Println("{}")
		}
	} else if len(*studentIDptr) > 0 {
		readFromFile(fmt.Sprintf("/var/www/reports/DGW_Report-%v.xsl", *studentIDptr), *jsonPtr)
	} else {
		fmt.Println("No flags were specified. Use `-h` or `--help` flags to get help.")
	}
}
