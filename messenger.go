package main

import "io"
import "io/ioutil"
import "fmt"
import "math/rand"
import "time"
import "strconv"
import "net/http"
import "net/url"
import "errors"
import "github.com/paked/messenger"

// a Msssenger is a session containing the compiled AST and a session Manager
type Messenger struct {
	manager *Manager
	tree    *AST
	configs map[string]string
	sessid  string
	bot     *messenger.Messenger
	output  *messenger.Response
}

// create a new Messenger instance using the specified manager
func NewMessenger(manager *Manager) http.Handler {
	this := new(Messenger)
	this.manager = manager
	this.tree = manager.GetTree()
	this.configs = this.tree.ConfigsMap
	this.bot = messenger.New(messenger.Options{
		Verify:      this.configs["verify-secret"] == "true",
		AppSecret:   this.configs["app-secret"],
		VerifyToken: this.configs["verify-token"],
		Token:       this.configs["page-token"],
	})
	return this
}

// send a welcome message to the newly created messenger chat
func (this *Messenger) SetGreetingInfo() error {
	description := this.configs["description"]
	if err := this.bot.GreetingSetting(description); err != nil {
		return err
	}
	return this.SetMainNav()
}

// send the main navigation menu "messegner related"
func (this *Messenger) SetMainNav() error {
	mainNav := this.tree.NavsIndex["main"]
	callToActions := []messenger.CallToActionsItem{}
	for _, link := range mainNav.Links {
		action := messenger.CallToActionsItem{}
		action.Title = link.Text
		if link.HrefParsed.Scheme == "http" || link.HrefParsed.Scheme == "https" {
			action.Type = "web_url"
			action.URL = link.Href
			if link.Embed {
				action.WebviewHeightRatio = link.EmbedRatio
				action.MessengerExtension = true
			}
		} else {
			payload, _ := url.Parse(link.Href)
			payloadQuery := payload.Query()
			payloadQuery.Set("src_nav_id", mainNav.ID)
			payloadQuery.Set("src_link_id", link.ID)
			payloadQuery.Set("reset", "no")
			if link.Reset {
				payloadQuery.Set("reset", "yes")
			}
			payload.RawQuery = payloadQuery.Encode()
			action.Type = "postback"
			action.Payload = payload.String()
		}
		callToActions = append(callToActions, action)
	}
	return this.bot.CallToActionsSetting("existing_thread", callToActions)
}

// send a navigation menu
func (this *Messenger) SendNav(nav *Nav) error {
	buttons := []messenger.StructuredMessageButton{}
	for _, link := range nav.Links {
		button := messenger.StructuredMessageButton{}
		button.Title = link.Text
		if link.HrefParsed.Scheme == "http" || link.HrefParsed.Scheme == "https" {
			button.Type = "web_url"
			button.URL = link.Href
			if link.Embed {
				button.WebviewHeightRatio = link.EmbedRatio
				button.MessengerExtensions = true
			}
		} else {
			payload, _ := url.Parse(link.Href)
			payloadQuery := payload.Query()
			payloadQuery.Set("src_nav_id", nav.ID)
			payloadQuery.Set("src_link_id", link.ID)
			payloadQuery.Set("reset", "no")
			if link.Reset {
				payloadQuery.Set("reset", "yes")
			}
			payload.RawQuery = payloadQuery.Encode()
			button.Type = "postback"
			button.Payload = payload.String()
		}
		buttons = append(buttons, button)
	}
	return this.output.ButtonTemplate(nav.Title, &buttons)
}

// here we are detecting the current/next input
// then send it to the current messenger thread.
func (this *Messenger) SendForm(form *Form) error {
	if err := this.output.Text(form.Title); err != nil {
		return err
	}
	if len(form.Inputs) < 1 {
		return nil
	}
	sessCurrentInput := this.manager.GetCurrentInput(this.sessid)
	if sessCurrentInput == nil {
		sessCurrentInput = form.Inputs[0]
		this.manager.SetCurrentInput(this.sessid, sessCurrentInput.Path)
	}
	return this.SendInput(sessCurrentInput)
}

// send an input to the current messenger thread
func (this *Messenger) SendInput(input *Input) error {
	if !this.IsInputSendable(input) {
		sessNextInput := this.manager.GetNextInput(this.sessid)
		if nil != sessNextInput {
			this.manager.SetCurrentInput(this.sessid, sessNextInput.Path)
			return this.SendInput(sessNextInput)
		} else {
			this.Submit(this.tree.FormsIndex[this.tree.InputsIndex[input.Path].NS].Submit)
		}
	}
	switch input.Type {
	case "text", "file":
		return this.output.Text(input.Title)
	case "options":
		options := []messenger.StructuredMessageButton{}
		for index, opt := range input.Options {
			indexStr := strconv.Itoa(index)
			option := messenger.StructuredMessageButton{}
			option.Title = opt.Text
			option.Type = "postback"
			option.Payload = "answer://" + input.NS + "?input_path=" + input.Path + "&selected_option_index=" + indexStr
			options = append(options, option)
		}
		return this.output.ButtonTemplate(input.Title, &options)
	default:
		return this.output.Text(this.configs["error"])
	}
	return nil
}

// whether the specified input is sendable "interpretting the if stmnt"
func (this *Messenger) IsInputSendable(input *Input) bool {
	if input.IfPlain == "" {
		return true
	}
	leftInput := this.tree.InputsIndex[input.NS+"/"+input.IfParsed.Left]
	if leftInput == nil {
		return false
	}
	leftVal := this.manager.GetAnswer(this.sessid, leftInput.Path)
	rightVal := input.IfParsed.Right
	leftValInt, _ := strconv.Atoi(leftVal)
	rightValInt, _ := strconv.Atoi(rightVal)
	switch input.IfParsed.Op {
	case "==":
		return leftVal == rightVal
	case "!=":
		return leftVal != rightVal
	case ">":
		return leftValInt > rightValInt
	case ">=":
		return leftValInt >= rightValInt
	case "<":
		return leftValInt < rightValInt
	case "<=":
		return leftValInt <= rightValInt
	}
	return false
}

// store the specified answer into the session storage
func (this *Messenger) SetAnswer(path string, value interface{}) error {
	this.manager.SetAnswer(this.sessid, path, value)
	sessNextInput := this.manager.GetNextInput(this.sessid)
	if nil != sessNextInput {
		this.manager.SetCurrentInput(this.sessid, sessNextInput.Path)
		return this.SendInput(sessNextInput)
	} else {
		this.Submit(this.tree.FormsIndex[this.tree.InputsIndex[path].NS].Submit)
	}
	return nil
}

// finally submit the whole collected data
func (this *Messenger) Submit(thankyou string) {
	this.output.Text(thankyou)
	this.manager.Submit(this.sessid)
	this.manager.Reset(this.sessid)
}

// a helper function to just attach the specified session id and our main writer to the current messenger thread
func (this *Messenger) SetupSession(sessid string, writer *messenger.Response) *Messenger {
	this.sessid = sessid
	this.output = writer
	return this
}

// here we compile the `<template>` tags and send it to the current thread
func (this *Messenger) ProcessFromTemplates(needle string) error {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	for _, tpl := range this.tree.Templates {
		if tpl.MatchCompiled.MatchString(needle) {
			reply := tpl.Replies[r.Intn(len(tpl.Replies))]
			if reply.Label != "" {
				this.output.Text(reply.Label)
			}
			if reply.Source != "" && reply.Type != "" {
				this.output.Attachment(messenger.AttachmentType(reply.Type), reply.Source)
			}
			return nil
		}
	}
	return errors.New("Cannot find a correct reply")
}

// calling the `composer` backend to return an answer to us
func (this *Messenger) ProcessFromComposer(needle string) error {
	url := fmt.Sprintf(this.configs["composer"], url.QueryEscape(needle))
	resp, err := http.Get(url)
	if err != nil {
		return this.output.Text(this.configs["error"])
	}
	resp.Body = ioutil.NopCloser(io.LimitReader(resp.Body, 500*1024))
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	if len(data) < 1 {
		return this.output.Text(this.configs["error"])
	}
	return this.output.Text(string(data))
}

// our thread messenger server
func (this *Messenger) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	this.bot.HandleOptIn(func(_ messenger.OptIn, r *messenger.Response) {
		r.SenderAction("mark_seen")
		r.SenderAction("typing_on")
		defer r.SenderAction("typing_off")
		this.SetGreetingInfo()
		this.SendNav(this.tree.NavsIndex["main"])
	})

	this.bot.HandleMessage(func(m messenger.Message, r *messenger.Response) {
		r.SenderAction("mark_seen")
		r.SenderAction("typing_on")
		defer r.SenderAction("typing_off")
		sessid := strconv.Itoa(int(m.Sender.ID))
		if this.sessid == "" {
			this.SetupSession(sessid, r)
		}
		sessCurrentInput := this.manager.GetCurrentInput(sessid)
		if sessCurrentInput == nil {
			done := false
			if len(this.tree.Templates) > 0 {
				done = this.ProcessFromTemplates(m.Text) == nil
			}
			if ! done && this.configs["composer"] != "" {
				done = true
				this.ProcessFromComposer(m.Text)
			}
			if ! done {
				this.output.Text(this.configs["error"])
			}
		} else if sessCurrentInput.Type == "text" {
			this.SetAnswer(sessCurrentInput.Path, m.Text)
		} else if sessCurrentInput.Type == "file" {
			if len(m.Attachments) < 1 {
				r.Text("Please, upload a valid file :)")
			} else {
				this.SetAnswer(sessCurrentInput.Path, m.Attachments[0].Payload.URL)
			}
		}
	})

	this.bot.HandlePostBack(func(p messenger.PostBack, r *messenger.Response) {
		r.SenderAction("mark_seen")
		r.SenderAction("typing_on")
		defer r.SenderAction("typing_off")
		sessid := strconv.Itoa(int(p.Sender.ID))
		if this.sessid == "" {
			this.SetupSession(sessid, r)
		}
		if p.Payload == "get_started" {
			this.SetGreetingInfo()
			this.SendNav(this.tree.NavsIndex["main"])
			return
		}
		needle, _ := url.Parse(p.Payload)
		if needle.Query().Get("reset") == "yes" {
			this.manager.Reset(this.sessid)
		}
		switch needle.Scheme {
		case "nav":
			trgtNav := this.tree.NavsIndex[needle.Host]
			if trgtNav == nil {
				r.Text(this.configs["error"])
				break
			}
			this.SendNav(trgtNav)
		case "form":
			trgtForm := this.tree.FormsIndex[needle.Host]
			if trgtForm == nil {
				r.Text(this.configs["error"])
				break
			}
			this.SendForm(trgtForm)
		case "answer":
			sessCurrentInput := this.manager.GetCurrentInput(sessid)
			if sessCurrentInput == nil {
				this.SendNav(this.tree.NavsIndex["main"])
			} else if sessCurrentInput.Type == "options" {
				selectedOptionIndex, _ := strconv.Atoi(needle.Query().Get("selected_option_index"))
				this.SetAnswer(sessCurrentInput.Path, sessCurrentInput.Options[selectedOptionIndex].Key)
			} else if sessCurrentInput.Type == "file" {
				this.output.Text("Please upload a valid file :)")
			} else {
				this.output.Text(this.configs["error"])
			}
		default:
			this.output.Text(this.configs["error"])
		}
	})

	this.bot.Handler().ServeHTTP(res, req)
}
