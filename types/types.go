package types

import (
	"fmt"
	"math"
	"strings"
)

func cmpGrade(grA string, grB string) int64 {
	if (len(grA) == 0) && (len(grB) == 0) {
		return 0
	} else if len(grA) == 0 {
		return math.MinInt64
	} else if len(grB) == 0 {
		return math.MaxInt64
	}
	
	cmpGrA := grA
	cmpGrA = strings.Replace(cmpGrA, "+", "", -1)
	cmpGrA = strings.Replace(cmpGrA, "-", "", -1)
	cmpGrA = strings.ToUpper(cmpGrA)
	
	cmpGrB := grB
	cmpGrB = strings.Replace(cmpGrB, "+", "", -1)
	cmpGrB = strings.Replace(cmpGrB, "-", "", -1)
	cmpGrB = strings.ToUpper(cmpGrB)
	
	var comparison int64 = 0
	if cmpGrA == cmpGrB {
		grAplus := strings.Contains(grA, "+")
		grAminus := strings.Contains(grA, "-")
		
		grBplus := strings.Contains(grB, "+")
		grBminus := strings.Contains(grB, "-")
		
		if grAplus && !grBplus {
			comparison = -1
		} else if !grAplus && grBplus {
			comparison = 1
		} else if !grAminus && grBminus {
			comparison = -1
		} else if grAminus && !grBminus {
			comparison = 1
		}
	} else if cmpGrA < cmpGrB {
		comparison = -1
	} else if cmpGrA > cmpGrB {
		comparison = 1
	}
	
	return comparison
}

type Course struct {
	Department    string     `json:"department"`
	Number        string     `json:"number"`
	Title         string     `json:"title"`
	Grade         string     `json:"grade"`
	Prerequisites [][]string `json:"prerequisites"`
}

func (course Course) Key() string {
	return strings.Replace(strings.ToUpper(course.Department + course.Number), " ", "", -1)
}

func (course Course) ClearedPrereqs(courses map[string]Course, taken map[string]bool) bool {
	for _, prereqsAND := range course.Prerequisites {
		satisfied := false
		for _, prereqOR := range prereqsAND {
			if !satisfied {
				splitPrrq := strings.Split(prereqOR, "|")
				prereq := strings.Replace(splitPrrq[0], " ", "", -1)
				satisfied = taken[prereq]
				if !satisfied && strings.HasPrefix(splitPrrq[0], "NO ") {
					prereq = strings.TrimPrefix(prereq, "NO")
					satisfied = !taken[prereq]
				} else if satisfied && (len(splitPrrq) == 2) {
					c := courses[prereq]
					grade := splitPrrq[1]
					if (len(c.Grade) != 0) && (len(grade) != 0) {
						if cmpGrade(c.Grade, grade) > 0 {
							satisfied = false
						}
					}
				}
				// ALSO, WHAT TO DO WITH ITEMS LIKE "LOWER DIVISION WRITING" OR "UPPER DIVISION STANDING ONLY"
				//
				// UPPER DIVISION STANDING ONLY:
				//   Only students within the junior and senior class levels
				//   (90 or more units) are eligible for enrollment.
				//
				// SENIOR STANDING ONLY:
				//   Only students within the Senior class level
				//   (135 or more units) are eligible for enrollment.
				//
				// JUNIOR STANDING ONLY:
				//   Only students within the Junior class level
				//   (90.0 through 134.9 units) are eligible for enrollment.
				//
				// information from:
				// https://www.reg.uci.edu/enrollment/restrict_codes.html
				//
				if satisfied {
					break
				}
			}
		}
		if !satisfied {
			fmt.Printf("  %v %v -- ", course.Department, course.Number)
			fmt.Printf("[\"%s\"]\n", strings.Join(prereqsAND, "\", \""))
			return false
		}
	}
	return true
}

type Requirement struct {
	Required int      `json:"required"`
	Options  []string `json:"options"`
}

type Block struct {
	Taken        map[string]bool `json:"taken"`
	Requirements []Requirement   `json:"requirements"`
}
