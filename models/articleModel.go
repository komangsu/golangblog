package models

import (
	"golangblog/config"
	"time"
)

type Article struct {
	Id        uint64    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorId  uint64    `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CreateArticle(title, content string, authorid uint64) error {

	db := config.InitDB()
	defer db.Close()

	query := `insert into articles(title,content,author_id) values($1,$2,$3)`

	stmt, err := db.Prepare(query)
	if err != nil {
		panic(err)
	}

	stmt.QueryRow(title, content, authorid)

	return nil
}
