package helpers

import (
	"github.com/beevik/etree"
	"strconv"
	"strings"
)

type Course struct {
	Department    string   `json:"department"`
	Number        string   `json:"number"`
	Title         string   `json:"title"`
	Grade         string   `json:"grade"`
	Prerequisites []Course `json:"prerequisites"`
}

func (course Course) Key() (string, error) {
	return strings.Replace(strings.ToUpper(course.Department + course.Number), " ", "", -1)
}

type Requirement struct {
	Required int      `json:"required"`
	Options  []Course `json:"options"`
}

type Block struct {
	Taken        map[string]Course `json:"taken"`
	Requirements []Requirement     `json:"requirements"`
}

func Parse(doc *etree.Document) map[string]Block {
	blocks := make(map[string]Block, 0)
	root := doc.SelectElement("Report").SelectElement("Audit")
	for _, block := range root.SelectElements("Block") {
		reqType := block.SelectAttrValue("Req_type", "unknown")
		if (reqType == "MAJOR") || (reqType == "MINOR") {
			blocks[reqType] = parseBlock(block)
		}
	}
	return blocks
}

func parseBlock(block *etree.Element) Block {
	taken := make(map[string]Course, 0)
	requirements := make([]Requirement, 0)
	for _, rule := range block.SelectElements("Rule") {
		applied := rule.SelectElement("ClassesApplied")
		advice := rule.SelectElement("Advice")
		if advice != nil {
			// Remaining Classes
			required, _ := strconv.Atoi(advice.SelectAttrValue("Classes", "0"))
			options := make([]Course, 0)
			for _, course := range advice.SelectElements("Course") {
				cDept := course.SelectAttrValue("Disc", "DEPT")
				cNum := course.SelectAttrValue("Num", "0")
				cTitle := course.SelectAttrValue("Title", "")
				
				course := Course{Department: cDept, Number: cNum, Title: cTitle}
				options = append(options, course)
			}
			requirement := Requirement{Required: required, Options: options}
			requirements = append(requirements, requirement)
		} else if applied != nil {
			// Taken Classes
			for _, course := range applied.SelectElements("Class") {
				cDept := course.SelectAttrValue("Discipline", "DEPT")
				cNum := course.SelectAttrValue("Number", "0")
				cGrade := course.SelectAttrValue("Letter_grade", "")
				
				course := Course{Department: cDept, Number: cNum, Grade: cGrade}
				taken[course.Key()] = course
			}
		}
	}
	return Block{Taken: taken, Requirements: requirements}
}
