package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCamel(t *testing.T) {
	testCases := map[string]struct {
		In       string
		Expected string
	}{
		"simple": {
			In:       "hello",
			Expected: "Hello",
		},
		"compound": {
			In:       "hello_world",
			Expected: "HelloWorld",
		},
		"double _": {
			In:       "hello__world",
			Expected: "HelloWorld",
		},
		"many": {
			In:       "a_b_c",
			Expected: "ABC",
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			assert.Equal(t, tc.Expected, camel(tc.In))
		})
	}
}
