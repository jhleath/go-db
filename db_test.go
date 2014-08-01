package db

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

type Story struct {
	Id       PrimaryKey
	Name     string
	Body     string
	Slug     string
	SlugBody string
	Author   *HasOne `table:"author"`
}

type Author struct {
	Id      PrimaryKey
	Name    string
	Stories *HasMany `table:"story" on:"author"`
}

type Data struct {
	Statement  string
	Parameters map[string]interface{}
}

type TestResult struct{}

func (t TestResult) LastInsertId() (int64, error) {
	return -5, nil
}

func (t TestResult) RowsAffected() (int64, error) {
	return 1, nil
}

type TestDb struct {
	Data chan Data
}

func (t *TestDb) NamedExec(query string, arg interface{}) (sql.Result, error) {
	t.Data <- Data{
		Statement:  query,
		Parameters: arg.(map[string]interface{}),
	}
	return TestResult{}, nil
}

func (t *TestDb) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	return nil, nil
}

func TestSelect(t *testing.T) {
	dataChan := make(chan Data, 1)
	connection := &TestDb{
		Data: dataChan,
	}

	// Create Author Table
	authorTable, err := CreateTableFromStruct("author", connection, true, &Author{})
	if err != nil {
		t.Error(err.Error())
	}

	data := <-dataChan
	if data.Statement != "CREATE TABLE  author (id integer, name text, CONSTRAINT author_pk PRIMARY KEY (id))" {
		t.Error("Creating Authors Table Incorrect SQL")
	}
	fmt.Println(data.Statement, data.Parameters)

	// Create Story Table
	storyTable, err := CreateTableFromStruct("story", connection, false, &Story{})
	if err != nil {
		t.Error(err.Error())
	}

	data = <-dataChan
	if data.Statement != "CREATE TABLE IF NOT EXISTS story (id integer, name text, body text, slug text, author integer, CONSTRAINT story_pk PRIMARY KEY (id))" {
		t.Error("Creating Stories Table Incorrect SQL")
	}
	fmt.Println(data.Statement, data.Parameters)

	// Insert an Author
	newAuthor := &Author{
		Name: "Hunter Leath",
	}
	_, err = authorTable.Insert(newAuthor).Exec(connection)
	if err != nil {
		t.Error(err.Error())
	}

	if newAuthor.Id != -5 {
		t.Error("Id not set successfully.")
	}

	data = <-dataChan
	if data.Statement != "INSERT INTO author (name) VALUES (:name)" {
		t.Error("Inserting Author Incorrect SQL")
	}
	fmt.Println(data.Statement, data.Parameters)

	// Insert a Story
	newStory := &Story{
		Name:   "Going to the beach",
		Body:   "Lorem ipsum.",
		Slug:   "going-to-the-beach",
		Author: ForeignKey(newAuthor),
	}
	_, err = storyTable.Insert(newStory).Exec(connection)
	if err != nil {
		t.Error(err.Error())
	}

	if newStory.Id != -5 {
		t.Error("Id not set successfully.")
	}

	data = <-dataChan
	if !strings.HasPrefix(data.Statement, "INSERT INTO story") {
		t.Error("Inserting Story Incorrect SQL")
	}
	fmt.Println(data.Statement, data.Parameters)

	// Update a Story
	newStory.Body = "Lorem Ipsum 2"

	_, err = storyTable.Update(newStory).Exec(connection)
	if err != nil {
		t.Error(err.Error())
	}

	data = <-dataChan
	if !strings.HasPrefix(data.Statement, "UPDATE story SET") {
		t.Error("Updating Story Incorrect SQL")
	}
	fmt.Println(data.Statement, data.Parameters)

	// Basic Select Queries
	_, err = storyTable.Get().Where("slug", "going-to-the-beach").Where("author", -5).Order("slug", true).Limit(5).Exec(connection)
	if err != nil {
		t.Error(err.Error())
	}

	data = <-dataChan
	if data.Statement != "SELECT * FROM story WHERE (slug = :variable_slug AND author = :variable_author) ORDER BY slug ASC LIMIT 5" {
		t.Error("Selecting Story Incorrect SQL")
	}
	fmt.Println(data.Statement, data.Parameters)

	// HasOne Relationship
	_, err = newStory.Author.Exec(connection)
	if err != nil {
		t.Error(err.Error())
	}

	data = <-dataChan
	if data.Statement != "SELECT * FROM author WHERE (id = :variable_id)" {
		t.Error("Selecting Author Through Relationship Incorrect SQL")
	}
	fmt.Println(data.Statement, data.Parameters)

	// HasMany Relationship
	_, err = newAuthor.Stories.Limit(2).Exec(connection)
	if err != nil {
		t.Error(err.Error())
	}

	data = <-dataChan
	if data.Statement != "SELECT * FROM story WHERE (author = :variable_author) LIMIT 2" {
		t.Error("Selecting Stories Through Relationship Incorrect SQL")
	}
	fmt.Println(data.Statement, data.Parameters)

	// Deleting Items
	_, err = authorTable.Delete(newAuthor).Exec(connection)
	if err != nil {
		t.Error(err.Error())
	}

	data = <-dataChan
	if data.Statement != "DELETE FROM author WHERE id = :variable_id" {
		t.Error("Deleting Author Incorrect SQL")
	}
	fmt.Println(data.Statement, data.Parameters)
}
