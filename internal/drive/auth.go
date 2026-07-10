package drive

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func newOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID(),
		ClientSecret: clientSecret(),
		Scopes:       []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint:     google.Endpoint,
	}
}

// HTTPClient returns an authenticated client, reusing a cached token if one
// exists or running an interactive browser-based consent flow otherwise.
func HTTPClient(ctx context.Context) (*http.Client, error) {
	if clientID() == "" || clientSecret() == "" {
		return nil, fmt.Errorf("this build of falafal has no Google Drive credentials configured; drive support is unavailable")
	}

	path, err := tokenPath()
	if err != nil {
		return nil, err
	}

	tok, err := loadToken(path)
	if err != nil {
		tok, err = authenticate(ctx)
		if err != nil {
			return nil, fmt.Errorf("authenticating with Google: %w", err)
		}
		if err := saveToken(path, tok); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not cache auth token (%v); you'll need to sign in again next time\n", err)
		}
	}

	return newOAuthConfig().Client(ctx, tok), nil
}

func tokenPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "falafal", "token.json"), nil
}

func loadToken(path string) (*oauth2.Token, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tok oauth2.Token
	if err := json.Unmarshal(b, &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

func saveToken(path string, tok *oauth2.Token) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// authenticate runs the OAuth loopback flow: it starts a local HTTP server,
// opens the user's browser to Google's consent screen, and waits for the
// redirect carrying the authorization code.
func authenticate(ctx context.Context) (*oauth2.Token, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("starting local listener: %w", err)
	}
	defer listener.Close()

	cfg := *newOAuthConfig()
	cfg.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port)

	state, err := randomState()
	if err != nil {
		return nil, err
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("state"); got != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errCh <- fmt.Errorf("oauth state mismatch")
			return
		}
		if msg := r.URL.Query().Get("error"); msg != "" {
			fmt.Fprintln(w, "Sign-in failed. You can close this tab and check the terminal.")
			errCh <- fmt.Errorf("google returned an error: %s", msg)
			return
		}
		fmt.Fprintln(w, "Signed in. You can close this tab and go back to falafal.")
		codeCh <- r.URL.Query().Get("code")
	})
	srv := &http.Server{Handler: mux}
	go srv.Serve(listener)
	defer srv.Close()

	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
	fmt.Println("Opening your browser to sign in to Google Drive...")
	fmt.Println("If it doesn't open automatically, visit this URL:")
	fmt.Println(authURL)
	_ = openBrowser(authURL)

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging authorization code: %w", err)
	}
	return tok, nil
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
