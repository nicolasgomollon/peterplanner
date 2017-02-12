package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/beevik/etree"
	"github.com/nicolasgomollon/peterplanner/helpers"
	"golang.org/x/net/html/charset"
	"net/http"
	"regexp"
	"strconv"
)

const DegreeWorksURL = "https://www.reg.uci.edu/dgw/IRISLink.cgi"
// const FullBody = "SERVICE=SCRIPTER&REPORT=WEB31&SCRIPT=SD2GETAUD%26ContentType%3Dxml&USERID=33897573&USERCLASS=STU&BROWSER=NOT-NAV4&ACTION=REVAUDIT&AUDITTYPE=&DEGREETERM=ACTV&INTNOTES=&INPROGRESS=N&CUTOFFTERM=ACTV&REFRESHBRDG=N&AUDITID=&JSERRORCALL=SetError&NOTENUM=&NOTETEXT=&NOTEMODE=&PENDING=&INTERNAL=&RELOADSEP=TRUE&PRELOADEDPLAN=&ContentType=xml&STUID=33897573&SCHOOL=U&STUSCH=U&DEGREE=BS&STUDEG=BS&STUDEGLIT=B.S.&STUDI=&STULVL=Senior&STUMAJLIT=Software+Engineering&STUCATYEAR=&CLASSES=&DEBUG=OFF"

func fetchStudentID(cookie string) (string, error) {
	body := "SERVICE=SCRIPTER&SCRIPT=SD2STUCON"
	statusCode, responseHTML, err := server.Post(DegreeWorksURL, cookie, body)
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
	studentID := r.FindStringSubmatch(responseHTML)[1]
	return studentID, nil
}

func fetchXML(studentID string, cookie string) (string, error) {
	body := fmt.Sprintf("SERVICE=SCRIPTER&REPORT=WEB31&SCRIPT=SD2GETAUD%%26ContentType%%3Dxml&ACTION=REVAUDIT&ContentType=xml&STUID=%v&DEBUG=OFF", studentID)
	statusCode, responseXML, err := server.Post(DegreeWorksURL, cookie, body)
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

func parse(doc *etree.Document) {
	root := doc.SelectElement("Report").SelectElement("Audit")
	for _, block := range root.SelectElements("Block") {
		reqType := block.SelectAttrValue("Req_type", "unknown")
		fmt.Println(reqType)
		if (reqType == "MAJOR") || (reqType == "MINOR") {
			remaining := 0
			for _, rule := range block.SelectElements("Rule") {
				label := rule.SelectAttrValue("Label", "unknown")
				perComplete, _ := strconv.Atoi(rule.SelectAttrValue("Per_complete", "0"))
				inProg := (rule.SelectAttrValue("In_prog_incomplete", "No") == "Yes")
				if (perComplete < 100) && !inProg {
					fmt.Println("-", label)
					advice := rule.SelectElement("Advice")
					classes, _ := strconv.Atoi(advice.SelectAttrValue("Classes", "0"))
					remaining += classes
					fmt.Printf("    %d classes remaining in:\n", classes)
					for _, course := range advice.SelectElements("Course") {
						cDept := course.SelectAttrValue("Disc", "DEPT")
						cNum := course.SelectAttrValue("Num", "0")
						cTitle := course.SelectAttrValue("Title", "")
						if cTitle != "" {
							fmt.Printf("      %s %s: %s\n", cDept, cNum, cTitle)
						} else {
							fmt.Printf("      %s %s\n", cDept, cNum)
						}
					}
				}
			}
			fmt.Printf("  %d REQUIRED COURSES REMAIN\n", remaining)
		}
		// if title := block.SelectElement("title"); title != nil {
		// 	lang := title.SelectAttrValue("lang", "unknown")
		// 	fmt.Printf("  TITLE: %s (%s)\n", title.Text(), lang)
		// }
		// for _, attr := range block.Attr {
		// 	fmt.Printf("  ATTR: %s=%s\n", attr.Key, attr.Value)
		// }
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
