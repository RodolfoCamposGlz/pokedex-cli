package main

import (
	"testing"
)


func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		// add more casees here
	}
	
    for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) <= 0{
			t.Errorf("Not matching word")
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord{
				t.Errorf("Not matching word")
			}
		}
	}
}