package crumbs

import (
	"errors"
	"fmt"
	"strings"
)

// FormatError returns a detailed string representation of err. The error
// message chain is always included. When includeCrumbs is true, the crumbs
// of the outermost *Error wrapper in the chain are appended on subsequent
// lines as "  key: value" pairs. Returns "" for a nil err.
func FormatError(err error, includeCrumbs bool) string {
	if err == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(err.Error())

	if !includeCrumbs {
		return sb.String()
	}

	var cerr *Error
	if errors.As(err, &cerr) {
		if cs := cerr.GetCrumbs(); len(cs) > 0 {
			sb.WriteString("\nCrumbs:")
			for _, crumb := range cs {
				sb.WriteString(fmt.Sprintf("\n  %s: %v", crumb.Key, crumb.Value))
			}
		}
	}

	return sb.String()
}
