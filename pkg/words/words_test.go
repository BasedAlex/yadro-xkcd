package words

import (
	"errors"
	"testing"

	"github.com/go-playground/assert/v2"
)


func TestSteminator(t *testing.T) {
	tests := []struct {
		input         string
		mockError     error
		expectedOutput []string
		expectedError  error
	}{
		{
			input:         "is a test string.",
			mockError:     nil,
			expectedOutput: []string{"test", "string"},
			expectedError:  nil,
		},
		{
			input:         "",
			mockError:     errors.New("please provide a string to be stemmed"),
			expectedOutput: nil,
			expectedError:  errors.New("please provide a string to be stemmed"),
		},
		{
			input:         "she",
			mockError:     errors.New("result is empty, please provide a better string"),
			expectedOutput: nil,
			expectedError:  errors.New("result is empty, please provide a better string"),
		},
	}

	for _, test := range tests {
		output, err := Steminator(test.input)
		t.Log(output, err)
		assert.Equal(t, test.expectedOutput, output)
		assert.Equal(t, test.expectedError, err)
	}
}