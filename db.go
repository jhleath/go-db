package db

import (
	"bytes"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"unicode"
	"unicode/utf8"
)

//
type Executor interface {
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
}

//
func toSnakeCase(x string) string {
	if len(x) == 0 {
		return ""
	}

	output := make([]byte, 0)
	for len(x) > 0 {
		v, size := utf8.DecodeRuneInString(x)

		// If underscore, append and keep going.
		if v == '_' {
			output = append(output, byte('_'))
		} else if unicode.IsLetter(v) {
			if unicode.IsLower(v) {
				// Keep it the same if it is lower.
				buf := make([]byte, size)
				utf8.EncodeRune(buf, v)
				output = bytes.Join([][]byte{output, buf}, nil)
			} else if unicode.IsUpper(v) {
				// Lowercase it otherwise.
				buf := make([]byte, size)
				utf8.EncodeRune(buf, unicode.ToLower(v))
				output = bytes.Join([][]byte{output, buf}, []byte("_"))
			}
		}

		x = x[size:]
	}

	return string(output)
}

//
func toCamelCase(x string) string {
	if len(x) == 0 {
		return ""
	}

	output := make([]byte, 0)
	uppercase := true

	for len(x) > 0 {
		v, size := utf8.DecodeRuneInString(x)

		// If underscore, append and keep going.
		if v == '_' {
			uppercase = true
		} else if unicode.IsLetter(v) {
			if uppercase {
				uppercase = false
				buf := make([]byte, size)
				utf8.EncodeRune(buf, unicode.ToUpper(v))
				output = bytes.Join([][]byte{output, buf}, nil)
			} else if unicode.IsUpper(v) {
				buf := make([]byte, size)
				utf8.EncodeRune(buf, v)
				output = bytes.Join([][]byte{output, buf}, []byte("_"))
			}
		}

		x = x[size:]
	}

	return string(output)
}

//
func mapUnion(x map[string]interface{}, y map[string]interface{}) map[string]interface{} {
	z := make(map[string]interface{})
	if x != nil {
		for key, value := range x {
			z[key] = value
		}
	}
	if y != nil {
		for key, value := range y {
			_, ok := z[key]
			if ok {
				panic("mapUnion will overwrite.")
			}
			z[key] = value
		}
	}
	return z
}

//
func JoinClausesOn(c []Clause, on string) (string, map[string]interface{}) {
	outStmt := ""
	outObj := make(map[string]interface{})
	for i, v := range c {
		if i != 0 {
			outStmt += on
		}
		tempStmt, tempObjects := v.Compile()
		outStmt += tempStmt
		outObj = mapUnion(outObj, tempObjects)
	}
	return outStmt, outObj
}
