package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const defaultHTTPTimeout = 15 * time.Second

type ArcscanClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type ArcscanClientConfig struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Timeout    time.Duration
}

func NewArcscanClient(cfg ArcscanClientConfig) *ArcscanClient {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = DefaultArcscanBaseURL
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = defaultHTTPTimeout
		}
		httpClient = &http.Client{Timeout: timeout}
	}

	return &ArcscanClient{
		baseURL:    baseURL,
		apiKey:     strings.TrimSpace(cfg.APIKey),
		httpClient: httpClient,
	}
}

func (client *ArcscanClient) FetchAddressLogs(ctx context.Context, address string, pageParams map[string]string) (LogsPage, error) {
	if client == nil {
		return LogsPage{}, fmt.Errorf("arcscan client is nil")
	}
	address = strings.TrimSpace(address)
	if address == "" {
		return LogsPage{}, fmt.Errorf("address is required")
	}

	endpoint, err := url.Parse(client.baseURL + "/api/v2/addresses/" + url.PathEscape(address) + "/logs")
	if err != nil {
		return LogsPage{}, fmt.Errorf("build logs URL: %w", err)
	}
	query := endpoint.Query()
	for _, key := range sortedKeys(pageParams) {
		query.Set(key, pageParams[key])
	}
	if client.apiKey != "" {
		query.Set("apikey", client.apiKey)
	}
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return LogsPage{}, fmt.Errorf("build logs request: %w", err)
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return LogsPage{}, fmt.Errorf("fetch arcscan logs: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return LogsPage{}, fmt.Errorf("fetch arcscan logs: unexpected status %d", response.StatusCode)
	}

	var body blockscoutLogsResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return LogsPage{}, fmt.Errorf("decode arcscan logs response: %w", err)
	}

	items := make([]BlockscoutLog, 0, len(body.Items))
	for _, raw := range body.Items {
		var item BlockscoutLog
		if err := json.Unmarshal(raw, &item); err != nil {
			return LogsPage{}, fmt.Errorf("decode arcscan log item: %w", err)
		}
		item.Raw = append(json.RawMessage(nil), raw...)
		items = append(items, item)
	}

	return LogsPage{
		Items:          items,
		NextPageParams: stringMap(body.NextPageParams),
	}, nil
}

type blockscoutLogsResponse struct {
	Items          []json.RawMessage `json:"items"`
	NextPageParams map[string]any    `json:"next_page_params"`
}

func stringMap(input map[string]any) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		switch typed := value.(type) {
		case string:
			output[key] = typed
		case float64:
			output[key] = fmt.Sprintf("%.0f", typed)
		case bool:
			if typed {
				output[key] = "true"
			} else {
				output[key] = "false"
			}
		default:
			output[key] = fmt.Sprintf("%v", typed)
		}
	}
	return output
}

func sortedKeys(input map[string]string) []string {
	keys := make([]string, 0, len(input))
	for key, value := range input {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
