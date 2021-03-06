package goquery

import (
	"bytes"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var rxClassTrim = regexp.MustCompile("[\t\r\n]")

// Experimental func, input[type=text|hidden|radio|checkbox], textarea, select
func (s *Selection) Val() (val string) {
	if len(s.Nodes) == 0 {
		return
	}
	node := s.Filter("input[type #=(text|hidden)], input[type #=(radio|checkbox)][checked], textarea, select").Last()
	if len(node.Nodes) == 0 {
		return
	}
	node_type := strings.ToLower(node.Nodes[0].Data)
	if node_type == "input" {
		switch node.AttrOr("type", "text") {
			case "checkbox", "radio":
				return node.AttrOr("value", "on")
			case "text", "hidden":
				return node.AttrOr("value", "")
		}
	} else if node_type == "select" {
		node_option := node.Find("option[selected]")
		if len(node_option.Nodes) == 0 {
			node_option = node.Find("option").First()
			if len(node_option.Nodes) == 0 {
				return
			}
		}
		if val, ok := node_option.Attr("value"); ok {
			return val
		}
		val, _ = node_option.Html()
		return val
	} else if node_type == "textarea" {
		val, _ = node.Html()
		return val
	}
	return
}

// Experimental func, Get mime-type for object, param, embed tags
func (s *Selection) GetMimeType() (val string) {
	if len(s.Nodes) == 0 {
		return
	}
	node_type := strings.ToLower(s.Nodes[0].Data)
	if node_type == "object" {
		val = strings.Trim(s.AttrOr("codetype", ""), " \n\t\r")
		if len(val) == 0 {
			val = strings.Trim(s.Find("param[type #= (.*)]").AttrOr("type", ""), " \n\t\r")
		}
		if len(val) == 0 {
			val = strings.Trim(s.Find("embed[type #= (.*)]").AttrOr("type", ""), " \n\t\r")
		}
	} else if node_type == "embed" {
		val = strings.Trim(s.AttrOr("type", ""), " \n\t\r")
	} else if node_type == "param" {
		val = strings.Trim(s.AttrOr("type", ""), " \n\t\r")
	}
	val = strings.ToLower(val)
	return
}

// Experimental func, Get address file for object, param, embed tags
func (s *Selection) GetObjectSrc() (val string) {
	if len(s.Nodes) == 0 {
		return
	}
	node_type := strings.ToLower(s.Nodes[0].Data)
	if node_type == "object" {
		val = strings.Trim(s.AttrOr("data", ""), " \n\t\r")
		if len(val) == 0 {
			val = strings.Trim(s.Find("param[name=movie]").AttrOr("value", ""), " \n\t\r")
		}
		if len(val) == 0 {
			val = strings.Trim(s.Find("embed[src #= (.*)]").AttrOr("src", ""), " \n\t\r")
		}
	} else if node_type == "embed" {
		val = strings.Trim(s.AttrOr("src", ""), " \n\t\r")
	} else if node_type == "param" {
		val = strings.Trim(s.Filter("param[name=movie]").AttrOr("value", ""), " \n\t\r")
	} else if node_type == "iframe" {
		val = strings.Trim(s.AttrOr("src", ""), " \n\t\r")
	}
	val = strings.ToLower(val)
	return
}

// Return root Document for Selection
func (s *Selection) GetDocument() *Document {
    return s.document
}

// Return Selection of root Document
func (s *Selection) GetDocumentSelection() *Selection {
    return s.document.Clone()
}


// Attr gets the specified attribute's value for the first element in the
// Selection. To get the value for each element individually, use a looping
// construct such as Each or Map method.
func (s *Selection) Attr(attrName string) (val string, exists bool) {
	if len(s.Nodes) == 0 {
		return
	}
	attrName = strings.ToLower(attrName)
	return getAttributeValue(attrName, s.Nodes[0])
}

// AttrOr works like Attr but returns default value if attribute is not present.
func (s *Selection) AttrOr(attrName, defaultValue string) string {
	if len(s.Nodes) == 0 {
		return defaultValue
	}

	attrName = strings.ToLower(attrName)
	val, exists := getAttributeValue(attrName, s.Nodes[0])
	if !exists {
		return defaultValue
	}

	return val
}

// RemoveAttr removes the named attribute from each element in the set of matched elements.
func (s *Selection) RemoveAttr(attrName string) *Selection {
	attrName = strings.ToLower(attrName)
	attrNames := strings.Split(attrName, " ")

	for _, attr := range attrNames {
		if attr == "" {
			continue
		}
		for _, n := range s.Nodes {
			removeAttr(n, attr)
		}
	}

	return s
}

// Find and Remove attr - находит и удаляет объявленные атрибуты в не зависимости от содержания
func (s *Selection) FindRemoveAttr(attrName string) *Selection {
	attrName = strings.ToLower(attrName)
	var attrSelect string

	for _, attr := range strings.Split(attrName, " ") {
		if attr == "" {
			continue
		}
		attrSelect += ", ["+attr+" #= (.*)]"
	}
	if len(attrSelect) > 0 {
		return s.Find(attrSelect[2:]).RemoveAttr(attrName)
	}

	return s
}

// SetAttr sets the given attribute on each element in the set of matched elements.
func (s *Selection) SetAttr(attrName, val string) *Selection {
	attrName = strings.ToLower(attrName)
	for _, n := range s.Nodes {
		attr := getAttributePtr(attrName, n)
		if attr == nil {
			n.Attr = append(n.Attr, html.Attribute{Key: attrName, Val: val})
		} else {
			attr.Val = val
		}
	}

	return s
}

// Text gets the combined text contents of each element in the set of matched
// elements, including their descendants.
func (s *Selection) Text(params ...string) string {
	var buf bytes.Buffer
	var sep = ""
	if len(params) > 0 {
		sep = params[0]
	}

	// Slightly optimized vs calling Each: no single selection object created
	for _, n := range s.Nodes {
		buf.WriteString(getNodeText(n, sep))
	}
	return buf.String()
}


// Size is an alias for Length.
func (s *Selection) Size() int {
	return s.Length()
}

// Length returns the number of elements in the Selection object.
func (s *Selection) Length() int {
	return len(s.Nodes)
}

// Html gets the HTML contents of the first element in the set of matched
// elements. It includes text and comment nodes.
func (s *Selection) Html() (ret string, e error) {
	// Since there is no .innerHtml, the HTML content must be re-created from
	// the nodes using html.Render.
	var buf bytes.Buffer

	if len(s.Nodes) > 0 {
		for c := s.Nodes[0].FirstChild; c != nil; c = c.NextSibling {
			e = html.Render(&buf, c)
			if e != nil {
				return
			}
		}
		ret = buf.String()
	}

	return
}

func (s *Selection) NodeName() string {
	return NodeName(s)
}

func (s *Selection) OuterHtml() (string, error) {
	return OuterHtml(s)
}

// AddClass adds the given class(es) to each element in the set of matched elements.
// Multiple class names can be specified, separated by a space or via multiple arguments.
func (s *Selection) AddClass(class ...string) *Selection {
	classStr := strings.TrimSpace(strings.Join(class, " "))

	if classStr == "" {
		return s
	}

	tcls := getClassesSlice(classStr)
	for _, n := range s.Nodes {
		curClasses, attr := getClassesAndAttr(n, true)
		for _, newClass := range tcls {
			if strings.Index(curClasses, " "+newClass+" ") == -1 {
				curClasses += newClass + " "
			}
		}

		setClasses(n, attr, curClasses)
	}

	return s
}

// HasClass determines whether any of the matched elements are assigned the
// given class.
func (s *Selection) HasClass(class string) bool {
	class = " " + class + " "
	for _, n := range s.Nodes {
		classes, _ := getClassesAndAttr(n, false)
		if strings.Index(classes, class) > -1 {
			return true
		}
	}
	return false
}

// RemoveClass removes the given class(es) from each element in the set of matched elements.
// Multiple class names can be specified, separated by a space or via multiple arguments.
// If no class name is provided, all classes are removed.
func (s *Selection) RemoveClass(class ...string) *Selection {
	var rclasses []string

	classStr := strings.TrimSpace(strings.Join(class, " "))
	remove := classStr == ""

	if !remove {
		rclasses = getClassesSlice(classStr)
	}

	for _, n := range s.Nodes {
		if remove {
			removeAttr(n, "class")
		} else {
			classes, attr := getClassesAndAttr(n, true)
			for _, rcl := range rclasses {
				classes = strings.Replace(classes, " "+rcl+" ", " ", -1)
			}

			setClasses(n, attr, classes)
		}
	}

	return s
}

// ToggleClass adds or removes the given class(es) for each element in the set of matched elements.
// Multiple class names can be specified, separated by a space or via multiple arguments.
func (s *Selection) ToggleClass(class ...string) *Selection {
	classStr := strings.TrimSpace(strings.Join(class, " "))

	if classStr == "" {
		return s
	}

	tcls := getClassesSlice(classStr)

	for _, n := range s.Nodes {
		classes, attr := getClassesAndAttr(n, true)
		for _, tcl := range tcls {
			if strings.Index(classes, " "+tcl+" ") != -1 {
				classes = strings.Replace(classes, " "+tcl+" ", " ", -1)
			} else {
				classes += tcl + " "
			}
		}

		setClasses(n, attr, classes)
	}

	return s
}

// Get the specified node's text content.
func getNodeText(node *html.Node, sep string) string {
	if node.Type == html.TextNode {
		// Keep newlines and spaces, like jQuery
		return sep+node.Data+sep
	} else if node.FirstChild != nil {
		var buf bytes.Buffer
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			buf.WriteString(getNodeText(c, sep))
		}
		return buf.String()
	}

	return ""
}

func getAttributePtr(attrName string, n *html.Node) *html.Attribute {
	if n == nil {
		return nil
	}

	for i, a := range n.Attr {
		if a.Key == attrName {
			return &n.Attr[i]
		}
	}
	return nil
}

// Private function to get the specified attribute's value from a node.
func getAttributeValue(attrName string, n *html.Node) (val string, exists bool) {
	if a := getAttributePtr(attrName, n); a != nil {
		val = a.Val
		exists = true
	}
	return
}

// Get and normalize the "class" attribute from the node.
func getClassesAndAttr(n *html.Node, create bool) (classes string, attr *html.Attribute) {
	// Applies only to element nodes
	if n.Type == html.ElementNode {
		attr = getAttributePtr("class", n)
		if attr == nil && create {
			n.Attr = append(n.Attr, html.Attribute{
				Key: "class",
				Val: "",
			})
			attr = &n.Attr[len(n.Attr)-1]
		}
	}

	if attr == nil {
		classes = " "
	} else {
		classes = rxClassTrim.ReplaceAllString(" "+attr.Val+" ", " ")
	}

	return
}

func getClassesSlice(classes string) []string {
	return strings.Split(rxClassTrim.ReplaceAllString(" "+classes+" ", " "), " ")
}

func removeAttr(n *html.Node, attrName string) {
	for i, a := range n.Attr {
		if a.Key == attrName {
			n.Attr[i], n.Attr[len(n.Attr)-1], n.Attr =
				n.Attr[len(n.Attr)-1], html.Attribute{}, n.Attr[:len(n.Attr)-1]
			return
		}
	}
}

func setClasses(n *html.Node, attr *html.Attribute, classes string) {
	classes = strings.TrimSpace(classes)
	if classes == "" {
		removeAttr(n, "class")
		return
	}

	attr.Val = classes
}
