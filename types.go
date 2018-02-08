package main

import "regexp"
import "encoding/xml"
import "net/url"

type AST struct {
	ID          string            `xml:"id,attr"`
	XMLName     xml.Name          `json:"-" xml:"html"`
	Version     string            `json:"version" xml:"version,attr"`
	Meta        []*Meta           `json:"-" xml:"head>meta"`
	Navs        []*Nav            `json:"-" xml:"body>nav"`
	Forms       []*Form           `json:"-" xml:"body>form"`
	Templates   []*Template       `json:"-" xml:"body>template"`
	InputsIndex map[string]*Input `json:"inputs"  xml:"-"`
	FormsIndex  map[string]*Form  `json:"forms" xml:"-"`
	NavsIndex   map[string]*Nav   `json:"navs" xml:"-"`
	ConfigsMap  map[string]string `json:"configs" xml:"-"`
}

type Meta struct {
	Key   string `xml:"name,attr"`
	Value string `xml:"content,attr"`
}

type Nav struct {
	ID         string           `json:"id" xml:"id,attr"`
	Title      string           `json:"title" xml:"title,attr"`
	Links      []*Link          `json:"-" xml:"a"`
	LinksIndex map[string]*Link `json:"links" xml:"-"`
}

type Link struct {
	ID         string   `xml:"id,attr"`
	Text       string   `xml:",innerxml"`
	Href       string   `xml:"href,attr"`
	HrefParsed *url.URL `xml:"-"`
	Reset      bool     `xml:"reset,attr"`
	Embed      bool     `xml:"embed,attr"`
	EmbedRatio string   `xml:"ratio,attr"`
}

type Form struct {
	ID     string   `xml:"id,attr"`
	Title  string   `xml:"title,attr"`
	Submit string   `xml:"submit,attr"`
	Action string   `xml:"action,attr"`
	Method string   `xml:"method,attr"`
	Inputs []*Input `xml:"input"`
}

type Input struct {
	NS       string    `xml:"-"`
	Path     string    `xml:"-"`
	ID       string    `xml:"id,attr"`
	Title    string    `xml:"title,attr,omitempty"`
	Type     string    `xml:"type,attr,omitempty"`
	IfPlain  string    `xml:"if,attr"`
	IfParsed If        `xml:"-"`
	Options  []*Option `xml:"option,omitempty"`
}

type Template struct {
	Match         string         `xml:"match,attr"`
	MatchCompiled *regexp.Regexp `xml:"-"`
	Replies       []*Reply       `xml:"reply"`
}

type Reply struct {
	Type   string `xml:"type,attr"`
	Label  string `xml:"label,attr"`
	Source string `xml:"src,attr"`
}

type If struct {
	Right string
	Op    string
	Left  string
}

type Option struct {
	Key  string `xml:"value,attr"`
	Text string `xml:",innerxml"`
}
