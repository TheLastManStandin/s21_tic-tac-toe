package tests_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAuthLogin(t *testing.T) {
	api := newAPIClient(t)
	api.register(t)

	credentials := map[string]string{
		"login":    api.username,
		"password": api.password,
	}
	// На вход отправляем JSON, без заголовка Basic Auth.
	api.username, api.password = "", ""
	data := api.request(t, http.MethodPost, "/auth", credentials)

	var response struct {
		Type         string `json:"type"`
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("не удалось прочитать ответ: %v", err)
	}
	if response.Type != "Bearer" {
		t.Errorf("тип токена = %q, ожидался Bearer", response.Type)
	}
	if response.AccessToken == "" || response.RefreshToken == "" {
		t.Error("ответ должен содержать accessToken и refreshToken")
	}
}
