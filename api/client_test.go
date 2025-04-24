package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// has the correct username and password
func setupClient(t *testing.T, url string) *Client {
	return setupCustomClient(t, url, "testuser", "testpassword")
}

func setupCustomClient(t *testing.T, url, username, password string) *Client {
	t.Setenv("GATEWAY_USERNAME", username)
	t.Setenv("GATEWAY_PASSWORD", password)
	return NewClient(url, nil)
}

func setupServer(tokenExpiration time.Time) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/login" {
			if r.Method != "POST" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			var loginRequest struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			err := json.NewDecoder(r.Body).Decode(&loginRequest)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if loginRequest.Username != "testuser" || loginRequest.Password != "testpassword" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			authResponse := authResponse{
				Auth: authToken{
					Token:      "testtoken",
					Expiration: tokenExpiration.Unix(),
				},
			}
			json.NewEncoder(w).Encode(authResponse)
		} else if r.URL.Path == "/gateway/" {
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if r.Header.Get("Authorization") != "Bearer testtoken" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			gatewayResponse := GatewayResponse{}
			json.NewEncoder(w).Encode(gatewayResponse)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestLogin(t *testing.T) {

	tokenExpiration := time.Now().Add(15 * time.Minute)
	srv := setupServer(tokenExpiration)
	defer srv.Close()

	t.Run("Success", func(t *testing.T) {
		client := setupClient(t, srv.URL)
		err := client.Login(context.Background())
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		} else if client.auth == nil {
			t.Error("Expected auth token to be set")
		} else if client.auth.Token != "testtoken" {
			t.Errorf("Expected auth token to be 'testtoken', got '%s'", client.auth.Token)
			return
		} else if client.auth.Expiration != tokenExpiration.Unix() {
			t.Errorf("Expected auth expiration to be %d, got %d", tokenExpiration.Unix(), client.auth.Expiration)
		}
	})
	t.Run("Invalid Username", func(t *testing.T) {
		client := setupCustomClient(t, srv.URL, "invaliduser", "testpassword")
		err := client.Login(context.Background())
		if err == nil {
			t.Error("Expected error for invalid username")
		}
	})
	t.Run("Invalid Password", func(t *testing.T) {
		client := setupCustomClient(t, srv.URL, "testuser", "invalidpassword")
		err := client.Login(context.Background())
		if err == nil {
			t.Error("Expected error for invalid password")
		}
	})
}

func TestGetGateway(t *testing.T) {
	srv := setupServer(time.Now().Add(15 * time.Minute))
	defer srv.Close()

	t.Run("Success", func(t *testing.T) {
		client := setupClient(t, srv.URL)
		_, err := client.GetGateway(context.Background())
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
	})

	// Expect the token to be refreshed if expired and the gateway request to succeed.
	t.Run("Expired Token Refreshed Success", func(t *testing.T) {
		client := setupClient(t, srv.URL)
		client.auth = &authToken{
			Token:      "testtoken",
			Expiration: time.Now().Add(-1 * time.Minute).Unix(),
		}
		_, err := client.GetGateway(context.Background())
		if err != nil {
			t.Error("Expected expired token to request to succeed")
		}
	})

}
