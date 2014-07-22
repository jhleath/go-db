package db

import (
	"fmt"
)

const (
	sqlAnd   = " AND "
	sqlOr    = " OR "
	sqlComma = ", "
)

type Clause interface {
	Compile() (string, map[string]interface{})
}

type LimitClause struct {
	Number int
}

func (c LimitClause) Compile() (string, map[string]interface{}) {
	return fmt.Sprintf("%d", c.Number), nil
}

type OrderClause struct {
	Key       string
	Ascending bool
}

func (c OrderClause) Compile() (string, map[string]interface{}) {
	orderType := "ASC"
	if !c.Ascending {
		orderType = "DESC"
	}
	return fmt.Sprintf("%s %s", c.Key, orderType), nil
}

// SQL And Clauses
type AndClauses []Clause

func (c AndClauses) Compile() (string, map[string]interface{}) {
	return JoinClausesOn(c, sqlAnd)
}

// SQL Or Clauses
type OrClauses []Clause

func (c OrClauses) Compile() (string, map[string]interface{}) {
	return JoinClausesOn(c, sqlOr)
}

// SQL Set Clause
type SetClause []Clause

func (c SetClause) Compile() (string, map[string]interface{}) {
	return JoinClausesOn(c, sqlComma)
}

// Basic Variable Equality
type NamedEquality struct {
	Name  string
	Value interface{}
}

func (c *NamedEquality) Compile() (string, map[string]interface{}) {
	object := make(map[string]interface{})
	name := fmt.Sprintf("variable_%s", c.Name)
	object[name] = c.Value
	return fmt.Sprintf("%s = :%s", c.Name, name), object
}
