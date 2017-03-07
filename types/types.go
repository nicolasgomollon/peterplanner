package types

import (
	"math"
	"regexp"
	"strings"
	"time"
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

type Time struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func (t Time) Duration() time.Duration {
	return t.End.Sub(t.Start)
}

func ParseTime(cTime string) Time {
	cStart := strings.TrimSpace(cTime[0:5])
	cEnd := strings.TrimSpace(cTime[6:11])
	pm := (len(cTime) == 12)
	
	format := "3:04"
	start, _ := time.Parse(format, cStart)
	end, _ := time.Parse(format, cEnd)
	
	if (end.Hour() < 12) && pm {
		end = end.Add(time.Hour * time.Duration(12))
	}
	
	if end.Sub(start).Hours() >= 12.0 {
		start = start.Add(time.Hour * time.Duration(12))
	}
	
	return Time{Start: start, End: end}
}

func ParseDays(cDays string) []time.Weekday {
	r, _ := regexp.Compile(`[A-Z][a-z]?`)
	rawDays := r.FindAllStringSubmatch(cDays, -1)
	
	days := make([]time.Weekday, len(rawDays))
	for i, rawDay := range rawDays {
		switch rawDay[0] {
		case "Su":
			days[i] = time.Sunday
		case "M":
			days[i] = time.Monday
		case "Tu":
			days[i] = time.Tuesday
		case "W":
			days[i] = time.Wednesday
		case "Th":
			days[i] = time.Thursday
		case "F":
			days[i] = time.Friday
		case "Sa":
			days[i] = time.Saturday
		default:
			break
		}
	}
	
	return days
}

type Class struct {
	Code       string         `json:"code"`
	Type       string         `json:"type"`
	Section    string         `json:"section"`
	Instructor string         `json:"instructor"`
	Days       []time.Weekday `json:"days"`
	Time       Time           `json:"time"`
	Place      string         `json:"place"`
}

type Course struct {
	Department    string             `json:"department"`
	Number        string             `json:"number"`
	Title         string             `json:"title"`
	Grade         string             `json:"grade"`
	Prerequisites [][]string         `json:"prerequisites"`
	Classes       map[string][]Class `json:"classes"`
}

func (course Course) Key() string {
	return strings.Replace(strings.ToUpper(course.Department + course.Number), " ", "", -1)
}

func (course Course) ClearedPrereqs(student *Student) bool {
	for _, prereqsAND := range course.Prerequisites {
		satisfied := false
		for _, prereqOR := range prereqsAND {
			if !satisfied {
				splitPrrq := strings.Split(prereqOR, "|")
				prereq := strings.Replace(splitPrrq[0], " ", "", -1)
				satisfied = (*student).Taken[prereq]
				if !satisfied && strings.HasPrefix(splitPrrq[0], "NO ") {
					prereq = strings.TrimPrefix(prereq, "NO")
					satisfied = !(*student).Taken[prereq]
				} else if satisfied && (len(splitPrrq) == 2) {
					c := (*student).Courses[prereq]
					grade := splitPrrq[1]
					if (len(c.Grade) != 0) && (len(grade) != 0) {
						if cmpGrade(c.Grade, grade) > 0 {
							satisfied = false
						}
					}
				}
				//
				// ALSO, WHAT TO DO WITH ITEMS LIKE "LOWER DIVISION WRITING"?
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
			// TODO: Compile entire list of missing prerequisites.
			// fmt.Printf("  %v %v -- ", course.Department, course.Number)
			// fmt.Printf("[\"%s\"]\n", strings.Join(prereqsAND, "\", \""))
			return false
		}
	}
	return true
}

type Requirement struct {
	Required  int      `json:"required"`
	Options   []string `json:"options"`
	Completed []string `json:"completed"`
}

func (req Requirement) IsCompleted() bool {
	return len(req.Completed) >= req.Required
}

type Rule struct {
	Label        string        `json:"label"`
	Required     int           `json:"required"`
	Requirements []Requirement `json:"requirements"`
}

func (rule Rule) IsCompleted(student *Student) bool {
	completedCount := 0
	for _, req := range rule.Requirements {
		if req.IsCompleted() {
			completedCount++
			if completedCount >= rule.Required {
				return true
			}
		}
	}
	return false
}

type Block struct {
	ReqType string `json:"type"`
	Title   string `json:"title"`
	Rules   []Rule `json:"rules"`
}

type Student struct {
	StudentID       string            `json:"studentID"`
	Name            string            `json:"name"`
	Email           string            `json:"email"`
	GPA             float64           `json:"gpa"`
	PercentComplete float64           `json:"percentComplete"`
	CreditsApplied  float64           `json:"creditsApplied"`
	Courses         map[string]Course `json:"courses"`
	Taken           map[string]bool   `json:"taken"`
	Blocks          []Block           `json:"blocks"`
	Terms           []string          `json:"terms"`
}

func (student Student) ClassLevel() string {
	credits := student.CreditsApplied
	switch {
	case (135.0 <= credits):
		return "SENIOR"
	case (90.0 <= credits) && (credits < 135.0):
		return "JUNIOR"
	case (45.0 <= credits) && (credits < 90.0):
		return "SOPHOMORE"
	case (0.0 <= credits) && (credits < 45.0):
		return "FRESHMAN"
	}
	return ""
}

func (student Student) Standing() string {
	credits := student.CreditsApplied
	switch {
	case (90.0 <= credits):
		return "UPPER DIVISION"
	case (0.0 <= credits) && (credits < 90.0):
		return "LOWER DIVISION"
	}
	return ""
}
