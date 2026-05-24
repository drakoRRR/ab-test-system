package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	token := os.Getenv("FIREBASE_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "FIREBASE_TOKEN is not set.")
		fmt.Fprintln(os.Stderr, "Log in to the admin UI, then:")
		fmt.Fprintln(os.Stderr, "  DevTools → Application → Cookies → copy 'firebase-token'")
		fmt.Fprintln(os.Stderr, "  export FIREBASE_TOKEN=<value>")
		os.Exit(1)
	}

	apiURL := envOrDefault("API_URL", "http://localhost:8080/api/v1")
	demoDir := envOrDefault("DEMO_DIR", "demo")
	c := newClient(apiURL, token)

	fmt.Println("→ Resolving organization...")
	orgID, err := c.resolveOrg()
	must("resolve org", err)
	fmt.Printf("  org: %s\n", orgID)

	fmt.Println("→ Resolving project...")
	projectID, err := c.resolveProject("Demo Store")
	must("resolve project", err)
	fmt.Printf("  project: %s\n", projectID)

	fmt.Println("→ Resolving experiment...")
	expID, err := c.resolveExperiment(projectID)
	must("resolve experiment", err)
	fmt.Printf("  experiment: %s\n", expID)

	fmt.Println("→ Starting experiment (if not already running)...")
	if err := c.startExperiment(projectID, expID); err != nil {
		fmt.Printf("  skipped (%v)\n", err)
	}

	fmt.Println("→ Creating SDK key...")
	sdkKey, err := c.createSDKKey(projectID)
	must("create sdk key", err)

	envPath := demoDir + "/.env"
	content := fmt.Sprintf("SDK_API_KEY=%s\nSDK_SERVICE_URL=%s\n", sdkKey, apiURL)
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		must("write .env", err)
	}

	fmt.Printf("\n✓  %s written\n", envPath)
	fmt.Printf("   SDK_API_KEY=%s\n", sdkKey)
	fmt.Println("\nNext: make demo-run")
}

type apiClient struct {
	base string
	tok  string
	http *http.Client
}

func newClient(base, token string) *apiClient {
	return &apiClient{base: base, tok: token, http: &http.Client{Timeout: 10 * time.Second}}
}

func (c *apiClient) do(method, path string, body any) (*http.Response, error) {
	var b []byte
	if body != nil {
		var err error
		b, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, c.base+path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.tok)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.http.Do(req)
}

func decode(resp *http.Response, dst any) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(dst)
}

func (c *apiClient) resolveOrg() (string, error) {
	// GET /users/me avoids creating a duplicate org if one already exists.
	resp, err := c.do("GET", "/users/me", nil)
	if err != nil {
		return "", err
	}
	var user struct {
		OrganizationID string `json:"organizationId"`
	}
	if err := decode(resp, &user); err != nil {
		return "", err
	}
	if user.OrganizationID != "" {
		return user.OrganizationID, nil
	}

	resp, err = c.do("POST", "/organizations", map[string]string{"name": "Demo Org"})
	if err != nil {
		return "", err
	}
	var org struct{ ID string `json:"id"` }
	if err := decode(resp, &org); err != nil {
		return "", err
	}
	return org.ID, nil
}

func (c *apiClient) resolveProject(name string) (string, error) {
	resp, err := c.do("POST", "/projects", map[string]string{
		"name":        name,
		"description": "A/B testing diploma demo project",
	})
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusCreated {
		var proj struct{ ID string `json:"id"` }
		if err := decode(resp, &proj); err != nil {
			return "", err
		}
		return proj.ID, nil
	}
	if resp.StatusCode != http.StatusConflict {
		resp.Body.Close()
		return "", fmt.Errorf("create project: unexpected status %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 409: find by name in the list.
	resp, err = c.do("GET", "/projects", nil)
	if err != nil {
		return "", err
	}
	var projects []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := decode(resp, &projects); err != nil {
		return "", err
	}
	for _, p := range projects {
		if p.Name == name {
			return p.ID, nil
		}
	}
	return "", errors.New("project not found after 409")
}

func (c *apiClient) resolveExperiment(projectID string) (string, error) {
	body := map[string]any{
		"key":            "checkout-btn",
		"name":           "Checkout Button Experiment",
		"description":    "Control: blue Buy Now. Treatment: green Get It Now.",
		"trafficPercent": 100,
		"variants": []map[string]any{
			{"key": "control", "name": "Control", "weight": 50},
			{"key": "treatment", "name": "Treatment", "weight": 50},
		},
	}
	resp, err := c.do("POST", "/projects/"+projectID+"/experiments", body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusCreated {
		var exp struct{ ID string `json:"id"` }
		if err := decode(resp, &exp); err != nil {
			return "", err
		}
		return exp.ID, nil
	}
	if resp.StatusCode != http.StatusConflict {
		resp.Body.Close()
		return "", fmt.Errorf("create experiment: unexpected status %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 409: find by key in the list.
	resp, err = c.do("GET", "/projects/"+projectID+"/experiments", nil)
	if err != nil {
		return "", err
	}
	var experiments []struct {
		ID  string `json:"id"`
		Key string `json:"key"`
	}
	if err := decode(resp, &experiments); err != nil {
		return "", err
	}
	for _, e := range experiments {
		if e.Key == "checkout-btn" {
			return e.ID, nil
		}
	}
	return "", errors.New("experiment not found after 409")
}

func (c *apiClient) startExperiment(projectID, expID string) error {
	// Non-2xx is expected when already running — callers treat this as skippable.
	resp, err := c.do("POST", "/projects/"+projectID+"/experiments/"+expID+"/start", nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func (c *apiClient) createSDKKey(projectID string) (string, error) {
	resp, err := c.do("POST", "/projects/"+projectID+"/api-keys", map[string]string{"name": "demo"})
	if err != nil {
		return "", err
	}
	var result struct{ Key string `json:"key"` }
	if err := decode(resp, &result); err != nil {
		return "", err
	}
	if result.Key == "" {
		return "", errors.New("empty key in response")
	}
	return result.Key, nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func must(op string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗  %s: %v\n", op, err)
		os.Exit(1)
	}
}
