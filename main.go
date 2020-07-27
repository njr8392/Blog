package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx"
	"os"
)

type Post struct {
	Id   int
	User string
	Data string

	// Comments []*Comment
}
type Posts []Post

func list() (Posts, error) {
	var pts Posts

	rows, err := conn.Query(context.Background(), "SELECT * From posts")
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
		// wtf p prints the struct correctly but whne pushed into an array it is 0 and 1
		fmt.Printf("%+v\n", p)
		pts = append(pts, p)
		// fmt.Printf("%d:\t%s\t%s\t\n", p.Id, p.User, p.Data)
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

var conn *pgx.Conn

func main() {
	var err error
	conn, err = pgx.Connect(context.Background(), "user=postgres password=17reflekTor?71 host=localhost port=5432 dbname=blog")
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to database %v\n", err)
	}
	// insert("steve", "hey from the main function")
	posts, err := list()
	if err != nil {
		fmt.Println(err)
	}
	for po := range posts {
		fmt.Printf("%+v\t%T\n", po, po)
	}

}
