package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var BASE_URL = os.Getenv("BLOG_BASEURL")

func postMessage(post postModel, key string) {
	var toPost bytes.Buffer
	post.Timestamp = time.Now()
	enc := json.NewEncoder(&toPost)
	enc.Encode(post)
	req, err := http.NewRequest(http.MethodPost, BASE_URL+"/blog/new", &toPost)
	if err != nil {
		log.Panicln("HTTP NewReq err", err)
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("content-type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		fmt.Printf("An error occurred while sending the post!")
		os.Exit(2)
	}

}

func sendImage(imPath string, secretKey string) {
	theFile, err := os.Open(imPath)
	if err != nil {
		fmt.Print("The file path is invalid!")
		os.Exit(0)
	}
	defer theFile.Close()
	var toSend bytes.Buffer
	theImage, err := jpeg.Decode(theFile)
	if err != nil {
		fmt.Print("Not a JPEG!")
		os.Exit(0)
	}
	jpeg.Encode(&toSend, theImage, nil)
	req, _ := http.NewRequest(http.MethodPost, BASE_URL+"/blog/new/image", &toSend)
	req.Header.Set("Authorization", "Bearer "+secretKey)
	req.Header.Set("Content-type", "image/jpeg")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusCreated {
		fmt.Print("An error occured during image send!", err)
		os.Exit(2)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	fmt.Printf("Image created with ID %v.", string(bodyBytes))
}

type postModel struct {
	Timestamp time.Time `json:"timestamp,omitempty"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	Tags      []string  `json:"tags"`
	PostText  string    `json:"postText"`
}

func main() {
	imgPath := flag.String("img", "", "If set, uses this as the path for the image to upload, and then exits. Defaults to \"\" ")
	postOn := flag.Bool("post", true, "Opens the standard blog posting TUI, defaults to true")
	key := flag.String("key", "", "The secret key for the site. Must be set!")
	if _, test := os.LookupEnv("BLOG_BASEURL"); !test {
		fmt.Print("Set `BLOG_BASEURL` to ensure your posts will actually be sent somewhere.\n\n")
		os.Exit(0)
	}

	flag.Parse()

	if !strings.EqualFold(*imgPath, "") {
		sendImage(*imgPath, *key)
		os.Exit(0)
	}

	if !*postOn {
		fmt.Print("Ok, do nothing I guess, you twat")
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
		AddButton("Submit Post", func() { postMessage(theNewPost, *key); app.Stop(); fmt.Print("Post sent!") })
	form.SetBorder(true).SetBorderColor(tcell.NewRGBColor(252, 208, 177)).SetTitle("Alicolliar's Blog Posting App")
	if err := app.SetRoot(form, true).EnablePaste(true).Run(); err != nil {
		log.Panicln(err)
	}
}
