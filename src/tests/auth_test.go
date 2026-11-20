package tests_test

import (
	"net/http"
	"testing"
)

func TestAuthRegistration(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := newAPIClient(t)
		client.register(t)
		client.login(t)
	})
	t.Run("duplicate_login", func(t *testing.T) {
		client := newAPIClient(t)
		client.register(t)
		client.requestWithAuth(t, http.MethodPost, "/register", client.basicAuth(), nil, http.StatusConflict)
		// Неудачная повторная регистрация не ломает существующий аккаунт.
		client.login(t)
	})
	t.Run("missing_credentials", func(t *testing.T) {
		client := newAPIClient(t)
		client.request(t, http.MethodPost, "/register", nil, http.StatusUnauthorized)
	})
}

func TestAuthLogin(t *testing.T) {
	client := newAuthenticatedClient(t)
	// Проверяем не только выдачу токенов, но и выполнение защищённого запроса.
	client.availableGames(t)
}

func TestAuthLoginErrors(t *testing.T) {
	for _, name := range []string{"wrong_password", "unknown_user", "empty_password", "invalid_json"} {
		t.Run(name, func(t *testing.T) {
			client := newAPIClient(t)
			client.register(t)
			credentials := map[string]string{"login": client.username, "password": client.password}
			var body any = credentials
			status := http.StatusUnauthorized
			switch name {
			case "wrong_password":
				credentials["password"] = "wrong-password"
			case "unknown_user":
				credentials["login"] = "missing-" + client.username
			case "empty_password":
				credentials["password"] = ""
			case "invalid_json":
				body, status = `{"login":`, http.StatusBadRequest
			}
			client.request(t, http.MethodPost, "/auth", body, status)
		})
	}
}

func TestAuthProtectedRequests(t *testing.T) {
	client := newAuthenticatedClient(t)
	for _, tc := range []struct{ name, authorization string }{
		{"missing_token", ""},
		{"invalid_token", "Bearer invalid"},
		{"refresh_instead_of_access", "Bearer " + client.refreshToken},
		{"basic_instead_of_bearer", client.basicAuth()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client.requestWithAuth(t, http.MethodGet, "/game", tc.authorization, nil, http.StatusUnauthorized)
			client.requestWithAuth(t, http.MethodPost, "/game", tc.authorization, map[string]bool{"WithRobot": true}, http.StatusUnauthorized)
		})
	}
}

func TestAuthCurrentUser(t *testing.T) {
	t.Run("with_access_token", func(t *testing.T) {
		client := newAuthenticatedClient(t)
		var userID string
		decodeJSON(t, client.request(t, http.MethodGet, "/user", nil), &userID)
		if userID != client.userID {
			t.Errorf("UUID пользователя = %q, ожидался %q", userID, client.userID)
		}
	})
	t.Run("without_access_token", func(t *testing.T) {
		client := newAPIClient(t)
		client.request(t, http.MethodGet, "/user", nil, http.StatusUnauthorized)
	})
}

func TestAuthUserProfile(t *testing.T) {
	client := newAuthenticatedClient(t)
	var profile map[string]string
	decodeJSON(t, client.request(t, http.MethodGet, "/user/"+client.userID, nil), &profile)
	if profile["uuid"] != client.userID || profile["username"] != client.username {
		t.Errorf("неверные данные пользователя: %v", profile)
	}
	if _, exists := profile["password"]; exists {
		t.Error("ответ не должен содержать пароль или его хеш")
	}
}

func TestAuthRefresh(t *testing.T) {
	client := newAuthenticatedClient(t)
	client.refresh(t)
	client.availableGames(t)
}

func TestAuthRefreshErrors(t *testing.T) {
	for _, name := range []string{"invalid_token", "empty_token", "access_instead_of_refresh"} {
		t.Run(name, func(t *testing.T) {
			client := newAuthenticatedClient(t)
			token := "invalid"
			if name == "empty_token" {
				token = ""
			} else if name == "access_instead_of_refresh" {
				token = client.accessToken
			}
			client.requestWithAuth(t, http.MethodPost, "/refresh", "", map[string]string{"refreshToken": token}, http.StatusUnauthorized)
		})
	}
}

func TestAuthAccessExpiration(t *testing.T) {
	requireShortTTL(t, "E2E_AUTH_ACCESS_TTL")
	client := newAuthenticatedClient(t)
	client.availableGames(t)
	oldAccess := client.accessToken
	waitForTokenExpiration(t, oldAccess)
	client.request(t, http.MethodGet, "/game", nil, http.StatusUnauthorized)

	client.refresh(t)
	client.availableGames(t)
	// Выдача нового access не должна снова делать старый действительным.
	client.requestWithAuth(t, http.MethodGet, "/game", "Bearer "+oldAccess, nil, http.StatusUnauthorized)
}

func TestAuthRefreshExpiration(t *testing.T) {
	requireShortTTL(t, "E2E_AUTH_REFRESH_TTL")
	client := newAuthenticatedClient(t)
	client.refresh(t)
	waitForTokenExpiration(t, client.refreshToken)
	client.requestWithAuth(t, http.MethodPost, "/refresh", "", map[string]string{"refreshToken": client.refreshToken}, http.StatusUnauthorized)
	// После истечения refresh пользователь по-прежнему может войти заново.
	client.login(t)
	client.availableGames(t)
}
