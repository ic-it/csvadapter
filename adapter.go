package csvadapter

import (
	"encoding"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"iter"
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

	options *csvAdapterOptions
}

func (c CSVAdapter[T]) String() string {
	return fmt.Sprintf("CSVAdapter(%s)", c.structType.Name())
}

// NewCSVAdapter creates a new CSVAdapter
func NewCSVAdapter[T any](options ...csvAdapterOption) (*CSVAdapter[T], error) {
	var TEmpty T
	t := reflect.TypeOf(TEmpty)

	// TODO: Support for pointers/maps
	if t.Kind() != reflect.Struct {
		return nil, errors.Join(ErrorNotStruct, fmt.Errorf("type %s", t.Kind()))
	}

	csvAdapter := &CSVAdapter[T]{
		structType: t,
		fields:     make([]field, 0),
		options:    newCSVAdapterOptions(),
	}

	for _, option := range options {
		option(csvAdapter.options)
	}

iterOverFields:
	for i := 0; i < t.NumField(); i++ {
		field := field{}
		fld := t.Field(i)
		tag := fld.Tag.Get(_TAG)
		field.name = fld.Name
		if !csvAdapter.options.noImplicitAlias {
			field.alias = fld.Name // default alias
		}
		isAliasSet := false
		tagParts := strings.Split(tag, ",")
		for _, part := range tagParts {
			if part == "" {
				continue
			}
			if part == _TAG_SKIP {
				continue iterOverFields
			}
			kv := strings.Split(part, "=")
			var key, value string
			if len(kv) == 1 {
				key = kv[0]
			} else if len(kv) == 2 {
				key, value = kv[0], kv[1]
			} else {
				return nil, errors.Join(ErrInvalidTag, fmt.Errorf("tag %s", part))
			}
			switch key {
			case _TAG_ALIAS:
				field.alias = value
			case _TAG_OMITEMPTY:
				field.omitEmpty = true
			default:
				// first part without key is the alias
				if !isAliasSet {
					field.alias = key
					isAliasSet = true
				} else {
					return nil, errors.Join(ErrUnsupportedTag, fmt.Errorf("tag %s", part))
				}
			}
		}

		if field.alias == "" {
			return nil, errors.Join(ErrAliasNotFound, fmt.Errorf("field %s", field.name))
		}

		csvAdapter.fields = append(csvAdapter.fields, field)
	}

	return csvAdapter, nil
}

// FromCSV reads a csv file and fills a slice of structs
func (c *CSVAdapter[T]) FromCSV(reader io.Reader) (iter.Seq2[T, error], error) {
	csvReader := csv.NewReader(reader)
	c.options.applyReader(csvReader)

	header, err := csvReader.Read()
	if err != nil {
		return nil, errors.Join(ErrReadingCSVLines, err)
	}
	// create a map of the columns order
	columnsOrder := make(map[string]int, len(header))
	for i, h := range header {
		columnsOrder[h] = i
	}

	// check if all fields are present in the csv
	for _, f := range c.fields {
		if _, isFound := columnsOrder[f.alias]; !isFound {
			if f.omitEmpty {
				continue
			}
			return nil, errors.Join(ErrFieldNotFound, fmt.Errorf("field %s", f.alias))
		}
	}

	return func(yield func(T, error) bool) {
		var TEmpty T
		line := 0
	loopOverLines:
		for {
			line++
			record, err := csvReader.Read()
			if err == io.EOF {
				break loopOverLines
			}
			if err != nil {
				if !yield(TEmpty, errors.Join(ErrReadingCSVLines, err)) {
					return
				}
				continue loopOverLines
			}
			s := reflect.New(c.structType).Elem()
			for _, f := range c.fields {
				fieldErr := errors.Join(
					ErrProcessingCSVLines,
					ReadingError{
						Line:       line,
						Field:      f.name,
						FieldAlias: f.alias,
					})
				index, isFound := columnsOrder[f.alias]
				if !isFound && f.omitEmpty {
					continue
				} else if !isFound { // I think its actually impossible to reach this point
					if !yield(TEmpty, errors.Join(fieldErr, ErrFieldNotFound)) {
						return
					}
					continue loopOverLines
				}
				value := record[index]
				if value == "" && f.omitEmpty {
					continue
				} else if value == "" {
					if !yield(TEmpty, errors.Join(fieldErr, ErrEmptyValue)) {
						return
					}
					continue loopOverLines
				}
				field := s.FieldByName(f.name)
				if err := unmarshalField(field, value); err != nil {
					if !yield(TEmpty, errors.Join(fieldErr, err)) {
						return
					}
					continue loopOverLines
				}
			}
			if !yield(s.Interface().(T), nil) {
				return
			}
		}
	}, nil
}

// ToCSV writes a slice of structs to a csv file
func (c *CSVAdapter[T]) ToCSV(writer io.Writer, data iter.Seq[T]) error {
	csvWriter := csv.NewWriter(writer)
	c.options.applyWriter(csvWriter)
	defer csvWriter.Flush()

	// write header
	if c.options.writeHeader {
		header := make([]string, len(c.fields))
		for i, f := range c.fields {
			header[i] = f.alias
		}
		if err := csvWriter.Write(header); err != nil {
			return errors.Join(ErrReadingCSV, err)
		}
	}

	// write records
	line := 0
	for item := range data {
		line++
		itemV := reflect.ValueOf(item)
		record := make([]string, len(c.fields))
		for i, f := range c.fields {
			fieldErr := errors.Join(
				ErrProcessingCSVLines,
				ReadingError{
					Line:       line,
					Field:      f.name,
					FieldAlias: f.alias,
				})
			field := itemV.FieldByName(f.name)
			if !field.IsValid() {
				return errors.Join(fieldErr, ErrFieldNotFound)
			}
			if field.Kind() == reflect.Ptr && field.IsNil() {
				continue
			}
			str, err := marshalField(field)
			if err != nil {
				return errors.Join(fieldErr, err)
			}
			if str == "" && f.omitEmpty {
				continue
			} else if str == "" {
				return errors.Join(fieldErr, ErrEmptyValue)
			}
			record[i] = str
		}
		if err := csvWriter.Write(record); err != nil {
			return errors.Join(ErrReadingCSV, err)
		}
	}
	return nil
}

// unmarshals a string value to a field
// based on the type of the field
func unmarshalField(field reflect.Value, value string) error {
	switch field.Kind() {
	// strings
	case reflect.String:
		field.SetString(value)
	// integers
	case reflect.Int:
		i, err := strconv.Atoi(value)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetInt(int64(i))
	case reflect.Int8:
		i, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetInt(i)
	case reflect.Int16:
		i, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetInt(i)
	case reflect.Int32:
		i, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetInt(i)
	case reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetInt(i)
	// booleans
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetBool(b)
	// floats
	case reflect.Float32:
		f, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetFloat(f)
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetFloat(f)
	// unsigned integers
	case reflect.Uint:
		i, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetUint(i)
	case reflect.Uint8:
		i, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetUint(i)
	case reflect.Uint16:
		i, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetUint(i)
	case reflect.Uint32:
		i, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetUint(i)
	case reflect.Uint64:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return errors.Join(ErrParsingType, err)
		}
		field.SetUint(i)
	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return unmarshalField(field.Elem(), value)
	default:
		if field.CanAddr() {
			// check if the field implements encoding.TextUnmarshaler
			if u, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
				return u.UnmarshalText([]byte(value))
			}
		}
		return errors.Join(ErrUnprocessableType, fmt.Errorf("type %s", field.Kind()))
	}
	return nil
}

// marshalField marshals a field to a string
// based on the type of the field
func marshalField(field reflect.Value) (string, error) {
	switch field.Kind() {
	// strings
	case reflect.String:
		return field.String(), nil
	// integers
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", field.Int()), nil
	// booleans
	case reflect.Bool:
		return fmt.Sprintf("%t", field.Bool()), nil
	// floats
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", field.Float()), nil
	// unsigned integers
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", field.Uint()), nil
	case reflect.Ptr:
		if field.IsNil() {
			return "", nil
		}
		return marshalField(field.Elem())
	default:
		if field.CanAddr() {
			// check if the field implements encoding.TextMarshaler
			if m, ok := field.Addr().Interface().(encoding.TextMarshaler); ok {
				b, err := m.MarshalText()
				if err != nil {
					return "", err
				}
				return string(b), nil
			}
			// check if the field implements fmt.Stringer
			if s, ok := field.Addr().Interface().(fmt.Stringer); ok {
				return s.String(), nil
			}
		}
		return "", errors.Join(ErrUnprocessableType, fmt.Errorf("type %s", field.Kind()))
	}
}

// Errors
var (
	ErrUnsupportedTag      = fmt.Errorf("unsupported tag")
	ErrInvalidTag          = fmt.Errorf("invalid tag")
	ErrorNotStruct         = fmt.Errorf("not a struct")
	ErrReadingCSV          = fmt.Errorf("error reading csv")
	ErrReadingCSVLines     = fmt.Errorf("error reading csv lines")
	ErrProcessingCSVLines  = fmt.Errorf("error processing csv lines")
	ErrFieldNotFound       = fmt.Errorf("field not found in csv")
	ErrUnprocessableType   = fmt.Errorf("unprocessable type")
	ErrParsingType         = fmt.Errorf("error parsing type")
	ErrEmptyValue          = fmt.Errorf("empty value")
	ErrAliasNotFound       = fmt.Errorf("alias not found")
	ErrWrongNumberOfFields = fmt.Errorf("wrong number of fields")
)

const (
	_TAG           = "csva"
	_TAG_OMITEMPTY = "omitempty"
	_TAG_ALIAS     = "alias"
	_TAG_SKIP      = "-"
)
