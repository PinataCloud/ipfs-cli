package agents

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const (
	codexClientID    = "app_EMoamEEZ73f0CkXaXp7hrann"
	codexRedirectURI = "http://localhost:1455/auth/callback"
	codexAuthURL     = "https://auth.openai.com/oauth/authorize"
	codexTokenURL    = "https://auth.openai.com/oauth/token"
	codexSecretName  = "OPENAI_OAUTH_TOKEN"
)

type codexTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// --- PKCE helpers ---

func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateCodeChallenge(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func buildCodexAuthURL(challenge, state string) string {
	params := url.Values{
		"response_type":              {"code"},
		"client_id":                  {codexClientID},
		"redirect_uri":               {codexRedirectURI},
		"scope":                      {"openid profile email offline_access"},
		"code_challenge":             {challenge},
		"code_challenge_method":      {"S256"},
		"state":                      {state},
		"id_token_add_organizations": {"true"},
		"codex_cli_simplified_flow":  {"true"},
	}
	return codexAuthURL + "?" + params.Encode()
}

func openBrowser(u string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", u).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
	default:
		return exec.Command("xdg-open", u).Start()
	}
}

// --- Token exchange ---

func exchangeCodexToken(code, verifier string) (*codexTokenResponse, error) {
	params := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {codexRedirectURI},
		"client_id":     {codexClientID},
		"code_verifier": {verifier},
	}
	resp, err := http.PostForm(codexTokenURL, params)
	if err != nil {
		return nil, fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d", resp.StatusCode)
	}
	var tokens codexTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	return &tokens, nil
}

// --- Secret upsert helpers ---

// findSecretIDByName returns the ID of a secret with the given name, or "" if not found.
func findSecretIDByName(name string) (string, error) {
	var list SecretListResponse
	if err := doSecretsJSON(http.MethodGet, "", nil, &list); err != nil {
		return "", fmt.Errorf("failed to list secrets: %w", err)
	}
	for _, s := range list.Secrets {
		if s.Name == name {
			return s.ID, nil
		}
	}
	return "", nil
}

// upsertOAuthSecret creates or updates the named secret with the full OAuth bundle
// serialized as a JSON string, e.g. {"access_token":"...","refresh_token":"...","expires_at":"..."}.
func upsertOAuthSecret(name, accessToken, refreshToken, expiresAt string) error {
	bundle := codexBundle{
		Access:    accessToken,
		Refresh:   refreshToken,
		ExpiresAt: expiresAt,
	}
	bundleJSON, err := json.Marshal(bundle)
	if err != nil {
		return fmt.Errorf("failed to marshal OAuth bundle: %w", err)
	}
	value := string(bundleJSON)

	existingID, err := findSecretIDByName(name)
	if err != nil {
		return err
	}

	if existingID != "" {
		body := UpdateSecretBody{Value: value}
		return doSecretsJSON(http.MethodPut, "/"+existingID, body, nil)
	}

	body := CreateSecretBody{Name: name, Value: value}
	var resp CreateSecretResponse
	return doSecretsJSON(http.MethodPost, "", body, &resp)
}

// --- Local bundle (cache only — source of truth is the agents API) ---

type codexBundle struct {
	Access    string `json:"access_token"`
	Refresh   string `json:"refresh_token"`
	ExpiresAt string `json:"expires_at"`
}

func localBundlePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pinata-openai-oauth"), nil
}

func saveLocalBundle(b *codexBundle) {
	path, err := localBundlePath()
	if err != nil {
		return
	}
	data, _ := json.Marshal(b)
	_ = os.WriteFile(path, data, 0600)
}

// --- Public API ---

// CodexOAuthLogin runs the PKCE browser flow, stores the full OAuth bundle in
// the agents API (access token + refresh token + expiry), and caches it locally.
func CodexOAuthLogin() (*CreateSecretResponse, error) {
	verifier, err := generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}
	challenge := generateCodeChallenge(verifier)
	state, err := generateOAuthState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	authURL := buildCodexAuthURL(challenge, state)
	fmt.Println("Opening browser for OpenAI Codex authentication...")
	fmt.Printf("If the browser does not open automatically, visit:\n%s\n\n", authURL)
	_ = openBrowser(authURL)

	type callbackResult struct {
		code string
		err  error
	}
	ch := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux}

	mux.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if errParam := q.Get("error"); errParam != "" {
			http.Redirect(w, r, "/error?msg="+url.QueryEscape(errParam), http.StatusFound)
			ch <- callbackResult{err: fmt.Errorf("oauth error: %s", errParam)}
			return
		}
		if q.Get("state") != state {
			http.Redirect(w, r, "/error?msg=state+mismatch", http.StatusFound)
			ch <- callbackResult{err: fmt.Errorf("state mismatch: possible CSRF attack")}
			return
		}
		http.Redirect(w, r, "/success", http.StatusFound)
		ch <- callbackResult{code: q.Get("code")}
	})

	mux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html><body style="font-family:sans-serif;text-align:center;padding:60px">
<h2>Authentication Successful</h2>
<p>You can close this tab and return to the terminal.</p>
</body></html>`)
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		msg := r.URL.Query().Get("msg")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html><body style="font-family:sans-serif;text-align:center;padding:60px">
<h2>Authentication Failed</h2><p>%s</p>
<p>Please close this tab and try again.</p>
</body></html>`, msg)
	})

	ln, err := net.Listen("tcp", ":1455")
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server on port 1455 (is it already in use?): %w", err)
	}
	go func() { _ = srv.Serve(ln) }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var res callbackResult
	select {
	case res = <-ch:
	case <-ctx.Done():
		_ = srv.Shutdown(context.Background())
		return nil, fmt.Errorf("authentication timed out after 5 minutes")
	}

	time.Sleep(500 * time.Millisecond)
	_ = srv.Shutdown(context.Background())

	if res.err != nil {
		return nil, res.err
	}

	fmt.Println("Exchanging authorization code for tokens...")
	tokens, err := exchangeCodexToken(res.code, verifier)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second).UTC().Format(time.RFC3339)

	saveLocalBundle(&codexBundle{
		Access:    tokens.AccessToken,
		Refresh:   tokens.RefreshToken,
		ExpiresAt: expiresAt,
	})

	fmt.Printf("Storing secret '%s'...\n", codexSecretName)
	if err := upsertOAuthSecret(codexSecretName, tokens.AccessToken, tokens.RefreshToken, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to store secret: %w", err)
	}

	return nil, nil
}
