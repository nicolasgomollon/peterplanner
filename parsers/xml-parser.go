package parsers

import (
	"github.com/beevik/etree"
	"github.com/nicolasgomollon/peterplanner/types"
	"strconv"
	"strings"
)

/* DegreeWorks XML Parser */

func Parse(doc *etree.Document) (map[string]types.Course, map[string]types.Block, map[string]bool) {
	courses := make(map[string]types.Course, 0)
	blocks := make(map[string]types.Block, 0)
	prereqDepts := make(map[string]bool, 0)
	root := doc.SelectElement("Report").SelectElement("Audit")
	for _, block := range root.SelectElements("Block") {
		reqType := block.SelectAttrValue("Req_type", "unknown")
		if (reqType == "MAJOR") || (reqType == "MINOR") {
			blocks[reqType] = parseBlock(block, &courses, &prereqDepts)
		}
	}
	return courses, blocks, prereqDepts
}

func parseBlock(block *etree.Element, courses *map[string]types.Course, prereqDepts *map[string]bool) types.Block {
	taken := make(map[string]bool, 0)
	requirements := make([]types.Requirement, 0)
	for _, rule := range block.SelectElements("Rule") {
		rules := rule.SelectElements("Rule")
		if len(rules) > 0 {
			// TODO: Need to determine how to handle this case, because only certain rule blocks may be necessary.
			for _, r := range rules {
				parseRule(r, courses, prereqDepts, &taken, &requirements)
			}
		} else {
			parseRule(rule, courses, prereqDepts, &taken, &requirements)
		}
	}
	return types.Block{Taken: taken, Requirements: requirements}
}

func parseRule(rule *etree.Element, courses *map[string]types.Course, prereqDepts *map[string]bool, taken *map[string]bool, requirements *[]types.Requirement) {
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
