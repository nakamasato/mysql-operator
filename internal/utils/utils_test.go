package utils

import "testing"

func TestGenerateRandomString(t *testing.T) {
	t.Run("Generated string length equals to the given value", func(t *testing.T) {
		expectedLength := 10
		random := GenerateRandomString(expectedLength)
		if len(random) != expectedLength {
			t.Errorf("Length was expected to be %d, but got %d", expectedLength, len(random))
		}
	})
}
