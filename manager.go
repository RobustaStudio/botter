package main

import "strings"
import "net/url"
import "net/http"
import "github.com/go-redis/redis"

// Our session (state) Manager
type Manager struct {
	db   *redis.Client
	tree *AST
}

// Initialize a new Session Manager
func NewManager(client *redis.Client, tree *AST) *Manager {
	return &Manager{client, tree}
}

// Retrieve the current input of the specified session id from the datastore
func (m *Manager) GetCurrentInput(sessid string) *Input {
	path := m.db.Get(m.tree.ID + ":session:" + sessid + ":current").Val()
	if path == "" {
		return nil
	}
	return m.tree.InputsIndex[path]
}

// Update the current input of the specified session id with the new provided path
func (m *Manager) SetCurrentInput(sessid, path string) bool {
	return m.isOk(m.db.Set(m.tree.ID+":session:"+sessid+":current", path, 0).Val())
}

// Detect the next step (input) based on the current state
func (m *Manager) GetNextInput(sessid string) *Input {
	current := m.GetCurrentInput(sessid)
	if current == nil {
		return nil
	}
	prev := ""
	form := m.tree.FormsIndex[current.NS]
	if form == nil {
		return nil
	}
	for _, input := range m.tree.FormsIndex[current.NS].Inputs {
		if prev == current.ID {
			return input
		}
		prev = input.ID
	}
	return nil
}

// Reset means clearing/flushing the current input of the specified session in the store
func (m *Manager) Reset(sessid string) bool {
	return m.SetCurrentInput(sessid, "")
}

// Answer the specified form input
func (m *Manager) SetAnswer(sessid, path string, value interface{}) bool {
	return m.db.HSet(m.tree.ID+":session:"+sessid+":answer", path, value).Val()
}

func (m *Manager) GetAnswer(sessid, path string) string {
	return m.db.HGet(m.tree.ID+":session:"+sessid+":answer", path).Val()
}

// Submit session form data to the target
func (m *Manager) Submit(sessid string) bool {
	var form *Form
	data := url.Values{}
	for k, v := range m.db.HGetAll(m.tree.ID + ":session:" + sessid + ":answer").Val() {
		input := m.tree.InputsIndex[k]
		if nil == form {
			form = m.tree.FormsIndex[input.NS]
		}
		data.Set(input.ID, v)
	}
	go http.PostForm(form.Action, data)
	m.db.HDel(m.tree.ID + ":session:" + sessid + ":answer")
	m.db.Del(m.tree.ID + ":session:" + sessid + ":current")
	return true
}

// Just return the underlying tree
func (m *Manager) GetTree() *AST {
	return m.tree
}

// just to tell us whether some redis commands return true or false
func (m *Manager) isOk(res string) bool {
	return strings.TrimSpace(strings.ToLower(res)) == "ok"
}
