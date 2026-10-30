// Extract structured data from bank and card statements using the Photon Commerce API.
//
// Submits a statement (PDF or image) and returns account details, balances,
// and a full transaction ledger including client name, bank name, account number,
// starting/ending balances, and all transactions (date, type, description, amount).
//
// Processing times (Managed Agents):
//   Trial accounts:  up to 24 hours
//   Production:      5 minutes to 24 hours
//
// AI extraction (seconds, no Managed Agents):
//   Contact support@photoncommerce.com to activate.
//   Once active, submit to /api/v4 instead of /api/pro.
//
// Docs:    https://apidocs.photoncommerce.com
// Sandbox: https://sandbox-api.photoncommerce.com/api/v4/register (20 free calls)

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	clientID  = "YOUR_CLIENT_ID"
	username  = "YOUR_USERNAME"
	apiKey    = "YOUR_API_KEY"
	password  = "YOUR_PASSWORD"
	secretKey = "YOUR_SECRET_KEY"

	// Sandbox: https://sandbox-api.photoncommerce.com  (20 free calls, no card needed)
	// Production: https://api.photoncommerce.com
	baseURL = "https://sandbox-api.photoncommerce.com"
)

func addAuthHeaders(req *http.Request) {
	req.Header.Set("CLIENT-ID", clientID)
	req.Header.Set("AUTHORIZATION", fmt.Sprintf("apikey %s:%s", username, apiKey))
	req.Header.Set("PASSWORD", password)
	req.Header.Set("SECRET-KEY", secretKey)
}

type SubmitOptions struct {
	FilePath   string
	URL        string
	WebhookURL string
	AuthToken  string
	ID         string
	Subaccount string
	PageStart  int
	PageEnd    int
}

func submitStatement(opts SubmitOptions) (string, error) {
	if opts.FilePath == "" && opts.URL == "" {
		return "", fmt.Errorf("provide either FilePath or URL")
	}

	params := url.Values{"doctype": {"statement"}}
	if opts.URL != ""        { params.Set("url", opts.URL) }
	if opts.WebhookURL != "" { params.Set("webhook_url", opts.WebhookURL) }
	if opts.AuthToken != ""  { params.Set("auth_token", opts.AuthToken) }
	if opts.ID != ""         { params.Set("ID", opts.ID) }
	if opts.Subaccount != "" { params.Set("subaccount", opts.Subaccount) }
	if opts.PageStart > 0    { params.Set("page_start", fmt.Sprint(opts.PageStart)) }
	if opts.PageEnd > 0      { params.Set("page_end", fmt.Sprint(opts.PageEnd)) }

	// For AI extraction (seconds), replace /api/pro with /api/v4 — contact support@photoncommerce.com to activate.
	endpoint := fmt.Sprintf("%s/api/pro?%s", baseURL, params.Encode())

	var req *http.Request
	if opts.FilePath != "" {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		file, err := os.Open(opts.FilePath)
		if err != nil { return "", err }
		defer file.Close()
		part, err := writer.CreateFormFile("pdf", filepath.Base(opts.FilePath))
		if err != nil { return "", err }
		if _, err = io.Copy(part, file); err != nil { return "", err }
		writer.Close()
		req, err = http.NewRequest("POST", endpoint, &buf)
		if err != nil { return "", err }
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else {
		var err error
		req, err = http.NewRequest("POST", endpoint, nil)
		if err != nil { return "", err }
	}

	addAuthHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return "", err }
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	photonKey, _ := result["photon_key"].(string)
	return photonKey, nil
}

func fetchResult(photonKey string) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v4/json?photon_key=%s", baseURL, photonKey), nil)
	addAuthHeaders(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	data, _ := body["data"].(map[string]interface{})
	return data, nil
}

func waitForResult(photonKey string, pollInterval, timeout time.Duration) (map[string]interface{}, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		result, err := fetchResult(photonKey)
		if err != nil { return nil, err }
		status, _ := result["Status"].(string)
		if status != "" && status != "pending" && status != "processing" { return result, nil }
		if status == "" { status = "pending" }
		fmt.Printf("  Status: %s — retrying in %s...\n", status, pollInterval)
		time.Sleep(pollInterval)
	}
	return nil, fmt.Errorf("extraction not complete after %s", timeout)
}

func str(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func main() {
	// --- Option A: submit from a local file ---
	photonKey, err := submitStatement(SubmitOptions{FilePath: "statement.pdf"})

	// --- Option B: submit via a publicly accessible URL ---
	// photonKey, err := submitStatement(SubmitOptions{URL: "https://example.com/statement.pdf"})

	if err != nil { panic(err) }
	fmt.Println("Submitted. photon_key:", photonKey)
	fmt.Println("Waiting for extraction to complete...")

	result, err := waitForResult(photonKey, 20*time.Second, time.Hour)
	if err != nil { panic(err) }

	fmt.Println("\n--- Bank Statement Data ---")
	fmt.Println("Client:          ", str(result, "client_name"))
	fmt.Println("Bank:            ", str(result, "bank_name"))
	fmt.Println("Account Number:  ", str(result, "account_number"))
	fmt.Println("Account Type:    ", str(result, "account_type"))
	fmt.Println("Period:          ", str(result, "statement_start_date"), "→", str(result, "statement_end_date"))
	fmt.Println("Starting Balance:", result["starting_balance"])
	fmt.Println("Ending Balance:  ", result["ending_balance"])
	fmt.Println("Total Credits:   ", result["tot_credit"])
	fmt.Println("Total Debits:    ", result["tot_debit"])

	if txns, ok := result["transactions"].([]interface{}); ok && len(txns) > 0 {
		fmt.Printf("\n--- Transactions (%d) ---\n", len(txns))
		for _, raw := range txns {
			t, _ := raw.(map[string]interface{})
			fmt.Printf("  %-12s %-6s %10v  %s\n",
				str(t, "date"), str(t, "type"), t["amount"], str(t, "description"))
		}
	}
}
