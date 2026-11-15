package tests_test

import (
	"net/http"
	"regexp"
	"strings"
	"testing"
)

const missingUUID = "00000000-0000-0000-0000-000000000099"

// Общие вспомогательные функции для auth- и игровых тестов.
func (api apiClient) register(t *testing.T) string {
	t.Helper()
	id := strings.TrimSpace(string(api.request(t, http.MethodPost, "/register", nil)))
	assertUUID(t, id)
	return id
}

func assertUUID(t *testing.T, id string) {
	t.Helper()
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).MatchString(id) {
		t.Fatalf("invalid UUID in response: %q", id)
	}
}
