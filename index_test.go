package main

import (
	"testing"

	"github.com/anton2920/gofa/net/http"
)

func TestIndexPageHandler(t *testing.T) {
	const endpoint = "/"

	testGet(t, endpoint, http.StatusOK)

	t.Run("expectedOK", func(t *testing.T) {
		for _, token := range testTokens {
			token := token
			t.Run("", func(t *testing.T) {
				t.Parallel()
				testGetAuth(t, endpoint, token, http.StatusOK)
			})
		}
	})
}
