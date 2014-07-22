package db

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Query interface {
	One(db Executor, object interface{}) error
	All(db Executor, object []interface{}) error
}

// A Simple SQL Select Statement
type SelectStatement struct {
	Table       string
	WhereClause Clause
	LimitClause Clause
	OrderClause Clause
}

func (c *SelectStatement) Compile() (string, map[string]interface{}) {
	outStatement := fmt.Sprintf("SELECT * FROM %s", c.Table)
	outObjects := make(map[string]interface{})

	if c.WhereClause != nil {
		whereStmt, whereObj := c.WhereClause.Compile()
		outStatement = fmt.Sprintf("%s WHERE (%s)", outStatement, whereStmt)
		outObjects = mapUnion(outObjects, whereObj)
	}

	if c.OrderClause != nil {
		orderStmt, orderObj := c.OrderClause.Compile()
		outStatement = fmt.Sprintf("%s ORDER BY %s", outStatement, orderStmt)
		outObjects = mapUnion(outObjects, orderObj)
	}

	if c.LimitClause != nil {
		limitStmt, limitObj := c.LimitClause.Compile()
		outStatement = fmt.Sprintf("%s LIMIT %s", outStatement, limitStmt)
		outObjects = mapUnion(outObjects, limitObj)
	}

	return outStatement, outObjects
}

func (q *SelectStatement) Order(key string, ascending bool) *SelectStatement {
	q.OrderClause = &OrderClause{
		Key:       key,
		Ascending: ascending,
	}
	return q
}

func (q *SelectStatement) Limit(number int) *SelectStatement {
	q.LimitClause = &LimitClause{
		Number: number,
	}
	return q
}

func (q *SelectStatement) WhereClauseAnd(where Clause) *SelectStatement {
	if q.WhereClause == nil {
		a := make(AndClauses, 1)
		a[0] = where
		q.WhereClause = a
	} else {
		obj, ok := q.WhereClause.(AndClauses)
		if !ok {
			panic("Cannot and to something that isn't an AND Clause.")
		}
		obj = append(obj, where)
		q.WhereClause = obj
	}
	return q
}

func (q *SelectStatement) Where(key string, value interface{}) *SelectStatement {
	return q.WhereClauseAnd(&NamedEquality{
		Name:  key,
		Value: value,
	})
}

func (q *SelectStatement) One(db Executor, object interface{}) error {
	stmt, obj := q.Compile()
	rows, err := db.NamedQuery(stmt, obj)
	if err != nil {
		return err
	}

	return rows.StructScan(object)
}

func (q *SelectStatement) All(db Executor, object interface{}) error {
	stmt, obj := q.Compile()
	rows, err := db.NamedQuery(stmt, obj)
	if err != nil {
		return err
	}

	return sqlx.StructScan(rows, object)
}

func (c *SelectStatement) Exec(db Executor) (sql.Result, error) {
	stmt, obj := c.Compile()
	return db.NamedExec(stmt, obj)
}
