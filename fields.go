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

func ForeignKey(obj interface{}) *HasOne {
	id := -1
	examineObject(obj,
		func(p PrimaryKey, n string) {
			id = int(p)
		}, nil, nil, nil)
	return &HasOne{
		Value: id,
	}
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

var primaryKeyType = reflect.TypeOf(PrimaryKey(0))

var hasOneType = reflect.TypeOf((*HasOne)(nil))
var hasManyType = reflect.TypeOf((*HasMany)(nil))
