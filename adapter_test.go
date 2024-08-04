package csvadapter

import (
	"bytes"
	"errors"
	"testing"
)

type Person struct {
	Name  string `csvadapter:"name"`
	Age   int    `csvadapter:"age"`
	Email string `csvadapter:"email,omitempty"`
}

type PersonNoTags struct {
	Name string `csvadapter:""`
	Age  int
}

type PersonWrongTag struct {
	Name string `csvadapter:"name,omitempty,foo"`
	Age  int
}

type PersonWithManyTypes struct {
	Name      string                   `csvadapter:"name"`
	Age       int                      `csvadapter:"age"`
	Email     PersonWithManyTypesEmail `csvadapter:"email,omitempty"`
	SomeFloat float64                  `csvadapter:"some_float"`
	SomeBool  bool                     `csvadapter:"some_bool"`
	SomePtr   *string                  `csvadapter:"some_ptr"`
}

type PersonWithManyTypesEmail struct {
	Email string
}

func (p *PersonWithManyTypesEmail) UnmarshalText(text []byte) error {
	p.Email = string(text)
	return nil
}

func TestNewCSVAdapter(t *testing.T) {
	t.Run("with tags", func(t *testing.T) {
		adapter, err := NewCSVAdapter[Person]()
		if err != nil {
			t.Fatalf("failed to create CSVAdapter: %v", err)
		}
		if adapter.String() != "CSVAdapter(Person)" {
			t.Errorf("expected CSVAdapter(Person), got %s", adapter.String())
		}
		if len(adapter.fields) != 3 {
			t.Errorf("expected 3 fields, got %d", len(adapter.fields))
		}
	})
	t.Run("without tags", func(t *testing.T) {
		adapter, err := NewCSVAdapter[PersonNoTags]()
		if err != nil {
			t.Fatalf("failed to create CSVAdapter: %v", err)
		}
		if len(adapter.fields) != 0 {
			t.Errorf("expected 0 fields, got %d", len(adapter.fields))
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

func TestEatCSV(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		csvData := `name,age,email
John Doe,30,john@example.com
Jane Smith,25,jane@example.com
`

		reader := bytes.NewReader([]byte(csvData))
		adapter, err := NewCSVAdapter[Person]()
		if err != nil {
			t.Fatalf("failed to create CSVAdapter: %v", err)
		}

		var people []Person
		if err := adapter.EatCSV(reader, &people); err != nil {
			t.Fatalf("failed to read CSV: %v", err)
		}

		if len(people) != 2 {
			t.Fatalf("expected 2 people, got %d", len(people))
		}

		expected := []Person{
			{"John Doe", 30, "john@example.com"},
			{"Jane Smith", 25, "jane@example.com"},
		}

		for i, person := range people {
			if person != expected[i] {
				t.Errorf("expected %+v, got %+v", expected[i], person)
			}
		}
	})

	t.Run("empty file", func(t *testing.T) {
		csvData := ``
		reader := bytes.NewReader([]byte(csvData))
		adapter, err := NewCSVAdapter[Person]()
		if err != nil {
			t.Fatalf("failed to create CSVAdapter: %v", err)
		}

		var people []Person
		if err := adapter.EatCSV(reader, &people); err == nil {
			t.Fatalf("expected error, got nil")
		}

		if len(people) != 0 {
			t.Errorf("expected 0 people, got %d", len(people))
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
			t.Fatalf("failed to create CSVAdapter: %v", err)
		}

		var people []Person
		if err := adapter.EatCSV(reader, &people); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

}

func TestEatCSVWithOmitEmpty(t *testing.T) {
	t.Run("omit empty", func(t *testing.T) {
		csvData := `name,age,email
John Doe,30,
Jane Smith,25,jane@example.com`

		reader := bytes.NewReader([]byte(csvData))
		adapter, err := NewCSVAdapter[Person]()
		if err != nil {
			t.Fatalf("failed to create CSVAdapter: %v", err)
		}

		var people []Person
		if err := adapter.EatCSV(reader, &people); err != nil {
			t.Fatalf("failed to read CSV: %v", err)
		}

		if len(people) != 2 {
			t.Fatalf("expected 2 people, got %d", len(people))
		}

		expected := []Person{
			{"John Doe", 30, ""},
			{"Jane Smith", 25, "jane@example.com"},
		}

		for i, person := range people {
			if person != expected[i] {
				t.Errorf("expected %+v, got %+v", expected[i], person)
			}
		}
	})

	t.Run("unomit empty", func(t *testing.T) {
		csvData := `name,age,email
John Doe,,
Jane Smith,25,`

		reader := bytes.NewReader([]byte(csvData))
		adapter, err := NewCSVAdapter[Person]()
		if err != nil {
			t.Fatalf("failed to create CSVAdapter: %v", err)
		}

		var people []Person
		err = adapter.EatCSV(reader, &people)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrEmptyValue) {
			t.Errorf("expected ErrEmptyValue, got %v", err)
		}

	})

}

func TestEatCSVWithMissingField(t *testing.T) {
	csvData := `name,age
John Doe,30
Jane Smith,25
`

	reader := bytes.NewReader([]byte(csvData))
	adapter, err := NewCSVAdapter[Person]()
	if err != nil {
		t.Fatalf("failed to create CSVAdapter: %v", err)
	}

	var people []Person
	err = adapter.EatCSV(reader, &people)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrFieldNotFound) {
		t.Errorf("expected ErrFieldNotFound, got %v", err)
	}
}

func TestEatCSVWithInvalidData(t *testing.T) {
	csvData := `name,age,email
John Doe,thirty,john@example.com
Jane Smith,25,jane@example.com
`

	reader := bytes.NewReader([]byte(csvData))
	adapter, err := NewCSVAdapter[Person]()
	if err != nil {
		t.Fatalf("failed to create CSVAdapter: %v", err)
	}

	var people []Person
	err = adapter.EatCSV(reader, &people)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrParsingType) {
		t.Errorf("expected ErrParsingType, got %v", err)
	}
}

func TestEatCSVWithManyTypes(t *testing.T) {
	csvData := `name,age,email,some_float,some_bool,some_ptr
John Doe,30,test@mail.com,3.14,true,hello
Jane Smith,25,12@mail.com,2.71,false,123
`

	reader := bytes.NewReader([]byte(csvData))
	adapter, err := NewCSVAdapter[PersonWithManyTypes]()
	if err != nil {
		t.Fatalf("failed to create CSVAdapter: %v", err)
	}

	var people []PersonWithManyTypes
	err = adapter.EatCSV(reader, &people)
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	if len(people) != 2 {
		t.Fatalf("expected 2 people, got %d", len(people))
	}

	stringPtr := func(s string) *string {
		return &s
	}

	expected := []PersonWithManyTypes{
		{
			"John Doe",
			30,
			PersonWithManyTypesEmail{"test@mail.com"},
			3.14,
			true,
			stringPtr("hello"),
		},
		{
			"Jane Smith",
			25,
			PersonWithManyTypesEmail{"12@mail.com"},
			2.71,
			false,
			stringPtr("123"),
		},
	}

	for i, person := range people {
		if person.Name != expected[i].Name ||
			person.Age != expected[i].Age ||
			person.Email.Email != expected[i].Email.Email ||
			person.SomeFloat != expected[i].SomeFloat ||
			person.SomeBool != expected[i].SomeBool ||
			*person.SomePtr != *expected[i].SomePtr {
			t.Errorf("expected %+v, got %+v", expected[i], person)
		}
	}
}
