package omlox

import "fmt"

// parameterToString is a very simple function that returns the string
// representation of the given parameter value.
func parameterToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}
