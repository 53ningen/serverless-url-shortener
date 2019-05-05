package models

import (
	"testing"
)

func TestHandler(t *testing.T) {
	t.Run("calc getSHA256", func(t *testing.T) {
		dict := make(map[string]string)
		dict[""] = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		dict["http://example.com"] = "f0e6a6a97042a4f1f1c87f5f7d44315b2d852c2df5c7991cc66241bf7072d1c4"

		for key, value := range dict {
			actual := getSHA256(key)
			if actual != value {
				t.Fatalf("getMD5Hash failure: expected %s, actual %s", value, actual)
			}
		}
	})
}
