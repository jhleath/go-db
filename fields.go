package db

import (
	"reflect"
)

// story.author     (parent <- child)
// SELECT * FROM AUTHOR WHERE ID = ?

// author.story_set (parent -> children)
// SELECT * FROM STORY WHERE AUTHOR_ID = ?

type HasMany struct {
	*SelectStatement
}

type HasOne struct {
	*SelectStatement
	Value int
	// Internal
	column string
}

func (f *HasOne) Set(obj interface{}) {
	id := -1
	examineObject(obj,
		func(p PrimaryKey, n string) {
			id = int(p)
		}, nil, nil, nil)
	f.Value = id
	f.SelectStatement.WhereClause = nil
	f.SelectStatement = f.SelectStatement.Where(f.column, id)
}

type PrimaryKey int

var PrimaryType = reflect.TypeOf(PrimaryKey(0)).Elem()

var HasOneType = reflect.TypeOf(HasOne{})
var HasManyType = reflect.TypeOf(HasMany{})
