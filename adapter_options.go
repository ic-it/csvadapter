package csvadapter

import "encoding/csv"

func newCSVAdapterOptions() *csvAdapterOptions {
	return &csvAdapterOptions{
		// default encoding/csv options
		comma: ',',

		// default other options
		writeHeader:     true,
		noImplicitAlias: false,
	}
}

// csvAdapterOption is a function that sets an option on the CSVAdapter
type csvAdapterOption func(*csvAdapterOptions)

// Comma sets the field separator
//
// more info: https://pkg.go.dev/encoding/csv#Reader and https://pkg.go.dev/encoding/csv#Writer
func Comma(r rune) csvAdapterOption {
	return func(o *csvAdapterOptions) {
		o.comma = r
	}
}

// Comment sets the comment character
//
// more info: https://pkg.go.dev/encoding/csv#Reader
func Comment(r rune) csvAdapterOption {
	return func(o *csvAdapterOptions) {
		o.comment = r
	}
}

// LazyQuotes sets the lazy quotes flag
//
// more info: https://pkg.go.dev/encoding/csv#Reader
func LazyQuotes(lazyQuotes bool) csvAdapterOption {
	return func(o *csvAdapterOptions) {
		o.lazyQuotes = lazyQuotes
	}
}

// TrimLeadingSpace sets the trim leading space flag
//
// more info: https://pkg.go.dev/encoding/csv#Reader
func TrimLeadingSpace(trimLeadingSpace bool) csvAdapterOption {
	return func(o *csvAdapterOptions) {
		o.trimLeadingSpace = trimLeadingSpace
	}
}

// ReuseRecord sets the reuse record flag
//
// more info: https://pkg.go.dev/encoding/csv#Reader
func ReuseRecord(reuseRecord bool) csvAdapterOption {
	return func(o *csvAdapterOptions) {
		o.reuseRecord = reuseRecord
	}
}

// sets the use CRLF flag.
//
// more info: https://pkg.go.dev/encoding/csv#Writer
func UseCRLF(useCRLF bool) csvAdapterOption {
	return func(o *csvAdapterOptions) {
		o.useCRLF = useCRLF
	}
}

// sets the write header flag
//
// when set to true, the header will be written when calling ToCSV
func WriteHeader(writeHeader bool) csvAdapterOption {
	return func(o *csvAdapterOptions) {
		o.writeHeader = writeHeader
	}
}

// sets the no implicit alias flag
//
// when set to true, field names will not be used as aliases when not specified.
func NoImplicitAlias(noImplicitAlias bool) csvAdapterOption {
	return func(o *csvAdapterOptions) {
		o.noImplicitAlias = noImplicitAlias
	}
}

type csvAdapterOptions struct {
	// encoding/csv options
	comma            rune
	comment          rune
	lazyQuotes       bool
	trimLeadingSpace bool
	reuseRecord      bool
	useCRLF          bool

	// other options
	writeHeader     bool
	noImplicitAlias bool
}

func (c csvAdapterOptions) applyReader(reader *csv.Reader) {
	reader.Comma = c.comma
	reader.Comment = c.comment
	reader.LazyQuotes = c.lazyQuotes
	reader.TrimLeadingSpace = c.trimLeadingSpace
	reader.ReuseRecord = c.reuseRecord
}

func (c csvAdapterOptions) applyWriter(writer *csv.Writer) {
	writer.Comma = c.comma
	writer.UseCRLF = c.useCRLF
}
