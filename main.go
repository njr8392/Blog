package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx"
	"html/template"
	"net/http"
	"os"
	"strconv"
)

type Post struct {
	Id   int
	User string
	Data string

	// Comments []*Comment
}
type Posts []*Post

var tpl *template.Template
var conn *pgx.Conn

func init() {
	var err error
	conn, err = pgx.Connect(context.Background(), "user="+os.Getenv("PGUSER")+" "+"password="+os.Getenv("PGPASSWORD")+" "+"port="+os.Getenv("PGPORT")+" "+"dbname="+os.Getenv("PGDATABASE"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to database %v\n", err)
	}
	tpl = template.Must(template.ParseGlob("./templates/*.html"))
}
func list() (*Posts, error) {
	var pts Posts
	rows, err := conn.Query(context.Background(), "SELECT * From posts order by id")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.Id, &p.User, &p.Data)
		if err != nil {
			fmt.Printf("%s", err)
			return nil, err
		}
		pts = append(pts, &p)
	}
	defer rows.Close()
	return &pts, rows.Err()
}

func insert(user string, data string) error {
	_, err := conn.Exec(context.Background(), "insert into posts(person, data) values($1, $2)", user, data)
	return err
}

func delete(id string) error {
	_, err := conn.Exec(context.Background(), "delete from posts where id=$1", id)
	return err
}
func getPostId(id string) (Post, error) {
	row := conn.QueryRow(context.Background(), "select * from posts where id=$1", id)

	post := Post{}
	er := row.Scan(&post.Id, &post.User, &post.Data)
	if er != nil {
		fmt.Printf("Error scanning post: %s", er)
	}
	return post, nil
}

// marshal dbquery into json
func toJson(p Posts) []byte {
	data, err := json.MarshalIndent(p, "", "    ")
	if err != nil {
		fmt.Printf("JSON marshaling failed: %s", err)
	}
	return data
}

// route handlers
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
	}
	posts, _ := list()
	tpl.ExecuteTemplate(w, "home.html", posts)
}
func createBlogForm(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "new.html", nil)
}
func createBlogPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}
	// get form values
	p := Post{}
	p.User = r.FormValue("user")
	p.Data = r.FormValue("data")

	if p.User == "" || p.Data == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
	}
	err := insert(p.User, p.Data)
	if err != nil {
		fmt.Printf("error inserting into database: ", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/blog", http.StatusSeeOther)
}
func editPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	id := r.FormValue("id")
	post, err := getPostId((id))
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
	}
	tpl.ExecuteTemplate(w, "edit.html", post)
}
func editPostProcess(w http.ResponseWriter, r *http.Request){
	var err error
	if r.Method != "POST"{
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
	}
	p := Post{}
	p.User = r.FormValue("user")
	p.Data = r.FormValue("data")
	p.Id, err = strconv.Atoi(r.FormValue("id"))
	if err != nil{
		fmt.Printf("Error converting to an int: %s", p.Id)
	}
	if p.User == "" || p.Data== "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
	}
	_, er := conn.Exec(context.Background(), "update posts set person = $1, data = $2 where id = $3", p.User, p.Data, p.Id)
	if er != nil{
		fmt.Printf("error updating post %s: %s",p.Id, er) 
	}
	http.Redirect(w,r, "/blog", http.StatusSeeOther)
}
func deletePost(w http.ResponseWriter, r *http.Request){
	if r.Method != "GET"{
		http.Error(w, http.StatusText(405),http.StatusMethodNotAllowed)
	}
	id := r.FormValue("id")
	fmt.Printf("%s", id)
	fmt.Println(len(id))

	err := delete(id)
	if err != nil {
	fmt.Printf("Error deleting post from database:\t %s ", err)
	}
	http.Redirect(w, r, "/blog", http.StatusSeeOther)

}

func main() {
	// css files not working as links in html docs, why?
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("public"))))
	http.HandleFunc("/", home)
	http.HandleFunc("/blog", homeHandler)
	http.HandleFunc("/blog/new", createBlogForm)
	http.HandleFunc("/blog/new/process", createBlogPost)
	http.HandleFunc("/blog/edit/", editPost)
	http.HandleFunc("/blog/edit/process", editPostProcess)
	http.HandleFunc("/blog/delete/process/", deletePost)
	http.ListenAndServe(":8000", nil)

}
