package csvadapter

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"slices"
	"testing"
)

type Person struct {
	Name  string `csva:"name"`
	Age   int    `csva:"age"`
	Email string `csva:"email,omitempty"`
}

type PersonNoTags struct {
	Name string `csva:""`
	Age  int
}

type PersonWrongTag struct {
	Name string `csva:"name,omitempty,foo"`
	Age  int
}

type PersonWithManyTypes struct {
	Name      string                    `csva:"name"`
	Age       int                       `csva:"age"`
	Email     *PersonWithManyTypesEmail `csva:"email,omitempty"`
	SomeFloat float64                   `csva:"some_float"`
	SomeBool  bool                      `csva:"some_bool"`
	SomePtr   *string                   `csva:"some_ptr"`
}

type PersonWithManyTypesEmail struct {
	Email string
}

func (p *PersonWithManyTypesEmail) UnmarshalText(text []byte) error {
	p.Email = string(text)
	return nil
}

func (p PersonWithManyTypesEmail) MarshalText() ([]byte, error) {
	return []byte(p.Email), nil
}

func TestNoImplicitAlias(t *testing.T) {
	type PersonWithImplicitAlias struct {
		Name  string
		Age   int    `csva:"age"`
		Email string `csva:"email,omitempty"`
	}

	type PersonWithNoImplicitAlias struct {
		Name  string `csva:"name"`
		Age   int    `csva:"age"`
		Email string `csva:"email,omitempty"`
	}

	_, err0 := NewCSVAdapter[PersonWithImplicitAlias](
		NoImplicitAlias(false),
	)
	_, err1 := NewCSVAdapter[PersonWithImplicitAlias](
		NoImplicitAlias(true),
	)
	_, err2 := NewCSVAdapter[PersonWithNoImplicitAlias](
		NoImplicitAlias(false),
	)
	_, err3 := NewCSVAdapter[PersonWithNoImplicitAlias](
		NoImplicitAlias(true),
	)

	if err0 != nil {
		t.Fatalf("expected error, got nil")
	}

	if err1 == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err1, ErrAliasNotFound) {
		t.Errorf("expected ErrAliasNotFound, got %v", err1)
	}

	if err2 != nil {
		t.Fatalf("expected error, got nil")
	}
	if err3 != nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestNewCSVAdapterSkipField(t *testing.T) {
	type PersonWithSkipField struct {
		Name  string `csva:"name"`
		Age   int    `csva:"age"`
		Email string `csva:"-"`
	}

	adapter, err := NewCSVAdapter[PersonWithSkipField]()
	if err != nil {
		t.Fatalf("failed to create csva: %v", err)
	}
	if len(adapter.fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(adapter.fields))
	}
}

func TestNewCSVAdapterWrongType(t *testing.T) {
	_, err := NewCSVAdapter[string]()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrorNotStruct) {
		t.Errorf("expected ErrorNotStruct, got %v", err)
	}
}

func TestNewCSVAdapterAliasAsKV(t *testing.T) {
	type PersonWithAliasAsKV struct {
		Name  string `csva:"alias=name"`
		Age   int    `csva:"alias=age"`
		Email string `csva:"alias=email,omitempty"`
	}

	adapter, err := NewCSVAdapter[PersonWithAliasAsKV]()
	if err != nil {
		t.Fatalf("failed to create csva: %v", err)
	}
	if len(adapter.fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(adapter.fields))
	}

	type InvalidTagged struct {
		Name  string `csva:"alias==name"`
		Age   int    `csva:"alias=age=A"`
		Email string `csva:"alias=email,omitempty"`
	}

	_, err = NewCSVAdapter[InvalidTagged]()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidTag) {
		t.Errorf("expected ErrInvalidTag, got %v", err)
	}
}

func TestNewCSVAdapter(t *testing.T) {
	t.Run("with tags", func(t *testing.T) {
		adapter, err := NewCSVAdapter[Person]()
		if err != nil {
			t.Fatalf("failed to create csva: %v", err)
		}
		if adapter.String() != "CSVAdapter(Person)" {
			t.Errorf("expected CSVAdapter(Person), got %s", adapter.String())
		}
		if len(adapter.fields) != 3 {
			t.Errorf("expected 3 fields, got %d", len(adapter.fields))
		}
	})
	t.Run("with wrong tag", func(t *testing.T) {
		_, err := NewCSVAdapter[PersonWrongTag]()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrUnsupportedTag) {
			t.Errorf("expected ErrUnsupportedTag, got %v", err)
		}
	})
}

func TestFromCSV(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		csvData := `name,age,email
John Doe,30,` + fakemail + `
Jane Smith,25,` + otherfakemail + `
`

		reader := bytes.NewReader([]byte(csvData))
		adapter, err := NewCSVAdapter[Person]()
		if err != nil {
			t.Fatalf("failed to create csva: %v", err)
		}

		people, err := adapter.FromCSV(reader)
		if err != nil {
			t.Fatalf("failed to read CSV: %v", err)
		}

		expected := []Person{
			{"John Doe", 30, fakemail},
			{"Jane Smith", 25, otherfakemail},
		}

		idx := 0
		for person, err := range people {
			if err != nil {
				t.Fatalf("failed to read person: %v", err)
			}
			if person != expected[idx] {
				t.Errorf("expected %+v, got %+v", expected[idx], person)
			}
			idx++
		}
	})

	t.Run("empty file", func(t *testing.T) {
		csvData := ``
		reader := bytes.NewReader([]byte(csvData))
		adapter, err := NewCSVAdapter[Person]()
		if err != nil {
			t.Fatalf("failed to create csva: %v", err)
		}

		_, err = adapter.FromCSV(reader)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, io.EOF) {
			t.Errorf("expected ErrEmptyFile, got %v", err)
		}

	})

	t.Run("read wrong csv", func(t *testing.T) {
		csvData := `name,age
John Doe,30,
Jane Smith,25,
`

		reader := bytes.NewReader([]byte(csvData))
		adapter, err := NewCSVAdapter[Person]()
		if err != nil {
			t.Fatalf("failed to create csva: %v", err)
		}

		p, err := adapter.FromCSV(reader)
		if err != nil {
			t.Fatalf("failed to read CSV: %v", err)
		}
		for _, err := range p {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, csv.ErrFieldCount) {
				t.Fatalf("expected ErrFieldNotFound, got %v", err)
			}
		}
	})
}

func TestFromCSVWithOmitEmpty(t *testing.T) {
	t.Run("omit empty", func(t *testing.T) {
		csvData := `name,age,email
John Doe,30,
Jane Smith,25,` + otherfakemail

		reader := bytes.NewReader([]byte(csvData))
		adapter, err := NewCSVAdapter[Person]()
		if err != nil {
			t.Fatalf("failed to create csva: %v", err)
		}

		people, err := adapter.FromCSV(reader)
		if err != nil {
			t.Fatalf("failed to read CSV: %v", err)
		}

		expected := []Person{
			{"John Doe", 30, ""},
			{"Jane Smith", 25, otherfakemail},
		}

		idx := 0
		for person, err := range people {
			if err != nil {
				t.Fatalf("failed to read person: %v", err)
			}
			if person != expected[idx] {
				t.Errorf("expected %+v, got %+v", expected[idx], person)
			}
			idx++
		}

	})

	t.Run("unomit empty", func(t *testing.T) {
		csvData := `name,age,email
John Doe,,
`

		reader := bytes.NewReader([]byte(csvData))
		adapter, err := NewCSVAdapter[Person]()
		if err != nil {
			t.Fatalf("failed to create csva: %v", err)
		}

		people, err := adapter.FromCSV(reader)
		if err != nil {
			t.Fatalf("failed to read CSV: %v", err)
		}

		for _, err := range people {
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, ErrEmptyValue) {
				t.Errorf("expected ErrEmptyValue, got %v", err)
			}
			break
		}

	})

}

func TestFromCSVWithMissingField(t *testing.T) {
	csvData := `name
John Doe
Jane Smith
`

	reader := bytes.NewReader([]byte(csvData))
	adapter, err := NewCSVAdapter[Person]()
	if err != nil {
		t.Fatalf("failed to create csva: %v", err)
	}

	_, err = adapter.FromCSV(reader)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, ErrFieldNotFound) {
		t.Errorf("expected ErrFieldNotFound, got %v", err)
	}
}

func TestFromCSVWithInvalidData(t *testing.T) {
	csvData := `name,age,email
John Doe,thirty,` + fakemail + `
`

	reader := bytes.NewReader([]byte(csvData))
	adapter, err := NewCSVAdapter[Person]()
	if err != nil {
		t.Fatalf("failed to create csva: %v", err)
	}

	people, err := adapter.FromCSV(reader)
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	for _, err := range people {
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrParsingType) {
			t.Fatalf("expected ErrParsingType, got %v", err)
		}
		break
	}
}

func TestFromCSVWithManyTypes(t *testing.T) {
	csvData := `name,age,email,some_float,some_bool,some_ptr
John Doe,30,` + fakemail + `,3.14,true,hello
Jane Smith,25,` + otherfakemail + `,2.71,false,123
`

	reader := bytes.NewReader([]byte(csvData))
	adapter, err := NewCSVAdapter[PersonWithManyTypes]()
	if err != nil {
		t.Fatalf("failed to create csva: %v", err)
	}

	people, err := adapter.FromCSV(reader)
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	expected := []PersonWithManyTypes{
		{
			"John Doe",
			30,
			&PersonWithManyTypesEmail{fakemail},
			3.14,
			true,
			stringPtr("hello"),
		},
		{
			"Jane Smith",
			25,
			&PersonWithManyTypesEmail{otherfakemail},
			2.71,
			false,
			stringPtr("123"),
		},
	}

	idx := 0
	for person, err := range people {
		if err != nil {
			t.Fatalf("failed to read person: %v", err)
		}
		if person.Name != expected[idx].Name ||
			person.Age != expected[idx].Age ||
			person.Email.Email != expected[idx].Email.Email ||
			person.SomeFloat != expected[idx].SomeFloat ||
			person.SomeBool != expected[idx].SomeBool ||
			*person.SomePtr != *expected[idx].SomePtr {
			t.Errorf("expected %+v, got %+v", expected[idx], person)
		}
		idx++
	}
}

func TestToCSV(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		adapter, err := NewCSVAdapter[Person]()
		if err != nil {
			t.Fatalf("failed to create csva: %v", err)
		}

		people := []Person{
			{"John Doe", 30, fakemail},
			{"Jane Smith", 25, otherfakemail},
		}

		writer := &bytes.Buffer{}
		err = adapter.ToCSV(writer, slices.Values(people))
		if err != nil {
			t.Fatalf("failed to write CSV: %v", err)
		}

		expected := `name,age,email
John Doe,30,` + fakemail + `
Jane Smith,25,` + otherfakemail + `
`
		if writer.String() != expected {
			t.Errorf("expected %s, got %s", expected, writer.String())
		}
	})
}

func TestToCSVWithOmitEmpty(t *testing.T) {
	adapter, err := NewCSVAdapter[Person]()
	if err != nil {
		t.Fatalf("failed to create csva: %v", err)
	}

	people := []Person{
		{"John Doe", 30, ""},
		{"Jane Smith", 25, otherfakemail},
	}

	writer := &bytes.Buffer{}
	err = adapter.ToCSV(writer, slices.Values(people))
	if err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	expected := `name,age,email
John Doe,30,
Jane Smith,25,` + otherfakemail + `
`
	if writer.String() != expected {
		t.Errorf("expected %s, got %s", expected, writer.String())
	}
}

func TestToCSVWithMissingField(t *testing.T) {
	adapter, err := NewCSVAdapter[Person]()
	if err != nil {
		t.Fatalf("failed to create csva: %v", err)
	}

	people := []Person{
		{"John Doe", 30, fakemail},
		{"Jane Smith", 25, ""},
	}

	writer := &bytes.Buffer{}
	err = adapter.ToCSV(writer, slices.Values(people))
	if err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	expected := `name,age,email
John Doe,30,` + fakemail + `
Jane Smith,25,
`
	if writer.String() != expected {
		t.Errorf("expected %s, got %s", expected, writer.String())
	}
}

func TestToCSVWithManyTypes(t *testing.T) {
	adapter, err := NewCSVAdapter[PersonWithManyTypes]()
	if err != nil {
		t.Fatalf("failed to create csva: %v", err)
	}

	people := []PersonWithManyTypes{
		{
			"John Doe",
			30,
			&PersonWithManyTypesEmail{fakemail},
			3.14,
			true,
			stringPtr("hello"),
		},
		{
			"Jane Smith",
			25,
			&PersonWithManyTypesEmail{otherfakemail},
			2.71,
			false,
			stringPtr("123"),
		},
	}

	writer := &bytes.Buffer{}
	err = adapter.ToCSV(writer, slices.Values(people))
	if err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	expected := `name,age,email,some_float,some_bool,some_ptr
John Doe,30,` + fakemail + `,3.140000,true,hello
Jane Smith,25,` + otherfakemail + `,2.710000,false,123
`
	if writer.String() != expected {
		t.Errorf("expected\n%s, got\n%s", expected, writer.String())
	}
}

func TestToCSVWithNoTags(t *testing.T) {
	adapter, err := NewCSVAdapter[PersonNoTags]()
	if err != nil {
		t.Fatalf("failed to create csva: %v", err)
	}

	people := []PersonNoTags{
		{"John Doe", 30},
		{"Jane Smith", 25},
	}

	writer := &bytes.Buffer{}
	err = adapter.ToCSV(writer, slices.Values(people))
	if err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	expected := `Name,Age
John Doe,30
Jane Smith,25
`
	if writer.String() != expected {
		t.Errorf("expected %s, got %s", expected, writer.String())
	}
}

func TestToCSVWithNoTagsAndOmitEmpty(t *testing.T) {
	adapter, err := NewCSVAdapter[PersonNoTags]()
	if err != nil {
		t.Fatalf("failed to create csva: %v", err)
	}

	people := []PersonNoTags{
		{"John Doe", 30},
		{"Jane Smith", 0},
	}

	writer := &bytes.Buffer{}
	err = adapter.ToCSV(writer, slices.Values(people))
	if err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	expected := `Name,Age
John Doe,30
Jane Smith,0
`
	if writer.String() != expected {
		t.Errorf("expected %s, got %s", expected, writer.String())
	}
}

// Test data
const (
	fakemail      = "fakemail@mail.com"
	otherfakemail = "otherfakenail@mail.com"
	name          = "John Doe"
	othername     = "Jane Smith"
	age           = 30
	otherage      = 25
)

func stringPtr(s string) *string {
	return &s
}
