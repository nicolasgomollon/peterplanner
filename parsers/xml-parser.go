package parsers

import (
	"fmt"
	"github.com/beevik/etree"
	"github.com/nicolasgomollon/peterplanner/types"
	"strconv"
	"strings"
)

/* DegreeWorks XML Parser */

func Parse(doc *etree.Document, catalogue *types.Catalogue) types.Student {
	courses := make(map[string]types.Course, 0)
	taken := make(map[string]bool, 0)
	enrolled := make(map[string]string, 0)
	blocks := make([]types.Block, 0)
	
	root := doc.SelectElement("Report").SelectElement("Audit")
	audit := root.SelectElement("AuditHeader")
	studentID := audit.SelectAttrValue("Stu_id", "XXXXXXXX")
	name := audit.SelectAttrValue("Stu_name", "ANTEATER, PETER THE")
	email := audit.SelectAttrValue("Stu_email", "PTANTEATER@UCI.EDU")
	student := types.Student{StudentID: studentID, Name: name, Email: email}
	
	activeTerm := ""
	deginfo := root.SelectElement("Deginfo")
	if deginfo != nil {
		degreeData := deginfo.SelectElement("DegreeData")
		if degreeData != nil {
			activeTerm = degreeData.SelectAttrValue("Actv_term", "")
		}
	}
	
	clsinfo := root.SelectElement("Clsinfo")
	if clsinfo != nil {
		for _, class := range clsinfo.SelectElements("Class") {
			cDept := class.SelectAttrValue("Discipline", "DEPT")
			cNum := class.SelectAttrValue("Number", "0")
			cTitle := class.SelectAttrValue("Course_title", "")
			cTerm := class.SelectAttrValue("Term", "")
			cInProgress := class.SelectAttrValue("In_progress", "N")
			key := strings.Replace(strings.ToUpper(cDept + cNum), " ", "", -1)
			if (cInProgress == "Y") && (cTerm > activeTerm) {
				enrolled[key] = cTitle
			}
		}
	}
	
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
			parseProgram(block, catalogue, &courses, &taken, &blocks)
			break
		case "MAJOR", "MINOR":
			parseBlock(block, catalogue, &courses, &taken, &enrolled, &blocks)
			break
		default:
			break
		}
	}
	
	student.Courses = courses
	student.Taken = taken
	student.Blocks = blocks
	
	return student
}

func parseProgram(block *etree.Element, catalogue *types.Catalogue, courses *map[string]types.Course, taken *map[string]bool, blocks *[]types.Block) {
	// TODO: Parse for things like "LOWER DIVISION WRITING" and "UPPER DIVISION WRITING"
}

func parseBlock(block *etree.Element, catalogue *types.Catalogue, courses *map[string]types.Course, taken *map[string]bool, enrolled *map[string]string, blocks *[]types.Block) {
	rules := make([]types.Rule, 0)
	for _, r := range block.SelectElements("Rule") {
		label := r.SelectAttrValue("Label", "")
		rs := r.SelectElements("Rule")
		if len(rs) > 0 {
			req := r.SelectElement("Requirement")
			required, _ := strconv.Atoi(req.SelectAttrValue("NumGroups", "0"))
			rule := types.Rule{Label: label, Required: required}
			requirements := make([]types.Requirement, 0)
			for _, r2 := range rs {
				requirement := parseRule(r2, catalogue, courses, taken, enrolled)
				requirements = append(requirements, requirement)
			}
			rule.Requirements = requirements
			rules = append(rules, rule)
		} else {
			rule := types.Rule{Label: label, Required: 1}
			requirements := make([]types.Requirement, 0)
			requirement := parseRule(r, catalogue, courses, taken, enrolled)
			requirements = append(requirements, requirement)
			rule.Requirements = requirements
			rules = append(rules, rule)
		}
	}
	reqType := block.SelectAttrValue("Req_type", "UNKNOWN")
	title := block.SelectAttrValue("Title", "Untitled")
	theBlock := types.Block{ReqType: reqType, Title: title, Rules: rules}
	*blocks = append(*blocks, theBlock)
}

func parseRule(rule *etree.Element, catalogue *types.Catalogue, courses *map[string]types.Course, taken *map[string]bool, enrolled *map[string]string) types.Requirement {
	requirement := types.Requirement{}
	options := make([]string, 0)
	completed := make([]string, 0)
	
	// Required Classes
	requirementBlock := rule.SelectElement("Requirement")
	if requirementBlock != nil {
		required, _ := strconv.Atoi(requirementBlock.SelectAttrValue("Classes_begin", "0"))
		requirement.Required = required
		for _, course := range requirementBlock.SelectElements("Course") {
			cDept := course.SelectAttrValue("Disc", "DEPT")
			cNum := course.SelectAttrValue("Num", "0")
			
			key := strings.Replace(strings.ToUpper(cDept + cNum), " ", "", -1)
			if c, ok := (*catalogue).Courses[key]; ok {
				(*courses)[key] = c
			} else {
				c := types.Course{Department: cDept, Number: cNum}
				(*courses)[key] = c
			}
			options = append(options, key)
		}
	}
	
	// Taken Classes
	applied := rule.SelectElement("ClassesApplied")
	if applied != nil {
		for _, course := range applied.SelectElements("Class") {
			cDept := course.SelectAttrValue("Discipline", "DEPT")
			cNum := course.SelectAttrValue("Number", "0")
			cGrade := course.SelectAttrValue("Letter_grade", "")
			
			c := types.Course{Department: cDept, Number: cNum}
			key := c.Key()
			if cc, ok := (*courses)[key]; ok {
				c = cc
			} else if cc, ok := (*catalogue).Courses[key]; ok {
				c = cc
			}
			
			if cGrade != "IP" {
				c.Grade = cGrade
				(*taken)[key] = true
				completed = append(completed, key)
			} else if cTitle, ok := (*enrolled)[key]; ok {
				if len(c.ShortTitle) == 0 {
					c.ShortTitle = cTitle
				}
			} else {
				(*taken)[key] = true
				completed = append(completed, key)
			}
			(*courses)[key] = c
		}
	}
	
	requirement.Options = options
	requirement.Completed = completed
	return requirement
}
