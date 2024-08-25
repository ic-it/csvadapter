# CSVAdapter

The `csvadapter` package provides a way to map CSV data into Go structs. This 
package allows you to easily read CSV files and convert them into a sequence of
structs. The package uses struct tags to map CSV columns to struct fields.

## Overview

The main component of this package is the `CSVAdapter` struct. It adapts a 
struct type to a CSV file by mapping CSV columns to struct fields based on 
tags defined in the struct.

## Installation

To install the package, run:
```sh
go get github.com/ic-it/csvadapter
```

## Usage

### Defining Structs

To use the `CSVAdapter`, define your struct with `csva` tags specifying 
the CSV column names:

```go
type Person struct {
    Name     string `csva:"name"`
    Age      int    `csva:"alias=age"`
    Email    string `csva:"email,omitempty"` // `omitempty` allows the field to be empty
    SomeDataToIgnore string `csva:"-"`
}
```

### Creating a CSVAdapter

Create a new `CSVAdapter` for your struct type:

```go
adapter, err := NewCSVAdapter[Person]()
if err != nil {
    log.Fatalf("failed to create CSVAdapter: %v", err)
}
```

#### Options

The `NewCSVAdapter` function supports the following options:

- `Comma(r rune)`: Sets the field separator. (default: `,`) ([more info](https://pkg.go.dev/encoding/csv#Reader) and [more info](https://pkg.go.dev/encoding/csv#Writer))
- `Comment(r rune)`: Sets the comment character. ([more info](https://pkg.go.dev/encoding/csv#Reader))
- `LazyQuotes(lazyQuotes bool)`: Sets the lazy quotes flag. ([more info](https://pkg.go.dev/encoding/csv#Reader))
- `TrimLeadingSpace(trimLeadingSpace bool)`: Sets the trim leading space flag. ([more info](https://pkg.go.dev/encoding/csv#Reader))
- `ReuseRecord(reuseRecord bool)`: Sets the reuse record flag. ([more info](https://pkg.go.dev/encoding/csv#Reader))
- `UseCRLF(useCRLF bool)`: Sets the use CRLF flag. ([more info](https://pkg.go.dev/encoding/csv#Writer))
- `WriteHeader(writeHeader bool)`: Sets the write header flag. When set to `true`, the header will be written when calling `ToCSV`.
- `NoImplicitAlias(noImplicitAlias bool)`: Sets the no implicit alias flag. When set to `true`, field names will not be used as aliases when not specified.

### Reading a CSV File

To read a CSV file and populate a slice of structs:

```go
file, err := os.Open("data.csv")
if err != nil {
    log.Fatalf("failed to open file: %v", err)
}
defer file.Close()


people, err := adapter.FromCSV(reader)
if err != nil {
    log.Fatalf("failed to read CSV: %v", err)
}

for person, err := range people {
    if err != nil {
        log.Fatalf("failed to read person: %v", err)
    }
    fmt.Printf("Person: %+v\n", person)
}
```

### Writing a CSV File

To write a slice of structs to a CSV file:

```go
file, err := os.Create("data.csv")
if err != nil {
    log.Fatalf("failed to create file: %v", err)
}
defer file.Close()

people := []Person{
    {Name: "Alice", Age: 30, Email: "testme@gmail.com"},
    {Name: "Bob", Age: 25, Email: "testmes2@gmail.com"},
}

if err := adapter.ToCSV(file, slices.Values(people)); err != nil {
    log.Fatalf("failed to write CSV: %v", err)
}

fmt.Println("CSV written successfully")
```

## CSVAdapter Type

The `CSVAdapter` type is a generic struct that adapts a Go struct to a CSV file:

### Allowed Types

The `CSVAdapter` supports the following types:

- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool`
- **Any type that implements the `encoding.TextUnmarshaler` interface**

## License

This project is licensed under the MIT License - see the LICENSE file for details.
