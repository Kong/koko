package persistence

import (
	"fmt"
)

// ErrInvalidRowsAffected is used when the number of affected rows does not match the
// expected amount. This error should usually not be seen under normal circumstances.
var ErrInvalidRowsAffected = fmt.Errorf("invalid rows affected")

// ErrNotFound is used to indicate the key being looked up was not found in the datastore.
type ErrNotFound struct {
	Key string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%v not found", e.Key)
}
