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

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var BASE_URL = os.Getenv("BLOG_BASEURL")

func postMessage(post postModel) {
	var toPost bytes.Buffer
	post.Timestamp = time.Now()
	enc := json.NewEncoder(&toPost)
	enc.Encode(post)
	req, err := http.NewRequest(http.MethodPost, BASE_URL+"/blog/new", &toPost)
	if err != nil {
		log.Panicln("HTTP NewReq err", err)
	}
	req.Header.Set("Authorization", "Bearer "+post.SecretKey)
	req.Header.Set("content-type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusAccepted {
		log.Panicf("Post Title: %v\nAuth Key: %v\n\nHTTP Send Err: %v", post.Title, post.SecretKey, err)
	}
}

type postModel struct {
	Timestamp time.Time `json:"timestamp,omitempty"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	Tags      []string  `json:"tags"`
	PostText  string    `json:"postText"`
	SecretKey string    `json:"-"`
}

func (thePost postModel) splitTags(mashText string) {
	thePost.Tags = strings.Split(mashText, ",")
}

func (thePost postModel) setPostTitle(title string) {
	thePost.Title = title
}

func (thePost postModel) setSubTitle(sub string) {
	thePost.Subtitle = sub
}

func (thePost postModel) setPostText(text string) {
	thePost.PostText = text
}

func (thePost postModel) setSecret(text string) {
	thePost.SecretKey = text
}

func main() {
	if _, test := os.LookupEnv("BLOG_BASEURL"); !test {
		fmt.Print("Set `BLOG_BASEURL` to ensure your posts will actually be sent somewhere.\n\n")
		os.Exit(0)
	}
	theNewPost := postModel{}
	app := tview.NewApplication()
	form := tview.NewForm().
		AddTextView("Posting to ", BASE_URL, 50, 0, false, false).
		AddInputField("Title", "", 50, nil, func(text string) { theNewPost.Title = text }).
		AddInputField("Subtitle", "", 50, nil, func(text string) { theNewPost.Subtitle = text }).
		AddInputField("Tags", "", 50, nil, func(text string) { theNewPost.Tags = strings.Split(text, ",") }).
		AddTextArea("Actual Post", "", 0, 25, 0, func(text string) { theNewPost.PostText = text }).
		AddPasswordField("Secret Key", "", 50, '*', func(text string) { theNewPost.SecretKey = text }).
		AddButton("Submit Post", func() { postMessage(theNewPost); app.Stop() })
	form.SetBorder(true).SetBorderColor(tcell.NewRGBColor(252, 208, 177)).SetTitle("Alicolliar's Blog Posting App")
	if err := app.SetRoot(form, true).EnablePaste(true).Run(); err != nil {
		log.Panicln(err)
	}
}
