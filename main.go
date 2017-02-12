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
	"net/http"
	"regexp"
)

const DegreeWorksURL = "https://www.reg.uci.edu/dgw/IRISLink.cgi"

func fetchStudentID(cookie string) (string, error) {
	body := "SERVICE=SCRIPTER&SCRIPT=SD2STUCON"
	statusCode, responseHTML, err := helpers.Post(DegreeWorksURL, cookie, body)
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

func fetchXML(studentID string, cookie string) (string, error) {
	body := fmt.Sprintf("SERVICE=SCRIPTER&REPORT=WEB31&SCRIPT=SD2GETAUD%%26ContentType%%3Dxml&ACTION=REVAUDIT&ContentType=xml&STUID=%v&DEBUG=OFF", studentID)
	statusCode, responseXML, err := helpers.Post(DegreeWorksURL, cookie, body)
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
	courses, blocks, prereqDepts := parsers.Parse(doc)
	term, deptOptions, _ := parsers.DepartmentOptions()
	for dept, _ := range prereqDepts {
		option := deptOptions[dept]
		parsers.ParsePrerequisites(term, option, &courses)
	}
	
	fmt.Println("Cleared Classes:")
	canTake := make([]types.Course, 0)
	for key, block := range blocks {
		// TODO: What happens with double-majors or double-minors? Wouldn’t MAJOR/MINOR keys overwrite?
		fmt.Println(key)
		for _, req := range block.Requirements {
			fmt.Printf("- %d classes remaining in:\n", req.Required)
			for _, option := range req.Options {
				course := courses[option]
				if course.ClearedPrereqs(courses, block.Taken) {
					fmt.Printf("  %v %v\n", course.Department, course.Number)
					canTake = append(canTake, course)
				}
				// fmt.Printf("    %v: %v\n", course.Title)
			}
		}
	}
	
	// jsonInterface := make(map[string]interface{})
	// jsonInterface["courses"] = courses
	// jsonInterface["blocks"] = blocks
	// exportJSON, err := json.Marshal(jsonInterface)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(exportJSON))
}

// func readFromFile(fileName string) {
// 	doc := etree.NewDocument()
// 	doc.ReadSettings.CharsetReader = charset.NewReaderLabel
// 	if err := doc.ReadFromFile(fileName); err != nil {
// 		panic(err)
// 	}
// 	parse(doc)
// }

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
		responseXML, err := fetchXML(*studentIDptr, *cookiePtr)
		if err != nil {
			panic(err)
		}
		readFromString(responseXML)
	} else {
		fmt.Println("No flags were specified. Use `-h` or `--help` flags to get help.")
		// readFromFile("DGW_Report.xsl")
	}
}

