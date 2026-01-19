package libecto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Client is a Ghost Admin API client.
// It handles authentication and provides methods for all Ghost Admin API endpoints.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// ClientOption is a function that configures a Client.
// Use with NewClient to customize client behavior.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client for the Ghost API client.
// This is useful for testing or for configuring custom timeouts, transports, etc.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// NewClient creates a new Ghost Admin API client.
// The url parameter is the Ghost site URL (e.g., "https://mysite.ghost.io").
// The apiKey parameter is the Admin API key in "id:secret" format.
// Optional ClientOption functions can be passed to customize the client.
func NewClient(url, apiKey string, opts ...ClientOption) *Client {
	url = strings.TrimSuffix(url, "/")
	c := &Client{
		baseURL:    url + "/ghost/api/admin",
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// BaseURL returns the base URL for API requests.
// This is useful for testing or debugging.
func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) request(method, path string, body interface{}) (*http.Response, error) {
	token, err := GenerateToken(c.apiKey)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Ghost "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

func (c *Client) do(method, path string, body, result interface{}) error {
	resp, err := c.request(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && len(errResp.Errors) > 0 {
			msg := errResp.Errors[0].Message
			if errResp.Errors[0].Context != "" {
				msg += ": " + errResp.Errors[0].Context
			}
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
		}
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		return json.Unmarshal(respBody, result)
	}
	return nil
}

// Posts

// ListPosts returns a list of posts from the Ghost site.
// The status parameter can be "draft", "published", "scheduled", or "all" (empty string also returns all).
// The limit parameter controls the number of results (0 for default).
func (c *Client) ListPosts(status string, limit int) (*PostsResponse, error) {
	path := "/posts/?formats=html"
	if status != "" && status != "all" {
		path += "&filter=status:" + status
	}
	if limit > 0 {
		path += fmt.Sprintf("&limit=%d", limit)
	}

	var resp PostsResponse
	if err := c.do("GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPost returns a single post by ID or slug.
// It first tries to find by ID, then falls back to slug lookup.
func (c *Client) GetPost(idOrSlug string) (*Post, error) {
	var resp PostsResponse
	err := c.do("GET", "/posts/"+idOrSlug+"/?formats=html", nil, &resp)
	if err != nil {
		// Try by slug
		err = c.do("GET", "/posts/slug/"+idOrSlug+"/?formats=html", nil, &resp)
		if err != nil {
			return nil, err
		}
	}
	if len(resp.Posts) == 0 {
		return nil, fmt.Errorf("post not found: %s", idOrSlug)
	}
	return &resp.Posts[0], nil
}

// CreatePost creates a new post with the given data.
// At minimum, the post should have a Title set.
func (c *Client) CreatePost(post *Post) (*Post, error) {
	body := map[string][]Post{"posts": {*post}}
	var resp PostsResponse
	if err := c.do("POST", "/posts/?source=html&formats=html", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Posts) == 0 {
		return nil, fmt.Errorf("no post returned")
	}
	return &resp.Posts[0], nil
}

// UpdatePost updates an existing post by ID.
// The post.UpdatedAt field should be set to the current updated_at value for conflict detection.
func (c *Client) UpdatePost(id string, post *Post) (*Post, error) {
	body := map[string][]Post{"posts": {*post}}
	var resp PostsResponse
	if err := c.do("PUT", "/posts/"+id+"/?source=html&formats=html", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Posts) == 0 {
		return nil, fmt.Errorf("no post returned")
	}
	return &resp.Posts[0], nil
}

// DeletePost permanently deletes a post by ID.
func (c *Client) DeletePost(id string) error {
	return c.do("DELETE", "/posts/"+id+"/", nil, nil)
}

// PublishPost publishes a draft post by ID or slug.
// It first retrieves the post to get the current updated_at for conflict detection.
func (c *Client) PublishPost(idOrSlug string) (*Post, error) {
	existing, err := c.GetPost(idOrSlug)
	if err != nil {
		return nil, err
	}
	return c.UpdatePost(existing.ID, &Post{
		UpdatedAt: existing.UpdatedAt,
		Status:    "published",
	})
}

// UnpublishPost unpublishes a post (sets to draft) by ID or slug.
// It first retrieves the post to get the current updated_at for conflict detection.
func (c *Client) UnpublishPost(idOrSlug string) (*Post, error) {
	existing, err := c.GetPost(idOrSlug)
	if err != nil {
		return nil, err
	}
	return c.UpdatePost(existing.ID, &Post{
		UpdatedAt: existing.UpdatedAt,
		Status:    "draft",
	})
}

// SchedulePost schedules a post for publication at a specific time.
// The publishAt parameter should be an ISO8601 timestamp (e.g., "2025-01-15T12:00:00Z").
func (c *Client) SchedulePost(idOrSlug, publishAt string) (*Post, error) {
	existing, err := c.GetPost(idOrSlug)
	if err != nil {
		return nil, err
	}
	return c.UpdatePost(existing.ID, &Post{
		UpdatedAt:   existing.UpdatedAt,
		Status:      "scheduled",
		PublishedAt: publishAt,
	})
}

// Pages

// ListPages returns a list of pages from the Ghost site.
// The status parameter can be "draft", "published", or "all".
// The limit parameter controls the number of results (0 for default).
func (c *Client) ListPages(status string, limit int) (*PagesResponse, error) {
	path := "/pages/?formats=html"
	if status != "" && status != "all" {
		path += "&filter=status:" + status
	}
	if limit > 0 {
		path += fmt.Sprintf("&limit=%d", limit)
	}

	var resp PagesResponse
	if err := c.do("GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPage returns a single page by ID or slug.
// It first tries to find by ID, then falls back to slug lookup.
func (c *Client) GetPage(idOrSlug string) (*Page, error) {
	var resp PagesResponse
	err := c.do("GET", "/pages/"+idOrSlug+"/?formats=html", nil, &resp)
	if err != nil {
		err = c.do("GET", "/pages/slug/"+idOrSlug+"/?formats=html", nil, &resp)
		if err != nil {
			return nil, err
		}
	}
	if len(resp.Pages) == 0 {
		return nil, fmt.Errorf("page not found: %s", idOrSlug)
	}
	return &resp.Pages[0], nil
}

// CreatePage creates a new page with the given data.
// At minimum, the page should have a Title set.
func (c *Client) CreatePage(page *Page) (*Page, error) {
	body := map[string][]Page{"pages": {*page}}
	var resp PagesResponse
	if err := c.do("POST", "/pages/?formats=html", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Pages) == 0 {
		return nil, fmt.Errorf("no page returned")
	}
	return &resp.Pages[0], nil
}

// UpdatePage updates an existing page by ID.
// The page.UpdatedAt field should be set to the current updated_at value for conflict detection.
func (c *Client) UpdatePage(id string, page *Page) (*Page, error) {
	body := map[string][]Page{"pages": {*page}}
	var resp PagesResponse
	if err := c.do("PUT", "/pages/"+id+"/?formats=html", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Pages) == 0 {
		return nil, fmt.Errorf("no page returned")
	}
	return &resp.Pages[0], nil
}

// DeletePage permanently deletes a page by ID.
func (c *Client) DeletePage(id string) error {
	return c.do("DELETE", "/pages/"+id+"/", nil, nil)
}

// PublishPage publishes a draft page by ID or slug.
// It first retrieves the page to get the current updated_at for conflict detection.
func (c *Client) PublishPage(idOrSlug string) (*Page, error) {
	existing, err := c.GetPage(idOrSlug)
	if err != nil {
		return nil, err
	}
	return c.UpdatePage(existing.ID, &Page{
		UpdatedAt: existing.UpdatedAt,
		Status:    "published",
	})
}

// Tags

// ListTags returns a list of tags from the Ghost site.
// The limit parameter controls the number of results (0 for default).
// Results include post counts for each tag.
func (c *Client) ListTags(limit int) (*TagsResponse, error) {
	path := "/tags/?include=count.posts"
	if limit > 0 {
		path += fmt.Sprintf("&limit=%d", limit)
	}

	var resp TagsResponse
	if err := c.do("GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTag returns a single tag by ID or slug.
// It first tries to find by ID, then falls back to slug lookup.
func (c *Client) GetTag(idOrSlug string) (*Tag, error) {
	var resp TagsResponse
	err := c.do("GET", "/tags/"+idOrSlug+"/", nil, &resp)
	if err != nil {
		err = c.do("GET", "/tags/slug/"+idOrSlug+"/", nil, &resp)
		if err != nil {
			return nil, err
		}
	}
	if len(resp.Tags) == 0 {
		return nil, fmt.Errorf("tag not found: %s", idOrSlug)
	}
	return &resp.Tags[0], nil
}

// CreateTag creates a new tag with the given data.
// At minimum, the tag should have a Name set.
func (c *Client) CreateTag(tag *Tag) (*Tag, error) {
	body := map[string][]Tag{"tags": {*tag}}
	var resp TagsResponse
	if err := c.do("POST", "/tags/", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Tags) == 0 {
		return nil, fmt.Errorf("no tag returned")
	}
	return &resp.Tags[0], nil
}

// UpdateTag updates an existing tag by ID.
func (c *Client) UpdateTag(id string, tag *Tag) (*Tag, error) {
	body := map[string][]Tag{"tags": {*tag}}
	var resp TagsResponse
	if err := c.do("PUT", "/tags/"+id+"/", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Tags) == 0 {
		return nil, fmt.Errorf("no tag returned")
	}
	return &resp.Tags[0], nil
}

// DeleteTag permanently deletes a tag by ID.
// This removes the tag from all posts that use it.
func (c *Client) DeleteTag(id string) error {
	return c.do("DELETE", "/tags/"+id+"/", nil, nil)
}

// Users

// ListUsers returns a list of all users on the Ghost site.
func (c *Client) ListUsers() (*UsersResponse, error) {
	var resp UsersResponse
	if err := c.do("GET", "/users/", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUser returns a single user by ID or slug.
// It first tries to find by ID, then falls back to slug lookup.
func (c *Client) GetUser(idOrSlug string) (*Author, error) {
	var resp UsersResponse
	err := c.do("GET", "/users/"+idOrSlug+"/", nil, &resp)
	if err != nil {
		err = c.do("GET", "/users/slug/"+idOrSlug+"/", nil, &resp)
		if err != nil {
			return nil, err
		}
	}
	if len(resp.Users) == 0 {
		return nil, fmt.Errorf("user not found: %s", idOrSlug)
	}
	return &resp.Users[0], nil
}

// Site

// GetSite returns information about the Ghost site including title, description, and version.
func (c *Client) GetSite() (*Site, error) {
	var resp SiteResponse
	if err := c.do("GET", "/site/", nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Site, nil
}

// Settings

// GetSettings returns the site settings as key-value pairs.
func (c *Client) GetSettings() (*SettingsResponse, error) {
	var resp SettingsResponse
	if err := c.do("GET", "/settings/", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Newsletters

// ListNewsletters returns a list of all newsletters configured on the Ghost site.
func (c *Client) ListNewsletters() (*NewslettersResponse, error) {
	var resp NewslettersResponse
	if err := c.do("GET", "/newsletters/", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetNewsletter returns a single newsletter by ID.
func (c *Client) GetNewsletter(id string) (*Newsletter, error) {
	var resp NewslettersResponse
	if err := c.do("GET", "/newsletters/"+id+"/", nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Newsletters) == 0 {
		return nil, fmt.Errorf("newsletter not found: %s", id)
	}
	return &resp.Newsletters[0], nil
}

// Webhooks

// ListWebhooks returns a list of all webhooks.
// Note: This endpoint may not be available in all Ghost versions.
func (c *Client) ListWebhooks() (*WebhooksResponse, error) {
	var resp WebhooksResponse
	if err := c.do("GET", "/webhooks/", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateWebhook creates a new webhook.
// The webhook should have Event and TargetURL set at minimum.
func (c *Client) CreateWebhook(webhook *Webhook) (*Webhook, error) {
	body := map[string][]Webhook{"webhooks": {*webhook}}
	var resp WebhooksResponse
	if err := c.do("POST", "/webhooks/", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Webhooks) == 0 {
		return nil, fmt.Errorf("no webhook returned")
	}
	return &resp.Webhooks[0], nil
}

// DeleteWebhook permanently deletes a webhook by ID.
func (c *Client) DeleteWebhook(id string) error {
	return c.do("DELETE", "/webhooks/"+id+"/", nil, nil)
}

// Images

// UploadImage uploads an image file to Ghost and returns the URL.
// The filePath should be a path to an image file on the local filesystem.
func (c *Client) UploadImage(filePath string) (*ImagesResponse, error) {
	token, err := GenerateToken(c.apiKey)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+"/images/upload/", &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Ghost "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var result ImagesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UploadImageReader uploads an image from an io.Reader.
// This is useful when the image data is not coming from a file.
// The filename parameter is used for the Content-Disposition header.
func (c *Client) UploadImageReader(r io.Reader, filename string) (*ImagesResponse, error) {
	token, err := GenerateToken(c.apiKey)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, r); err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+"/images/upload/", &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Ghost "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var result ImagesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
