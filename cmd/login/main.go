package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	userAgent = "Appie/9.28 (iPhone17,3; iPhone; CPU OS 26_1 like Mac OS X)"
	clientID  = "appie-ios"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "refresh" {
		if len(os.Args) < 3 {
			fmt.Println("Usage: login refresh <refresh_token>")
			os.Exit(1)
		}
		refreshToken(os.Args[2])
		return
	}

	loginFlow()
}

func loginFlow() {
	redirectURI := "appie://login-exit"
	loginURL := fmt.Sprintf(
		"https://login.ah.nl/login?client_id=%s&response_type=code&redirect_uri=%s",
		clientID, redirectURI,
	)

	fmt.Println("=== Albert Heijn Login ===")
	fmt.Println()
	fmt.Println("1. Open this URL in your browser:")
	fmt.Println()
	fmt.Printf("   %s\n", loginURL)
	fmt.Println()
	fmt.Println("2. Login with your credentials")
	fmt.Println("3. After login, browser will try to open 'appie://login-exit?code=...'")
	fmt.Println("4. Copy the 'code' value from the URL (or the full URL)")
	fmt.Println()
	fmt.Print("Paste code here: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	code := extractCode(strings.TrimSpace(input))
	if code == "" {
		fmt.Fprintf(os.Stderr, "Could not extract code from input\n")
		os.Exit(1)
	}

	fmt.Printf("Got code: %s\n", code)
	fmt.Println()
	fmt.Println("Exchanging code for tokens...")

	exchangeCode(code)
}

func extractCode(input string) string {
	// If it's just the code (no URL structure)
	if !strings.Contains(input, "=") && !strings.Contains(input, "?") {
		return input
	}

	// Parse appie://login-exit?code=XXX or just code=XXX
	if idx := strings.Index(input, "code="); idx != -1 {
		code := input[idx+5:]
		// Remove any trailing parameters
		if ampIdx := strings.Index(code, "&"); ampIdx != -1 {
			code = code[:ampIdx]
		}
		return code
	}

	return ""
}

func exchangeCode(code string) {
	body := map[string]string{
		"clientId": clientID,
		"code":     code,
	}
	bodyJSON, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, "https://api.ah.nl/mobile-auth/v1/auth/token", bytes.NewReader(bodyJSON))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create request: %v\n", err)
		os.Exit(1)
	}
	setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Token exchange failed (%d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		MemberID     string `json:"member_id"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== LOGIN SUCCESSFUL ===")
	fmt.Println()
	fmt.Printf("Access Token:  %s\n", tokenResp.AccessToken)
	fmt.Printf("Refresh Token: %s\n", tokenResp.RefreshToken)
	fmt.Printf("Member ID:     %s\n", tokenResp.MemberID)
	fmt.Printf("Expires In:    %d seconds (~%d days)\n", tokenResp.ExpiresIn, tokenResp.ExpiresIn/86400)
	fmt.Println()
	fmt.Println("Store the refresh token for future sessions!")
}

func refreshToken(token string) {
	body := map[string]string{
		"clientId":     clientID,
		"refreshToken": token,
	}
	bodyJSON, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, "https://api.ah.nl/mobile-auth/v1/auth/token/refresh", bytes.NewReader(bodyJSON))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create request: %v\n", err)
		os.Exit(1)
	}
	setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Token refresh failed (%d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== TOKEN REFRESH SUCCESSFUL ===")
	fmt.Println()
	fmt.Printf("Access Token:  %s\n", tokenResp.AccessToken)
	fmt.Printf("Refresh Token: %s\n", tokenResp.RefreshToken)
	fmt.Printf("Expires In:    %d seconds (~%d days)\n", tokenResp.ExpiresIn, tokenResp.ExpiresIn/86400)
}

func setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("x-client-name", clientID)
	req.Header.Set("x-client-version", "9.28")
	req.Header.Set("x-application", "AHWEBSHOP")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
}
