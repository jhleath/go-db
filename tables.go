package db

import (
	"github.com/jmoiron/sqlx"
	"reflect"
)

type handlePrimaryType func(PrimaryKey, string)
type handleHasOneType func(*HasOne, string)
type handleHasManyType func(*HasMany, string)
type handleDefaultType func(interface{}, reflect.Kind, string)

func examineObject(object interface{}, pt handlePrimaryType, ho handleHasOneType, hm handleHasManyType, d handleDefaultType) {
	// Value of Object
	val := reflect.ValueOf(object).Elem()

	// Loop Through Fields
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		switch typeField.Type {
		case PrimaryType:
			if pt != nil {
				pt(valueField.Interface().(PrimaryKey), typeField.Name)
			}
		case HasOneType:
			if ho != nil {
				ho(valueField.Interface().(*HasOne), typeField.Name)
			}
		case HasManyType:
			if hm != nil {
				hm(valueField.Interface().(*HasMany), typeField.Name)
			}
		default:
			if d != nil {
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
			if val.Type().Field(i).Type == PrimaryType {
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
		case HasOneType:
			// Current Field Value
			current := valueField.Interface().(*HasOne)
			// Default to Id for Foreign Column
			if foreignColumn == "" {
				foreignColumn = "id"
			}
			// Load Into New Value
			hasOne := &HasOne{
				Value:  current.Value,
				column: foreignColumn,
			}
			hasOne.SelectStatement = (&SelectStatement{
				Table: ToSnakeCase(foreignTable),
			}).Where(ToSnakeCase(foreignColumn), current.Value)
			// Set New Value
			valueField.Set(reflect.ValueOf(hasOne))
		case HasManyType:
			// Load Into New Value
			hasMany := &HasMany{}
			hasMany.SelectStatement = (&SelectStatement{
				Table: ToSnakeCase(foreignTable),
			}).Where(ToSnakeCase(foreignColumn), id)
			// Set New Value
			valueField.Set(reflect.ValueOf(hasMany))
		}
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
}

type BasicTable struct {
	TableName string
	Fieldset  []Field
	Key       string
	DB        *sqlx.DB
}

func ConvertKindToDB(r reflect.Kind) string {
	switch r {
	case reflect.Int:
		return "integer"
	case reflect.String:
		return "text"
	case reflect.Slice:
		return "blob"
	case reflect.Float64:
		return "real"
	case reflect.Bool:
		return "numeric"
	}
	return "unknown"
}

func CreateTableFromStruct(name string, db *sqlx.DB, force bool, object interface{}) (*BasicTable, error) {
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
				Name: ToSnakeCase(name),
				Type: ConvertKindToDB(reflect.Int),
			})
			out.Key = name
		},
		func(p *HasOne, name string) {
			out.Fieldset = append(out.Fieldset, Field{
				Name: ToSnakeCase(name + "Id"),
				Type: ConvertKindToDB(reflect.Int),
			})
		},
		nil,
		func(p interface{}, r reflect.Kind, name string) {
			out.Fieldset = append(out.Fieldset, Field{
				Name: ToSnakeCase(name),
				Type: ConvertKindToDB(r),
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

	examineObject(object, func(p PrimaryKey, n string) {
		id = int(p)
	}, nil, nil, nil)

	return &DeleteStatement{
		Table: b.TableName,
		Where: &NamedEquality{
			Name:  "id",
			Value: id,
		},
	}
}

func (b BasicTable) Update(object interface{}) *UpdateStatement {
	if reflect.TypeOf(object).Kind() == reflect.Ptr {
		// Change the Fields, yo.
	}

	id := 0
	columnsClause := make(SetClause, 0)

	examineObject(object,
		func(p PrimaryKey, n string) {
			id = int(p)
		},
		func(ho *HasOne, name string) {
			columnsClause = append(columnsClause, &NamedEquality{
				Name:  ToSnakeCase(name + "Id"),
				Value: ho.Value,
			})
		},
		nil,
		func(d interface{}, r reflect.Kind, name string) {
			columnsClause = append(columnsClause, &NamedEquality{
				Name:  ToSnakeCase(name),
				Value: d,
			})
		})

	return &UpdateStatement{
		Table: b.TableName,
		Where: &NamedEquality{
			Name:  "id",
			Value: id,
		},
		Columns: columnsClause,
		postExec: func() {
			loadRelationships(object, -1)
		},
	}
}

func (b BasicTable) Insert(object interface{}) *InsertStatement {
	if reflect.TypeOf(object).Kind() == reflect.Ptr {
		// Change the Fields, yo.
	}

	values := make(map[string]interface{})

	examineObject(object,
		nil,
		func(ho *HasOne, name string) {
			values[ToSnakeCase(name+"Id")] = ho
		},
		nil,
		func(d interface{}, r reflect.Kind, name string) {
			values[ToSnakeCase(name)] = d
		})

	return &InsertStatement{
		Table:  b.TableName,
		Values: values,
		postExec: func(id int64) {
			loadRelationships(object, id)
		},
	}
}