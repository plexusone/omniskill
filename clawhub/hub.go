package clawhub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultHubURL is the default ClawHub API endpoint.
	DefaultHubURL = "https://api.clawhub.ai/v1"
)

// Hub provides access to the ClawHub skills marketplace API.
type Hub struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// Option is a functional option for configuring the Hub.
type Option func(*Hub)

// WithBaseURL sets a custom base URL for the Hub API.
func WithBaseURL(url string) Option {
	return func(h *Hub) {
		h.baseURL = url
	}
}

// WithAPIKey sets the API key for authenticated requests.
func WithAPIKey(apiKey string) Option {
	return func(h *Hub) {
		h.apiKey = apiKey
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(h *Hub) {
		h.httpClient = client
	}
}

// New creates a new Hub client.
func New(opts ...Option) *Hub {
	h := &Hub{
		baseURL: DefaultHubURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// SkillInfo contains information about a skill from the Hub.
type SkillInfo struct {
	// Name is the skill name.
	Name string `json:"name"`

	// Version is the latest version.
	Version string `json:"version"`

	// Description is the skill description.
	Description string `json:"description"`

	// Author is the skill author.
	Author string `json:"author"`

	// Repository is the source repository URL.
	Repository string `json:"repository"`

	// Downloads is the total download count.
	Downloads int `json:"downloads"`

	// Stars is the star count.
	Stars int `json:"stars"`

	// License is the SPDX license identifier.
	License string `json:"license"`

	// Keywords are searchable tags.
	Keywords []string `json:"keywords"`

	// CreatedAt is when the skill was first published.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is when the skill was last updated.
	UpdatedAt time.Time `json:"updatedAt"`

	// SecurityState indicates the security verification state.
	SecurityState SecurityState `json:"securityState"`

	// Versions lists all available versions.
	Versions []string `json:"versions,omitempty"`
}

// SecurityState represents the security verification state of a skill.
type SecurityState string

const (
	// SecurityStateUnknown indicates the skill has not been verified.
	SecurityStateUnknown SecurityState = "unknown"

	// SecurityStatePending indicates verification is in progress.
	SecurityStatePending SecurityState = "pending"

	// SecurityStateVerified indicates the skill passed security checks.
	SecurityStateVerified SecurityState = "verified"

	// SecurityStateFlagged indicates security issues were found.
	SecurityStateFlagged SecurityState = "flagged"
)

// SearchResult contains the result of a skill search.
type SearchResult struct {
	// Skills is the list of matching skills.
	Skills []SkillInfo `json:"skills"`

	// Total is the total number of matching skills.
	Total int `json:"total"`

	// Page is the current page number.
	Page int `json:"page"`

	// PerPage is the number of results per page.
	PerPage int `json:"perPage"`
}

// Search searches for skills matching the query.
func (h *Hub) Search(ctx context.Context, query string, page, perPage int) (*SearchResult, error) {
	if perPage <= 0 {
		perPage = 20
	}
	if page <= 0 {
		page = 1
	}

	params := url.Values{
		"q":       {query},
		"page":    {fmt.Sprintf("%d", page)},
		"perPage": {fmt.Sprintf("%d", perPage)},
	}

	var result SearchResult
	if err := h.get(ctx, "/skills/search?"+params.Encode(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Get retrieves information about a specific skill.
func (h *Hub) Get(ctx context.Context, name string) (*SkillInfo, error) {
	var info SkillInfo
	if err := h.get(ctx, "/skills/"+url.PathEscape(name), &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// GetVersion retrieves information about a specific version of a skill.
func (h *Hub) GetVersion(ctx context.Context, name, version string) (*SkillInfo, error) {
	var info SkillInfo
	path := fmt.Sprintf("/skills/%s/versions/%s", url.PathEscape(name), url.PathEscape(version))
	if err := h.get(ctx, path, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// ListVersions lists all versions of a skill.
func (h *Hub) ListVersions(ctx context.Context, name string) ([]string, error) {
	var versions []string
	if err := h.get(ctx, "/skills/"+url.PathEscape(name)+"/versions", &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

// Resolve resolves a skill reference to a SkillInfo with repository details.
func (h *Hub) Resolve(ctx context.Context, ref string) (*SkillInfo, error) {
	// If ref starts with @clawhub/, strip the prefix
	if len(ref) > 9 && ref[:9] == "@clawhub/" {
		ref = ref[9:]
	}

	return h.Get(ctx, ref)
}

// get performs a GET request to the Hub API.
func (h *Hub) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("skill not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}
