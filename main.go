package main

import "fmt"
import "github.com/beevik/etree"
import "strconv"

func main() {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile("DGW_Report.xsl"); err != nil {
		panic(err)
	}
	
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
