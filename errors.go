package csvadapter

import (
	"fmt"
)

type ReadingError struct {
	Line       int
	Field      string
	FieldAlias string
}

func (r ReadingError) Error() string {
	return fmt.Sprintf(
		"error reading field %s (%s) at line %d",
		r.Field,
		r.FieldAlias,
		r.Line,
	)
}
