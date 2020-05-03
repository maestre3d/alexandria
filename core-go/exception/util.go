package exception

import (
	"fmt"
	"strings"
)

// NewErrorDescription Attach/Wrap an error to a description, separated by '~' *Uses fmt package's error wrap
func NewErrorDescription(err error, description string) error {
	return fmt.Errorf("%w~%s", err, description)
}

// GetErrorDescription Obtain wrapped error's description
func GetErrorDescription(err error) string {
	errs := strings.Split(err.Error(), "~")
	if len(errs) > 1 {
		return errs[len(errs)-1]
	}

	return err.Error()
}
