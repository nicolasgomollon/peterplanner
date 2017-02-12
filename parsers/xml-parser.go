package parsers

import (
	"fmt"
	"github.com/beevik/etree"
	"github.com/nicolasgomollon/peterplanner/types"
	"strconv"
	"strings"
)

/* DegreeWorks XML Parser */

func Parse(doc *etree.Document) (types.Student, map[string]bool) {
	courses := make(map[string]types.Course, 0)
	taken := make(map[string]bool, 0)
	blocks := make([]types.Block, 0)
	prereqDepts := make(map[string]bool, 0)
	
	root := doc.SelectElement("Report").SelectElement("Audit")
	audit := root.SelectElement("AuditHeader")
	studentID := audit.SelectAttrValue("Stu_id", "XXXXXXXX")
	name := audit.SelectAttrValue("Stu_name", "ANTEATER, PETER THE")
	email := audit.SelectAttrValue("Stu_email", "PTANTEATER@UCI.EDU")
	student := types.Student{StudentID: studentID, Name: name, Email: email}
	
	for _, block := range root.SelectElements("Block") {
		reqType := block.SelectAttrValue("Req_type", "unknown")
		switch reqType {
		case "DEGREE":
			gpa, _ := strconv.ParseFloat(block.SelectAttrValue("GPA", "0.0"), 64)
			student.GPA = gpa
			
			percentComplete, _ := strconv.ParseFloat(block.SelectAttrValue("Per_complete", "0.0"), 64)
			student.PercentComplete = percentComplete
			
			creditsApplied, _ := strconv.ParseFloat(block.SelectAttrValue("Credits_applied", "0.0"), 64)
			student.CreditsApplied = creditsApplied
			
			classLevelKey := strings.Replace(fmt.Sprintf("%v STANDING ONLY", student.ClassLevel()), " ", "", -1)
			taken[classLevelKey] = true
			
			standingKey := strings.Replace(fmt.Sprintf("%v STANDING ONLY", student.Standing()), " ", "", -1)
			taken[standingKey] = true
			break
		case "PROGRAM":
			parseProgram(block, &courses, &taken, &blocks, &prereqDepts)
			break
		case "MAJOR", "MINOR":
			parseBlock(block, &courses, &taken, &blocks, &prereqDepts)
			break
		default:
			break
		}
	}
	
	student.Courses = courses
	student.Taken = taken
	student.Blocks = blocks
	
	return student, prereqDepts
}

func parseProgram(block *etree.Element, courses *map[string]types.Course, taken *map[string]bool, blocks *[]types.Block, prereqDepts *map[string]bool) {
	// TODO: Parse for things like "LOWER DIVISION WRITING" and "UPPER DIVISION WRITING"
}

func parseBlock(block *etree.Element, courses *map[string]types.Course, taken *map[string]bool, blocks *[]types.Block, prereqDepts *map[string]bool) {
	requirements := make([]types.Requirement, 0)
	for _, rule := range block.SelectElements("Rule") {
		rules := rule.SelectElements("Rule")
		if len(rules) > 0 {
			// TODO: Need to determine how to handle this case, because only certain rule blocks may be necessary.
			for _, r := range rules {
				parseRule(r, courses, taken, prereqDepts, &requirements)
			}
		} else {
			parseRule(rule, courses, taken, prereqDepts, &requirements)
		}
	}
	reqType := block.SelectAttrValue("Req_type", "UNKNOWN")
	title := block.SelectAttrValue("Title", "Untitled")
	theBlock := types.Block{ReqType: reqType, Title: title, Requirements: requirements}
	*blocks = append(*blocks, theBlock)
}

func parseRule(rule *etree.Element, courses *map[string]types.Course, taken *map[string]bool, prereqDepts *map[string]bool, requirements *[]types.Requirement) {
	// Remaining Classes
	advice := rule.SelectElement("Advice")
	if advice != nil {
		required, _ := strconv.Atoi(advice.SelectAttrValue("Classes", "0"))
		options := make([]string, 0)
		for _, course := range advice.SelectElements("Course") {
			cDept := course.SelectAttrValue("Disc", "DEPT")
			cNum := course.SelectAttrValue("Num", "0")
			cTitle := course.SelectAttrValue("Title", "")
			
			course := types.Course{Department: cDept, Number: cNum, Title: cTitle}
			(*courses)[course.Key()] = course
			(*prereqDepts)[strings.ToUpper(course.Department)] = true
			options = append(options, course.Key())
		}
		requirement := types.Requirement{Required: required, Options: options}
		*requirements = append(*requirements, requirement)
	}
	
	// Taken Classes
	applied := rule.SelectElement("ClassesApplied")
	if applied != nil {
		for _, course := range applied.SelectElements("Class") {
			cDept := course.SelectAttrValue("Discipline", "DEPT")
			cNum := course.SelectAttrValue("Number", "0")
			cGrade := course.SelectAttrValue("Letter_grade", "")
			
			course := types.Course{Department: cDept, Number: cNum, Grade: cGrade}
			(*courses)[course.Key()] = course
			(*taken)[course.Key()] = true
		}
	}
}
