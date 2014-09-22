package env

import (
	"os"
	"testing"
)

// TestIsDebug verifies that IsDebug is working properly.
func TestIsDebug(t *testing.T) {
	var tests = []struct {
		value    string
		expected bool
	}{
		// Garbage values
		{envTest, false},
		{"HELLO", false},
		{"WORLD", false},

		// Correct values
		{disabled, false},
		{enabled, true},
	}

	for i, test := range tests {
		if err := os.Setenv(envDebug, test.value); err != nil {
			t.Fatal(err)
		}

		if IsDebug() != test.expected {
			t.Fatalf("[%02d] unexpected environment variable value and expected value: %v", i, test)
		}
	}
}

// TestIsTest verifies that IsTest is working properly.
func TestIsTest(t *testing.T) {
	var tests = []struct {
		value    string
		expected bool
	}{
		// Garbage values
		{envDebug, false},
		{"HELLO", false},
		{"WORLD", false},

		// Correct values
		{disabled, false},
		{enabled, true},
	}

	for i, test := range tests {
		if err := os.Setenv(envTest, test.value); err != nil {
			t.Fatal(err)
		}

		if IsTest() != test.expected {
			t.Fatalf("[%02d] unexpected environment variable value and expected value: %v", i, test)
		}
	}
}

// TestSetDebug verifies that SetDebug is working properly.
func TestSetDebug(t *testing.T) {
	var tests = []struct {
		value bool
	}{
		// Correct values
		{true},
		{false},
	}

	for _, test := range tests {
		if err := SetDebug(test.value); err != nil {
			t.Fatal(err)
		}
	}
}

// TestSetTest verifies that SetTest is working properly.
func TestSetTest(t *testing.T) {
	var tests = []struct {
		value bool
	}{
		// Correct values
		{true},
		{false},
	}

	for _, test := range tests {
		if err := SetTest(test.value); err != nil {
			t.Fatal(err)
		}
	}
}
