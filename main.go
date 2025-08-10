package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gosimple/slug"

	c "github.com/ostafen/clover"
)

//go:embed templates static/*
var content embed.FS

type PageMeta struct {
	PageTitle string
}

type NeatModel struct {
	Template *template.Template
	Database *c.DB
}

const blogCollection = "blogPosts"

var secureToken string

func (theModel NeatModel) indexHandler(w http.ResponseWriter, req *http.Request) {
	indexMetaData := PageMeta{""}
	if err := theModel.Template.ExecuteTemplate(w, "index.tmpl", indexMetaData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type blogPost struct {
	Timestamp time.Time `json:"timestamp,omitempty"`
	UrlSlug   string    `json:"urlSlug,omitempty"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	Tags      []string  `json:"tags"`
	PostText  string    `json:"postText"`
}

func (theModel NeatModel) postReceivePath(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	log.Println("Post received")
	if req.Header.Get("Authorization") != ("Bearer " + secureToken) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Println("Post correctly authed")
	var newPost *blogPost
	err := decoder.Decode(&newPost)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		log.Println("Unprocessable request")
		return
	}
	if newPost.Timestamp.IsZero() {
		newPost.Timestamp = time.Now()
	}
	newPost.UrlSlug = slug.Make(newPost.Title)
	err = theModel.Database.Insert(blogCollection, c.NewDocumentOf(newPost))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
}

func (theModel NeatModel) renderAllPosts(w http.ResponseWriter, req *http.Request) {
	var blogPosts []blogPost
	dbDocs, _ := theModel.Database.Query(blogCollection).Sort(c.SortOption{Field: "Timestamp", Direction: -1}).FindAll()
	for _, dbPost := range dbDocs {
		var newPost blogPost
		dbPost.Unmarshal(&newPost)
		blogPosts = append(blogPosts, newPost)
	}
	pageMetaData := map[string]any{
		"PageTitle": "Blog",
		"Posts":     blogPosts,
	}
	if err := theModel.Template.ExecuteTemplate(w, "blogListPage.tmpl", pageMetaData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("All post error", err)
	}
}

func (theModel NeatModel) renderAPost(w http.ResponseWriter, r *http.Request) {
	var thePost blogPost
	blogSlug := r.PathValue("blogSlug")
	theQuery, err := theModel.Database.Query(blogCollection).Where(c.Field("UrlSlug").Eq(blogSlug)).FindFirst()
	if err != nil {
		if err == c.ErrDocumentNotExist {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Post render error", err)
		return
	}
	theQuery.Unmarshal(&thePost)
	pageMetaData := map[string]any{
		"Title": thePost.Title,
		"Post":  thePost,
	}

	if err := theModel.Template.ExecuteTemplate(w, "blogPostPage.tmpl", pageMetaData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Post render error", err)
	}
}

func (theModel NeatModel) renderTagList(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("sortTag")
	var blogPosts []blogPost
	dbDocs, err := theModel.Database.Query(blogCollection).Where(c.Field("Tags").Contains(tag)).FindAll()
	if err != nil {
		if err == c.ErrDocumentNotExist {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, dbDocs := range dbDocs {
		var newPost blogPost
		dbDocs.Unmarshal(&newPost)
		blogPosts = append(blogPosts, newPost)
	}
	pageMetaData := map[string]any{
		"Title":   (tag) + " Posts",
		"SortTag": tag,
		"Posts":   blogPosts,
	}
	if err := theModel.Template.ExecuteTemplate(w, "blogListPage.tmpl", pageMetaData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Tag List error", err)
	}
}

func main() {
	daMux := http.NewServeMux()
	db, _ := c.Open("blog-db")
	defer db.Close()
	if dbExis, _ := db.HasCollection(blogCollection); !dbExis {
		db.CreateCollection(blogCollection)
	}
	funcs := template.FuncMap{
		"markedDown": func(post string) template.HTML {
			return template.HTML(markdown.ToHTML([]byte(post), nil, nil))
		},
	}
	pageModel := NeatModel{
		Template: template.Must(template.New("TheTemplates").Funcs(funcs).ParseFS(content, `templates/*.tmpl`)),
		Database: db,
	}
	secureToken = fmt.Sprint(rand.Int())
	log.Println("Random thingy", secureToken)
	daMux.Handle("GET /static/", http.FileServer(http.FS(content)))
	daMux.HandleFunc("GET /", pageModel.indexHandler)
	daMux.HandleFunc("POST /blog/new", pageModel.postReceivePath)
	daMux.HandleFunc("GET /blog/", pageModel.renderAllPosts)
	daMux.HandleFunc("GET /blog/{blogSlug}", pageModel.renderAPost)
	daMux.HandleFunc("GET /blog/tags/{sortTag}", pageModel.renderTagList)
	fmt.Println("localhost:3000")
	if err := http.ListenAndServe(":3000", daMux); err != nil {
		log.Panicln(err)
	}
}
