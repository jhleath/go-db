package db

import (
	"database/sql"
	"fmt"
)

type insertHandler func(int64)
type statementHandler func()

type Statement interface {
	Exec(db Executor) (sql.Result, error)
}

type InsertStatement struct {
	Table    string
	Values   map[string]interface{}
	postExec insertHandler
}

func (c *InsertStatement) Compile() (string, map[string]interface{}) {
	columns := ""
	values := ""
	for key, _ := range c.Values {
		if columns != "" {
			columns += ", "
			values += ", "
		}
		columns += `"` + key + `"`
		values += (":" + key)
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", c.Table, columns, values), c.Values
}

func (c *InsertStatement) Exec(db Executor) (sql.Result, error) {
	stmt, obj := c.Compile()
	results, err := db.NamedExec(stmt, obj)
	if err == nil {
		id, err := results.LastInsertId()
		if err == nil {
			c.postExec(id)
		}
	}
	return results, err
}

// Update Statement Creates an SQL Update
type UpdateStatement struct {
	Table    string
	Where    Clause
	Columns  Clause
	postExec statementHandler
}

func (c *UpdateStatement) Compile() (string, map[string]interface{}) {
	where, whereObjects := c.Where.Compile()
	set, setObjects := c.Columns.Compile()

	return fmt.Sprintf("UPDATE %s SET %s WHERE %s", c.Table, set, where), mapUnion(whereObjects, setObjects)
}

func (c *UpdateStatement) Exec(db Executor) (sql.Result, error) {
	stmt, obj := c.Compile()
	results, err := db.NamedExec(stmt, obj)
	if err == nil {
		c.postExec()
	}
	return results, err
}

type DeleteStatement struct {
	Table string
	Where Clause
}

func (c *DeleteStatement) Compile() (string, map[string]interface{}) {
	whereStmt, whereObj := c.Where.Compile()
	return fmt.Sprintf("DELETE FROM %s WHERE %s", c.Table, whereStmt), whereObj
}

func (c *DeleteStatement) Exec(db Executor) (sql.Result, error) {
	stmt, obj := c.Compile()
	return db.NamedExec(stmt, obj)
}

type CreateTableStatement struct {
	Name   string
	Fields []Field
	Force  bool
	Key    string
}

func (c *CreateTableStatement) Compile() (string, map[string]interface{}) {
	exists := ""
	if !c.Force {
		exists = "IF NOT EXISTS"
	}

	columns := ""
	for i, v := range c.Fields {
		if i != 0 {
			columns += ", "
		}
		columns += fmt.Sprintf("\"%s\" %s", v.Name, v.Type)
	}

	if c.Key != "" {
		columns += fmt.Sprintf(", CONSTRAINT %s_pk PRIMARY KEY (%s)", c.Name, c.Key)
	}

	return fmt.Sprintf("CREATE TABLE %s %s (%s)", exists, c.Name, columns), nil
}

func (c *CreateTableStatement) Exec(db Executor) (sql.Result, error) {
	stmt, obj := c.Compile()
	return db.NamedExec(stmt, obj)
}
