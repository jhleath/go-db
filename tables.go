package db

import (
	"reflect"
)

type Database interface {
	Executor
	DriverName() string
}

type handleprimaryKeyType func(PrimaryKey, string)
type handlehasOneType func(*HasOne, string)
type handlehasManyType func(*HasMany, string)
type handleDefaultType func(interface{}, reflect.Kind, string)

func examineObject(object interface{}, pt handleprimaryKeyType, ho handlehasOneType, hm handlehasManyType, d handleDefaultType) {
	// Value of Object
	val := reflect.ValueOf(object).Elem()

	// Loop Through Fields
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		switch typeField.Type {
		case primaryKeyType:
			if pt != nil {
				pt(valueField.Interface().(PrimaryKey), typeField.Name)
			}
		case hasOneType:
			if ho != nil {
				ho(valueField.Interface().(*HasOne), typeField.Name)
			}
		case hasManyType:
			if hm != nil {
				hm(valueField.Interface().(*HasMany), typeField.Name)
			}
		default:
			if d != nil && typeField.Tag.Get("db") != "-" {
				d(valueField.Interface(), valueField.Kind(), typeField.Name)
			}
		}
	}
}

// Author   *db.HasOne  `table:"author"`
// StorySet *db.HasMany `table:"story", on:"author"`

func loadRelationships(object interface{}, id int64) {
	if reflect.TypeOf(object).Kind() != reflect.Ptr {
		panic("Can't load relationships on non-pointer object.")
	}
	// Value of Object
	val := reflect.ValueOf(object).Elem()

	if id == -1 {
		for i := 0; i < val.NumField(); i++ {
			if val.Type().Field(i).Type == primaryKeyType {
				id = int64(val.Field(i).Interface().(PrimaryKey))
			}
		}
	}

	// Loop Through Fields
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		// Get Foreign Table and Columns
		foreignTable := typeField.Tag.Get("table")
		foreignColumn := typeField.Tag.Get("on")

		switch typeField.Type {
		case primaryKeyType:
			valueField.SetInt(id)
		case hasOneType:
			// Current Field Value
			current := valueField.Interface().(*HasOne)
			// Default to Id for Foreign Column
			if foreignColumn == "" {
				foreignColumn = "id"
			}

			value := 0
			if current != nil {
				value = current.Value
			}
			// Load Into New Value
			hasOne := &HasOne{
				Value:  value,
				column: foreignColumn,
			}

			hasOne.SelectStatement = (&SelectStatement{
				Table: toSnakeCase(foreignTable),
			}).Where(toSnakeCase(foreignColumn), value)
			// Set New Value
			valueField.Set(reflect.ValueOf(hasOne))
		case hasManyType:
			// Load Into New Value
			hasMany := &HasMany{}
			hasMany.SelectStatement = (&SelectStatement{
				Table: toSnakeCase(foreignTable),
			}).Where(toSnakeCase(foreignColumn), id)
			// Set New Value
			valueField.Set(reflect.ValueOf(hasMany))
		}
	}
}

func scan(object interface{}) {
	if reflect.TypeOf(object).Kind() != reflect.Ptr {
		panic("Can't scan into object that isn't a pointer.")
	}
}

type Field struct {
	Name string
	Type string
}

type Table interface {
	// Access
	Get() *SelectStatement
	Update(object interface{}) *UpdateStatement
	Insert(object interface{}) *InsertStatement
	Delete(object interface{}) *DeleteStatement
}

type BasicTable struct {
	TableName string
	Fieldset  []Field
	Key       string
	DB        Executor
}

func ConvertKindToDB(db Database, r reflect.Kind, pk bool) string {
	if pk && db.DriverName() == "postgres" {
		return "serial"
	}

	switch r {
	case reflect.Int:
		return "integer"
	case reflect.String:
		return "text"
	case reflect.Slice:
		if db.DriverName() == "postgres" {
			return "bytea"
		}
		return "blob" // bytea
	case reflect.Float64:
		return "real"
	case reflect.Bool:
		return "numeric"
	}
	return "unknown"
}

func CreateTableFromStruct(name string, db Database, force bool, object interface{}) (*BasicTable, error) {
	// Create Table Struct
	out := &BasicTable{
		TableName: name,
		Fieldset:  make([]Field, 0),
		DB:        db,
	}

	// Fillout Fieldset
	examineObject(object,
		func(p PrimaryKey, name string) {
			out.Fieldset = append(out.Fieldset, Field{
				Name: toSnakeCase(name),
				Type: ConvertKindToDB(db, reflect.Int, true),
			})
			out.Key = toSnakeCase(name)
		},
		func(p *HasOne, name string) {
			out.Fieldset = append(out.Fieldset, Field{
				Name: toSnakeCase(name),
				Type: ConvertKindToDB(db, reflect.Int, false),
			})
		},
		nil,
		func(p interface{}, r reflect.Kind, name string) {
			out.Fieldset = append(out.Fieldset, Field{
				Name: toSnakeCase(name),
				Type: ConvertKindToDB(db, r, false),
			})
		})

	// Create Table
	_, err := out.CreateTable(force).Exec(db)
	return out, err
}

func (b BasicTable) CreateTable(force bool) *CreateTableStatement {
	return &CreateTableStatement{
		Name:   b.TableName,
		Fields: b.Fieldset,
		Force:  force,
		Key:    b.Key,
	}
}

func (b BasicTable) Get() *SelectStatement {
	return &SelectStatement{
		Table: b.TableName,
	}
}

func (b BasicTable) GetBy(object interface{}, key string, value interface{}) error {
	return b.Get().Where(key, value).One(b.DB, object)
}

func (b BasicTable) Delete(object interface{}) *DeleteStatement {
	id := 0
	idField := ""

	examineObject(object, func(p PrimaryKey, n string) {
		id = int(p)
		idField = toSnakeCase(n)
	}, nil, nil, nil)

	return &DeleteStatement{
		Table: b.TableName,
		Where: &NamedEquality{
			Name:  idField,
			Value: id,
		},
	}
}

func (b BasicTable) Update(object interface{}) *UpdateStatement {
	id := 0
	idField := ""

	columnsClause := make(SetClause, 0)

	examineObject(object,
		func(p PrimaryKey, n string) {
			id = int(p)
			idField = toSnakeCase(n)
		},
		func(ho *HasOne, name string) {
			columnsClause = append(columnsClause, &NamedEquality{
				Name:  toSnakeCase(name),
				Value: ho.Value,
			})
		},
		nil,
		func(d interface{}, r reflect.Kind, name string) {
			columnsClause = append(columnsClause, &NamedEquality{
				Name:  toSnakeCase(name),
				Value: d,
			})
		})

	return &UpdateStatement{
		Table: b.TableName,
		Where: &NamedEquality{
			Name:  idField,
			Value: id,
		},
		Columns: columnsClause,
		postExec: func() {
			loadRelationships(object, -1)
		},
	}
}

func (b BasicTable) Insert(object interface{}) *InsertStatement {
	values := make(map[string]interface{})

	examineObject(object,
		nil,
		func(ho *HasOne, name string) {
			value := 0
			if ho != nil {
				value = ho.Value
			}

			values[toSnakeCase(name)] = value
		},
		nil,
		func(d interface{}, r reflect.Kind, name string) {
			values[toSnakeCase(name)] = d
		})

	return &InsertStatement{
		Table:  b.TableName,
		Values: values,
		postExec: func(id int64) {
			loadRelationships(object, id)
		},
	}
}
