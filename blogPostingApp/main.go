package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	titleIn = iota
	subIn
	postIn
	tagsIn
	secretIn
)

var BASE_URL = os.Getenv("BLOG_BASEURL")

var (
	inputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fcd0b1"))
)

func postMessage(secret string, tagString string, post postModel) error {
	var toPost bytes.Buffer
	post.Tags = strings.Split(tagString, ",")
	post.Timestamp = time.Now()
	enc := json.NewEncoder(&toPost)
	enc.Encode(post)
	req, err := http.NewRequest(http.MethodPost, BASE_URL+"/blog/new", &toPost)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+secret)
	req.Header.Set("content-type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusAccepted {
		return err
	}
	return nil
}

type commandModel struct {
	Secret    string
	Post      postModel
	postIn    textarea.Model
	tagString string
	Inputs    []textinput.Model
	focused   int
	Err       string
}

type postModel struct {
	Timestamp time.Time `json:"timestamp,omitempty"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	Tags      []string  `json:"tags"`
	PostText  string    `json:"postText"`
}

func initialModel() commandModel {
	var newModel = commandModel{
		Inputs: make([]textinput.Model, 5),
	}
	newModel.Inputs[titleIn] = textinput.New()
	newModel.Inputs[titleIn].Focus()
	newModel.Inputs[subIn] = textinput.New()
	newModel.postIn = textarea.New()
	newModel.postIn.SetHeight(25)
	newModel.postIn.SetWidth(150)
	newModel.Inputs[tagsIn] = textinput.New()
	newModel.Inputs[secretIn] = textinput.New()
	newModel.focused = 0
	return newModel
}

func (model commandModel) Init() tea.Cmd {
	return textinput.Blink
}

func (model commandModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(model.Inputs)+1)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			model.focused++
			if model.focused > 4 {
				model.focused = 0
			}
		case tea.KeyShiftTab:
			model.focused--
			if model.focused < 0 {
				model.focused = 4
			}
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyCtrlQ:
			return model, tea.Quit
		case tea.KeyEnter:
			if model.focused == secretIn {
				if err := postMessage(model.Secret, model.tagString, model.Post); err != nil {
					model.Err = err.Error()
				}
			}
		}
		model.postIn.Blur()
		for i := range model.Inputs {
			model.Inputs[i].Blur()
		}
		if model.focused == postIn {
			model.postIn.Focus()
		} else {
			model.Inputs[model.focused].Focus()
		}
	}

	model.Post.Title = model.Inputs[titleIn].Value()
	model.Post.Subtitle = model.Inputs[subIn].Value()
	model.Post.PostText = model.postIn.Value()
	model.tagString = model.Inputs[tagsIn].Value()
	model.Secret = model.Inputs[secretIn].Value()

	for i := range model.Inputs {
		model.Inputs[i], cmds[i] = model.Inputs[i].Update(msg)
	}
	model.postIn, cmds[4] = model.postIn.Update(msg)

	return model, tea.Batch(cmds...)
}

func (model commandModel) View() string {
	return fmt.Sprintf(
		"%s\n%s %s\n%s %s\n%s %s\n%s \n%s\n\n%s %s\n%s %s\n%s",
		inputStyle.Bold(true).Render("Alicolliar's Blog Posting TUI"),
		inputStyle.Render("Posting to:"),
		inputStyle.Render(BASE_URL),
		inputStyle.Render("Post Title"),
		model.Inputs[titleIn].View(),
		inputStyle.Render("Post Subtitle"),
		model.Inputs[subIn].View(),
		inputStyle.Render("Post Text"),
		model.postIn.View(),
		inputStyle.Render("Post Tags (Comma-seperated)"),
		model.Inputs[tagsIn].View(),
		inputStyle.Render("Site Secret"),
		model.Inputs[secretIn].View(),
		inputStyle.Render("(Press Enter while selecting 'Site Secret' to Send to platform)"),
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
