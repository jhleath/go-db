# Go-DB

A simple ORM for the Go programming language.

This is in very early development, and currently only supports SQLite.

## Principles

- Utilize go interfaces as much as possible.
- Seamless user experience.
- Simple, understandable commands.
- Minimize boilerplate LOC.

## Example and Tutorial

#### Model Setup

Utilizes ActiveRecord-like syntax (HasOne and HasMany).

Special Types

##### db.PrimaryKey

Specify the auto-incrementing integer that will be used as the Primary Key for the table.

##### *db.HasOne

Specify a HasOne relationship to another table (ForeignKey). Requires a field tag that specifies the foreign table.

##### *db.HasMany

Specify a reverse to HasOne relationship. Requires field tag that specifies the table and column.


    type Story struct {
      Id     db.PrimaryKey
      Name   string
      Slug   string
      Body   string
      Author *db.HasOne `table:"author"`
    }

    type Author struct {
      Id      db.PrimaryKey
      Name    string
      Stories *db.HasMany `table:"story", on:"author"`
    }

#### Creating Tables

    stories := db.CreateTableFromStruct("story", true, &Story{})
    authors := db.CreateTableFromStruct("author", true, &Author{})

#### Inserting Records

    s := &Story {
      Name: "Hello",
    }

    stories.Insert(s)

#### Simple Queries

    results := []Story{}
    stories.Get().Where("slug", "hello-world").Where("author", 5).All(&results)

    stories.Get().Where("key", "value").Order("date", true).Limit(5).One()
    stories.Get().Where("key", "value").Order("date", true).Limit(5).All()

#### Relationships

    // Set a HasOne relationship to an object.
    s.Author.Set(&Author{
      Name: "Hunter Leath",
    })

    stories.Update(s)

    // Retrieve HasOne relationship
    author := &Author{}
    s.Author.One(author)

    // Change the actual value of the field.
    s.Author.Value = 5

    // Retrieve HasMany relationship
    stories := &Story{}
    author.Stories.All(stories)
    author.Stories.Order("views", true).All(stories)

### Extending Go-DB

    type JoinStatement struct {
      a db.Clause
      b db.Clause
    }

    func (j *JoinStatement) Compile() (string, map[string]interface{})

    "SELECT * FROM table WHERE x = :x" <-> map[string]interface{}{ "x" : 5 }

    func (j *JoinStatement) Exec(db db.Executor) error {}
