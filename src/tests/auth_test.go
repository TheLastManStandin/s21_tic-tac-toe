package tests_test

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

const missingUUID = "00000000-0000-0000-0000-000000000099"

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

func TestRegistrationAndAuthentication(t *testing.T) {
	api := newAPIClient(t)
	userUUID := api.register(t)

	t.Run("login returns registered UUID", func(t *testing.T) {
		got := strings.TrimSpace(string(api.request(t, http.MethodGet, "/auth", nil)))
		if got != userUUID {
			t.Fatalf("authenticated UUID = %q, want %q", got, userUUID)
		}
	})
	t.Run("duplicate registration", func(t *testing.T) {
		api.request(t, http.MethodPost, "/register", nil, http.StatusConflict)
		if got := strings.TrimSpace(string(api.request(t, http.MethodGet, "/auth", nil))); got != userUUID {
			t.Fatal("duplicate registration changed the existing user")
		}
	})
	t.Run("wrong password", func(t *testing.T) {
		wrong := api
		wrong.password = "wrong-password"
		wrong.request(t, http.MethodGet, "/auth", nil, http.StatusUnauthorized)
	})
	t.Run("unknown user", func(t *testing.T) {
		newAPIClient(t).request(t, http.MethodGet, "/auth", nil, http.StatusUnauthorized)
	})
	t.Run("missing credentials", func(t *testing.T) {
		anonymous := api
		anonymous.username, anonymous.password = "", ""
		anonymous.request(t, http.MethodGet, "/auth", nil, http.StatusUnauthorized)
		anonymous.request(t, http.MethodPost, "/register", nil, http.StatusUnauthorized)
	})
}

func TestProtectedEndpointsRequireAuthentication(t *testing.T) {
	api := newAPIClient(t)
	api.register(t)
	for _, credentials := range []string{"missing", "wrong password"} {
		t.Run(credentials, func(t *testing.T) {
			unauthorized := api
			if credentials == "missing" {
				unauthorized.username, unauthorized.password = "", ""
			} else {
				unauthorized.password = "wrong-password"
			}
			for _, route := range []struct {
				method, path string
				body         any
			}{
				{http.MethodPost, "/game", map[string]bool{"WithRobot": true}},
				{http.MethodPost, "/game/" + missingUUID, turnRequest{X: 0, Y: 0}},
				{http.MethodPost, "/game/" + missingUUID + "/join", nil},
				{http.MethodGet, "/game", nil},
				{http.MethodGet, "/game/" + missingUUID, nil},
				{http.MethodGet, "/user/" + missingUUID, nil},
			} {
				t.Run(route.method+" "+route.path, func(t *testing.T) {
					unauthorized.request(t, route.method, route.path, route.body, http.StatusUnauthorized)
				})
			}
		})
	}
}

func TestUserStatus(t *testing.T) {
	api := newAPIClient(t)
	userUUID := api.register(t)
	other := newAPIClient(t)
	otherUUID := other.register(t)
	for _, user := range []struct{ id, username string }{
		{userUUID, api.username}, {otherUUID, other.username},
	} {
		t.Run(user.id, func(t *testing.T) {
			data := api.request(t, http.MethodGet, "/user/"+user.id, nil)
			var profile map[string]string
			if err := json.Unmarshal(data, &profile); err != nil {
				t.Fatal(err)
			}
			if len(profile) != 2 || profile["uuid"] != user.id || profile["username"] != user.username {
				t.Fatalf("profile must contain only UUID and username: %v", profile)
			}
		})
	}
	t.Run("missing user", func(t *testing.T) {
		api.request(t, http.MethodGet, "/user/"+missingUUID, nil, http.StatusNotFound)
	})
	t.Run("invalid UUID", func(t *testing.T) {
		api.request(t, http.MethodGet, "/user/invalid", nil, http.StatusBadRequest)
	})
}
