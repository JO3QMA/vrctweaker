package automation

import "errors"

// ErrItemNotFound is returned when an automation item id does not exist.
var ErrItemNotFound = errors.New("automation item not found")
