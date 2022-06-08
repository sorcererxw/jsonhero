package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pkg/browser"
)

func main() {
	var data string

	if stats, err := os.Stdin.Stat(); err == nil && stats.Size() > 0 {
		v, err := io.ReadAll(os.Stdin)

		if err != nil {
			fmt.Println(err)
			return
		}
		data = string(v)
	} else if len(os.Args) > 1 {
		data = os.Args[len(os.Args)-1]
	} else {
		fmt.Println("no input")
		return
	}

	var dest any
	if err := json.Unmarshal([]byte(data), &dest); err != nil {
		fmt.Println("invalid json")
		return
	}

	url, err := createJsonhero(context.Background(), dest)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := openBrowser(url); err != nil {
		fmt.Println(err)
		return
	}
}

func createJsonhero(ctx context.Context, content any) (string, error) {
	var payload io.Reader
	{
		data, err := json.Marshal(map[string]any{
			"title":   "Untitled",
			"content": content,
		})
		if err != nil {
			return "", fmt.Errorf("failed to marshal req: %w", err)
		}
		payload = bytes.NewReader(data)
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://jsonhero.io/api/create.json", payload)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Location string `json:"location"`
		Message  string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create jsonhero: %s", body.Message)
	}

	return body.Location, nil
}

func openBrowser(url string) error {
	return browser.OpenURL(url)
}
