package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type OAuthDebugger struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	BaseURL      string
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}

	// Initialize debugger
	debugger := &OAuthDebugger{
		ClientID:     os.Getenv("GOHIGHLEVEL_CLIENT_ID"),
		ClientSecret: os.Getenv("GOHIGHLEVEL_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("GOHIGHLEVEL_REDIRECT_URI"),
		BaseURL:      os.Getenv("GOHIGHLEVEL_BASE_URL"),
	}

	fmt.Println("=== GoHighLevel OAuth Flow Debugger ===")
	fmt.Printf("Client ID: %s\n", debugger.ClientID)
	fmt.Printf("Client Secret: %s...\n", debugger.ClientSecret[:10])
	fmt.Printf("Redirect URI: %s\n", debugger.RedirectURI)
	fmt.Printf("Base URL: %s\n", debugger.BaseURL)
	fmt.Println()

	// Test 1: Generate OAuth URL
	fmt.Println("=== Test 1: Generate OAuth URL ===")
	oauthURL := debugger.generateOAuthURL("test-company-123")
	fmt.Printf("Generated OAuth URL: %s\n", oauthURL)
	fmt.Println()

	// Test 2: Validate URL structure
	fmt.Println("=== Test 2: Validate URL Structure ===")
	debugger.validateURLStructure(oauthURL)
	fmt.Println()

	// Test 3: Test OAuth endpoint accessibility
	fmt.Println("=== Test 3: Test OAuth Endpoint Accessibility ===")
	debugger.testOAuthEndpoint()
	fmt.Println()

	// Test 4: Test callback endpoint
	fmt.Println("=== Test 4: Test Callback Endpoint ===")
	debugger.testCallbackEndpoint()
	fmt.Println()

	// Test 5: Simulate callback with missing parameters
	fmt.Println("=== Test 5: Simulate Missing Parameters ===")
	debugger.simulateMissingParameters()
	fmt.Println()

	fmt.Println("=== Debug Complete ===")
	fmt.Println("\nNext Steps:")
	fmt.Println("1. Copy the OAuth URL above and test it in a browser")
	fmt.Println("2. Check if GoHighLevel redirects properly to your callback")
	fmt.Println("3. Verify Railway environment variables match local .env")
	fmt.Println("4. Check Railway logs for detailed error messages")
}

func (d *OAuthDebugger) generateOAuthURL(companyID string) string {
	state := fmt.Sprintf("debug-state-%d", time.Now().Unix())
	
	params := url.Values{
		"response_type": {"code"},
		"client_id":     {d.ClientID},
		"redirect_uri":  {d.RedirectURI},
		"scope":         {"locations.readonly contacts.readonly"},
		"state":         {state},
	}

	return fmt.Sprintf("%s?%s", d.BaseURL, params.Encode())
}

func (d *OAuthDebugger) validateURLStructure(oauthURL string) {
	parsedURL, err := url.Parse(oauthURL)
	if err != nil {
		fmt.Printf("❌ Invalid URL structure: %v\n", err)
		return
	}

	fmt.Printf("✅ URL Structure Valid\n")
	fmt.Printf("   Scheme: %s\n", parsedURL.Scheme)
	fmt.Printf("   Host: %s\n", parsedURL.Host)
	fmt.Printf("   Path: %s\n", parsedURL.Path)

	queryParams := parsedURL.Query()
	requiredParams := []string{"response_type", "client_id", "redirect_uri", "scope", "state"}
	
	for _, param := range requiredParams {
		if value := queryParams.Get(param); value != "" {
			fmt.Printf("   ✅ %s: %s\n", param, value)
		} else {
			fmt.Printf("   ❌ %s: MISSING\n", param)
		}
	}
}

func (d *OAuthDebugger) testOAuthEndpoint() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(d.BaseURL)
	if err != nil {
		fmt.Printf("❌ Cannot reach OAuth endpoint: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ OAuth endpoint accessible\n")
	fmt.Printf("   Status: %s\n", resp.Status)
	fmt.Printf("   Content-Type: %s\n", resp.Header.Get("Content-Type"))
}

func (d *OAuthDebugger) testCallbackEndpoint() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Test callback endpoint without parameters (should return error)
	resp, err := client.Get(d.RedirectURI)
	if err != nil {
		fmt.Printf("❌ Cannot reach callback endpoint: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ Callback endpoint accessible\n")
	fmt.Printf("   Status: %s\n", resp.Status)
	fmt.Printf("   Response: %s\n", string(body))
}

func (d *OAuthDebugger) simulateMissingParameters() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	testCases := []struct {
		name   string
		params string
	}{
		{"No parameters", ""},
		{"Only code", "?code=test123"},
		{"Only state", "?state=test456"},
		{"Code + State", "?code=test123&state=test456"},
		{"With error", "?error=access_denied"},
	}

	for _, tc := range testCases {
		fmt.Printf("Testing: %s\n", tc.name)
		testURL := d.RedirectURI + tc.params
		
		resp, err := client.Get(testURL)
		if err != nil {
			fmt.Printf("   ❌ Request failed: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   Status: %s\n", resp.Status)
		
		// Try to parse as JSON
		var jsonResp map[string]interface{}
		if err := json.Unmarshal(body, &jsonResp); err == nil {
			if errorMsg, exists := jsonResp["error"]; exists {
				fmt.Printf("   Error: %v\n", errorMsg)
			}
		} else {
			fmt.Printf("   Response: %s\n", strings.TrimSpace(string(body)))
		}
		fmt.Println()
	}
}