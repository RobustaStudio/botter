package main

import "fmt"
import "log"
import "regexp"
import "strings"
import "net/url"
import "io/ioutil"
import "encoding/xml"

func CompileFile(filename string) (tree *AST, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = xml.Unmarshal(data, &tree)
	if err != nil {
		return
	}
	if tree.ID == "" {
		tree.ID = filename
	}
	tree.NavsIndex = make(map[string]*Nav)
	tree.InputsIndex = make(map[string]*Input)
	tree.FormsIndex = make(map[string]*Form)
	for navIndex, nav := range tree.Navs {
		if nav.ID == "" {
			nav.ID = fmt.Sprintf("nav%d", navIndex+1)
		}
		if nav.Title == "" {
			nav.Title = "Please choose an action"
		}
		nav.LinksIndex = make(map[string]*Link)
		tree.NavsIndex[nav.ID] = nav
		for linkIndex, link := range nav.Links {
			if link.ID == "" {
				link.ID = fmt.Sprintf("link%d", linkIndex)
			}
			link.HrefParsed, _ = url.Parse(link.Href)
			if link.EmbedRatio == "" {
				link.EmbedRatio = "tall"
			}
			nav.LinksIndex[link.ID] = link
		}
	}
	if tree.NavsIndex["main"] == nil {
		log.Fatal("Cannot find the main `Nav` (`<nav id='main'> ... </nav>`)")
	}
	for formIndex, form := range tree.Forms {
		if form.ID == "" {
			form.ID = fmt.Sprintf("input%d", formIndex+1)
		}
		if form.Title == "" {
			form.Title = "please follow the next steps"
		}
		if form.Submit == "" {
			form.Submit = "Thank you"
		}
		for inputIndex, input := range form.Inputs {
			if input.ID == "" {
				input.ID = fmt.Sprintf("input%d", inputIndex+1)
			}
			if input.Title == "" {
				input.Title = "please write the required data"
			}
			if input.Type == "" {
				input.Type = "text"
			}
			input.Type = strings.ToLower(input.Type)
			input.NS = form.ID
			input.Path = form.ID + "/" + input.ID
			input.IfPlain = regexp.MustCompile(`\s`).ReplaceAllString(input.IfPlain, "")
			matches := regexp.MustCompile(`(\w+)(=+)(\w+)`).FindAllStringSubmatch(input.IfPlain, -1)
			if len(matches) > 0 && len(matches[0]) > 3 {
				input.IfParsed = If{
					Left:  matches[0][1],
					Op:    matches[0][2],
					Right: matches[0][3],
				}
				if input.IfParsed.Right == "null" || input.IfParsed.Right == "nil" {
					input.IfParsed.Right = ""
				}
			}
			tree.InputsIndex[input.Path] = input
		}
		tree.FormsIndex[form.ID] = form
	}
	tree.ConfigsMap = make(map[string]string)
	for _, meta := range tree.Meta {
		tree.ConfigsMap[meta.Key] = meta.Value
	}
	if tree.ConfigsMap["error"] == "" {
		tree.ConfigsMap["error"] = "Unhandled input :("
	}
	for _, tpl := range tree.Templates {
		tpl.MatchCompiled = regexp.MustCompile(tpl.Match)
	}
	return
}
