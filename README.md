# CSVAdapter

The `csvadapter` package provides a way to map CSV data into Go structs. This 
package allows you to easily read CSV files and convert them into a slice of structs.

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

To use the `CSVAdapter`, define your struct with `csvadapter` tags specifying 
the CSV column names:

```go
type Person struct {
    Name     string `csvadapter:"name"`
    Age      int    `csvadapter:"age"`
    Email    string `csvadapter:"email,omitempty"` // `omitempty` allows the field to be empty
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

### Reading a CSV File

To read a CSV file and populate a slice of structs:

```go
file, err := os.Open("data.csv")
if err != nil {
    log.Fatalf("failed to open file: %v", err)
}
defer file.Close()

var people []Person
if err := adapter.EatCSV(file, &people); err != nil {
    log.Fatalf("failed to read CSV: %v", err)
}

fmt.Printf("People: %+v\n", people)
```

## CSVAdapter Type

The `CSVAdapter` type is a generic struct that adapts a Go struct to a CSV file:

### Methods

- `NewCSVAdapter[T any]() (*CSVAdapter[T], error)`: Creates a new `CSVAdapter` 
    for the specified struct type.
- `EatCSV(reader io.Reader, v *[]T) error`: Reads a CSV file from the provided 
    `io.Reader` and populates the provided slice with the struct data.

### Example

Here is a complete example demonstrating how to use the `CSVAdapter`:

```go
package main

import (
    "log"
    "os"
    "github.com/ic-it/csvadapter"
)

type Person struct {
    Name     string `csvadapter:"name"`
    Age      int    `csvadapter:"age"`
    Email    string `csvadapter:"email,omitempty"`
}

func main() {
    adapter, err := csvadapter.NewCSVAdapter[Person]()
    if err != nil {
        log.Fatalf("failed to create CSVAdapter: %v", err)
    }

    file, err := os.Open("data.csv")
    if err != nil {
        log.Fatalf("failed to open file: %v", err)
    }
    defer file.Close()

    var people []Person
    if err := adapter.EatCSV(file, &people); err != nil {
        log.Fatalf("failed to read CSV: %v", err)
    }

    fmt.Printf("People: %+v\n", people)
}
```

## Error Handling

The `csvadapter` package defines several errors for handling various issues:

- `ErrUnsupportedTag`: Unsupported tag found in the struct definition.
- `ErrorNotStruct`: The provided type is not a struct.
- `ErrReadingCSV`: Error reading the CSV file.
- `ErrReadingCSVLines`: Error reading the lines of the CSV file.
- `ErrProcessingCSVLines`: Error processing the lines of the CSV file.
- `ErrFieldNotFound`: Field not found in the line.
- `ErrUnprocessableType`: The type of the field is unprocessable.
- `ErrParsingType`: Error parsing the type of the field.
- `ErrEmptyValue`: Encountered an empty value in a non-omitempty field.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
