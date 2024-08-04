package csvadapter

import (
	"encoding"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type field struct {
	name      string // name of the field in the struct
	alias     string // name of the field in the csv
	omitEmpty bool   // if the field can be empty
}

// CSVAdapter is a struct that adapts a struct to a csv file
type CSVAdapter[T any] struct {
	structType reflect.Type
	fields     []field // fields of the struct
}

func (c CSVAdapter[T]) String() string {
	return fmt.Sprintf("CSVAdapter(%s)", c.structType.Name())
}

// NewCSVAdapter creates a new CSVAdapter
func NewCSVAdapter[T any]() (*CSVAdapter[T], error) {
	var tempty T
	t := reflect.TypeOf(tempty)

	csvAdapter := &CSVAdapter[T]{
		structType: t,
		fields:     make([]field, 0),
	}

	for i := 0; i < t.NumField(); i++ {
		field := field{}
		fld := t.Field(i)
		tag := fld.Tag.Get("csvadapter")
		if tag == "" {
			continue
		}
		field.name = fld.Name
		tagParts := strings.Split(tag, ",")
		fielAlias := tagParts[0]
		if fielAlias == "" {
			fielAlias = fld.Name
		}
		field.alias = fielAlias
		for _, part := range tagParts[1:] {
			if part == "omitempty" {
				field.omitEmpty = true
			} else {
				return nil, errors.Join(ErrUnsupportedTag, fmt.Errorf("tag %s", part))
			}
		}

		csvAdapter.fields = append(csvAdapter.fields, field)
	}

	return csvAdapter, nil
}

// EatCSV reads a csv file and fills a slice of structs
func (c *CSVAdapter[T]) EatCSV(reader io.Reader, v *[]T) error {
	csvReader := csv.NewReader(reader)
	header, err := csvReader.Read()
	if err != nil {
		return errors.Join(ErrReadingCSV, err)
	}
	fieldsPositions := map[string]int{}
	for i, h := range header {
		fieldsPositions[h] = i
	}

	line := 0
	for {
		line++
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Join(ErrReadingCSVLines, err)
		}
		s := reflect.New(c.structType).Elem()
		for _, v := range c.fields {
			fieldErr := errors.Join(
				ErrProcessingCSVLines,
				fmt.Errorf("line %d, field %s", line, v.alias))
			pos, isFound := fieldsPositions[v.alias]
			if !isFound {
				if v.omitEmpty {
					continue
				}
				return errors.Join(fieldErr, ErrFieldNotFound)
			}
			value := record[pos]
			if value == "" && v.omitEmpty {
				continue
			} else if value == "" {
				return errors.Join(fieldErr, ErrEmptyValue)
			}
			field := s.FieldByName(v.name)
			if err := setField(field, value); err != nil {
				return errors.Join(fieldErr, err)
			}
		}
		*v = append(*v, s.Interface().(T))
	}
	return nil
}

// setField sets the value of a field in a struct
// based on the type of the field
func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.Atoi(value)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetInt(int64(i))
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetFloat(f)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetUint(i)
	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setField(field.Elem(), value)
	default:
		if field.CanAddr() {
			if u, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
				return u.UnmarshalText([]byte(value))
			}
		} else {
			return errors.Join(ErrUnprocessableType, fmt.Errorf("type %s", field.Kind()))
		}
	}
	return nil
}

// Errors
var (
	ErrUnsupportedTag     = fmt.Errorf("unsupported tag")
	ErrorNotStruct        = fmt.Errorf("not a struct")
	ErrReadingCSV         = fmt.Errorf("error reading csv")
	ErrReadingCSVLines    = fmt.Errorf("error reading csv lines")
	ErrProcessingCSVLines = fmt.Errorf("error processing csv lines")
	ErrFieldNotFound      = fmt.Errorf("field not found in csv")
	ErrUnprocessableType  = fmt.Errorf("unprocessable type")
	ErrParsingType        = fmt.Errorf("error parsing type")
	ErrEmptyValue         = fmt.Errorf("empty value")
)
