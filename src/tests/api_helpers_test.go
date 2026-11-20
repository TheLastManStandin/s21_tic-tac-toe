package tests_test

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const missingUUID = "00000000-0000-0000-0000-000000000099"

type apiClient struct {
	baseURL      string
	httpClient   *http.Client
	username     string
	password     string
	userID       string
	accessToken  string
	refreshToken string
}

type tokenResponse struct {
	Type         string `json:"type"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type gameResponse struct {
	UUID             string    `json:"uuid"`
	Table            [3][3]int `json:"table"`
	State            string    `json:"state"`
	Minimax          bool      `json:"minimax"`
	FirstPlayerUUID  string    `json:"firstPlayerUUID"`
	SecondPlayerUUID string    `json:"secondPlayerUUID"`
}

type turnRequest struct {
	X int `json:"X"`
	Y int `json:"Y"`
}

func newAPIClient(t *testing.T) *apiClient {
	t.Helper()
	baseURL := strings.TrimRight(os.Getenv("E2E_BASE_URL"), "/")
	if baseURL == "" {
		t.Skip("запусти task e2e из src или задай E2E_BASE_URL")
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	t.Cleanup(client.CloseIdleConnections)
	return &apiClient{
		baseURL: baseURL, httpClient: client,
		username: "test-" + rand.Text()[:16],
		password: "test-password",
	}
}

func newAuthenticatedClient(t *testing.T) *apiClient {
	t.Helper()
	client := newAPIClient(t)
	client.register(t)
	client.login(t)
	return client
}

// request отправляет access-токен, если клиент уже вошёл в систему.
// Строка в body отправляется без изменений: это удобно для проверки плохого JSON.
// По умолчанию ожидается 200; другие допустимые статусы передаются явно.
func (api *apiClient) request(t *testing.T, method, path string, body any, statuses ...int) []byte {
	t.Helper()
	authorization := ""
	if api.accessToken != "" {
		authorization = "Bearer " + api.accessToken
	}
	return api.requestWithAuth(t, method, path, authorization, body, statuses...)
}

func (api *apiClient) requestWithAuth(t *testing.T, method, path, authorization string, body any, statuses ...int) []byte {
	t.Helper()
	var payload []byte
	if raw, ok := body.(string); ok {
		payload = []byte(raw)
	} else if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("подготовка JSON: %v", err)
		}
	}
	req, err := http.NewRequestWithContext(t.Context(), method, api.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	res, err := api.httpClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) == 0 {
		statuses = []int{http.StatusOK}
	}
	t.Logf("%s %s -> %d", method, path, res.StatusCode)
	for _, status := range statuses {
		if res.StatusCode == status {
			return data
		}
	}
	t.Fatalf("%s %s: статус %d, ожидался %v; ответ: %.1000s", method, path, res.StatusCode, statuses, data)
	return nil
}

func (api *apiClient) basicAuth() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(api.username+":"+api.password))
}

func (api *apiClient) register(t *testing.T) string {
	t.Helper()
	data := api.requestWithAuth(t, http.MethodPost, "/register", api.basicAuth(), nil, http.StatusOK, http.StatusCreated)
	api.userID = strings.TrimSpace(string(data))
	assertUUID(t, api.userID)
	return api.userID
}

func (api *apiClient) login(t *testing.T) tokenResponse {
	t.Helper()
	data := api.requestWithAuth(t, http.MethodPost, "/auth", "", map[string]string{
		"login": api.username, "password": api.password,
	})
	tokens := readTokens(t, data)
	api.accessToken, api.refreshToken = tokens.AccessToken, tokens.RefreshToken
	return tokens
}

func (api *apiClient) refresh(t *testing.T) tokenResponse {
	t.Helper()
	// Обновление должно работать без действующего access-токена.
	data := api.requestWithAuth(t, http.MethodPost, "/refresh", "", map[string]string{
		"refreshToken": api.refreshToken,
	})
	tokens := readTokens(t, data)
	api.accessToken, api.refreshToken = tokens.AccessToken, tokens.RefreshToken
	return tokens
}

func readTokens(t *testing.T, data []byte) tokenResponse {
	t.Helper()
	var tokens tokenResponse
	decodeJSON(t, data, &tokens)
	if tokens.Type != "Bearer" || tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("ожидался тип Bearer и два непустых токена")
	}
	if tokens.AccessToken == tokens.RefreshToken {
		t.Fatal("access и refresh не должны совпадать")
	}
	return tokens
}

func (api *apiClient) createGame(t *testing.T, withRobot bool) string {
	t.Helper()
	data := api.request(t, http.MethodPost, "/game", map[string]bool{"WithRobot": withRobot}, http.StatusOK, http.StatusCreated)
	id := strings.TrimSpace(string(data))
	assertUUID(t, id)
	return id
}

func (api *apiClient) joinGame(t *testing.T, id string) {
	t.Helper()
	api.request(t, http.MethodPost, "/game/"+id+"/join", nil)
}

func (api *apiClient) getGame(t *testing.T, id string) gameResponse {
	t.Helper()
	var game gameResponse
	decodeJSON(t, api.request(t, http.MethodGet, "/game/"+id, nil), &game)
	return game
}

func (api *apiClient) makeTurn(t *testing.T, id string, x, y int) gameResponse {
	t.Helper()
	var game gameResponse
	decodeJSON(t, api.request(t, http.MethodPost, "/game/"+id, turnRequest{X: x, Y: y}), &game)
	return game
}

func (api *apiClient) availableGames(t *testing.T) []gameResponse {
	t.Helper()
	var games []gameResponse
	decodeJSON(t, api.request(t, http.MethodGet, "/game", nil), &games)
	if games == nil {
		t.Fatal("список игр должен быть JSON-массивом, а не null")
	}
	return games
}

func containsGame(games []gameResponse, id string) bool {
	for _, game := range games {
		if game.UUID == id {
			return true
		}
	}
	return false
}

func decodeJSON(t *testing.T, data []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("чтение JSON-ответа: %v", err)
	}
}

func assertUUID(t *testing.T, id string) {
	t.Helper()
	if err := uuid.Validate(id); err != nil {
		t.Fatalf("некорректный UUID %q: %v", id, err)
	}
}

func requireShortTTL(t *testing.T, name string) {
	t.Helper()
	value := os.Getenv(name)
	if value == "" {
		t.Skip("проверка истечения токена запускается через task e2e-auth с короткими TTL")
	}
	ttl, err := time.ParseDuration(value)
	if err != nil || ttl < time.Second || ttl > 30*time.Second {
		t.Fatalf("%s должен быть от 1s до 30s", name)
	}
}

// JWT здесь только декодируется, чтобы определить момент ожидания.
// Проверку подписи и истечения срока выполняет приложение через HTTP-запрос.
func waitForTokenExpiration(t *testing.T, token string) {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatal("ожидался JWT из трёх частей")
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var claims struct {
		ExpiresAt int64 `json:"exp"`
	}
	decodeJSON(t, data, &claims)
	if claims.ExpiresAt == 0 {
		t.Fatal("в токене отсутствует exp")
	}
	wait := time.Until(time.Unix(claims.ExpiresAt, 0).Add(250 * time.Millisecond))
	if wait > 31*time.Second {
		t.Fatal("слишком большой TTL: задай короткий срок в контейнере приложения")
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-t.Context().Done():
		t.Fatal("ожидание отменено")
	}
}
