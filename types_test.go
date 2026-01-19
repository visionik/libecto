package libecto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPost_JSONMarshal(t *testing.T) {
	tests := []struct {
		name string
		post Post
		want map[string]interface{}
	}{
		{
			name: "minimal post",
			post: Post{
				ID:    "123",
				Title: "Test Post",
			},
			want: map[string]interface{}{
				"id":    "123",
				"title": "Test Post",
			},
		},
		{
			name: "full post",
			post: Post{
				ID:          "abc123",
				UUID:        "uuid-123",
				Title:       "Full Post",
				Slug:        "full-post",
				HTML:        "<p>Content</p>",
				Status:      "published",
				Visibility:  "public",
				PublishedAt: "2025-01-15T12:00:00Z",
				CreatedAt:   "2025-01-14T12:00:00Z",
				UpdatedAt:   "2025-01-15T12:00:00Z",
				Featured:    true,
				Tags:        []Tag{{Name: "Test"}},
				Authors:     []Author{{Name: "Author"}},
			},
			want: map[string]interface{}{
				"id":           "abc123",
				"uuid":         "uuid-123",
				"title":        "Full Post",
				"slug":         "full-post",
				"html":         "<p>Content</p>",
				"status":       "published",
				"visibility":   "public",
				"published_at": "2025-01-15T12:00:00Z",
				"created_at":   "2025-01-14T12:00:00Z",
				"updated_at":   "2025-01-15T12:00:00Z",
				"featured":     true,
			},
		},
		{
			name: "draft post",
			post: Post{
				Title:  "Draft",
				Status: "draft",
			},
			want: map[string]interface{}{
				"title":  "Draft",
				"status": "draft",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.post)
			require.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			for k, v := range tt.want {
				assert.Equal(t, v, result[k], "field %s mismatch", k)
			}
		})
	}
}

func TestPost_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Post
		wantErr bool
	}{
		{
			name:  "minimal post",
			input: `{"id":"123","title":"Test"}`,
			want:  Post{ID: "123", Title: "Test"},
		},
		{
			name: "full post",
			input: `{
				"id": "abc",
				"uuid": "uuid-abc",
				"title": "Full",
				"slug": "full",
				"html": "<p>Test</p>",
				"status": "published",
				"visibility": "public",
				"published_at": "2025-01-15T12:00:00Z",
				"featured": true,
				"tags": [{"name": "Tag1"}],
				"authors": [{"name": "Author1"}]
			}`,
			want: Post{
				ID:          "abc",
				UUID:        "uuid-abc",
				Title:       "Full",
				Slug:        "full",
				HTML:        "<p>Test</p>",
				Status:      "published",
				Visibility:  "public",
				PublishedAt: "2025-01-15T12:00:00Z",
				Featured:    true,
				Tags:        []Tag{{Name: "Tag1"}},
				Authors:     []Author{{Name: "Author1"}},
			},
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var post Post
			err := json.Unmarshal([]byte(tt.input), &post)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, post)
		})
	}
}

func TestPostsResponse_JSONUnmarshal(t *testing.T) {
	input := `{
		"posts": [
			{"id": "1", "title": "Post 1"},
			{"id": "2", "title": "Post 2"}
		],
		"meta": {
			"pagination": {
				"page": 1,
				"limit": 15,
				"pages": 2,
				"total": 25,
				"next": 2,
				"prev": null
			}
		}
	}`

	var resp PostsResponse
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Posts, 2)
	assert.Equal(t, "1", resp.Posts[0].ID)
	assert.Equal(t, "Post 1", resp.Posts[0].Title)
	assert.NotNil(t, resp.Meta)
	assert.Equal(t, 1, resp.Meta.Pagination.Page)
	assert.Equal(t, 25, resp.Meta.Pagination.Total)
	assert.NotNil(t, resp.Meta.Pagination.Next)
	assert.Equal(t, 2, *resp.Meta.Pagination.Next)
	assert.Nil(t, resp.Meta.Pagination.Prev)
}

func TestPage_JSONRoundtrip(t *testing.T) {
	original := Page{
		ID:          "page123",
		Title:       "About Us",
		Slug:        "about-us",
		HTML:        "<h1>About</h1><p>We are a company.</p>",
		Status:      "published",
		Visibility:  "public",
		PublishedAt: "2025-01-01T00:00:00Z",
		CreatedAt:   "2024-12-01T00:00:00Z",
		UpdatedAt:   "2025-01-01T00:00:00Z",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Page
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original, decoded)
}

func TestTag_JSONRoundtrip(t *testing.T) {
	original := Tag{
		ID:          "tag123",
		Name:        "Technology",
		Slug:        "technology",
		Description: "Tech posts",
		Visibility:  "public",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Tag
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original, decoded)
}

func TestAuthor_JSONRoundtrip(t *testing.T) {
	original := Author{
		ID:           "author123",
		Name:         "John Doe",
		Slug:         "john-doe",
		Email:        "john@example.com",
		Bio:          "A writer",
		Location:     "New York",
		Website:      "https://johndoe.com",
		Twitter:      "@johndoe",
		Facebook:     "johndoe",
		ProfileImage: "https://example.com/john.jpg",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Author
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original, decoded)
}

func TestSite_JSONUnmarshal(t *testing.T) {
	input := `{
		"title": "My Blog",
		"description": "A great blog",
		"logo": "https://example.com/logo.png",
		"icon": "https://example.com/favicon.ico",
		"url": "https://example.com",
		"version": "5.0.0"
	}`

	var site Site
	err := json.Unmarshal([]byte(input), &site)
	require.NoError(t, err)

	assert.Equal(t, "My Blog", site.Title)
	assert.Equal(t, "A great blog", site.Description)
	assert.Equal(t, "https://example.com", site.URL)
	assert.Equal(t, "5.0.0", site.Version)
}

func TestSetting_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKey   string
		wantValue interface{}
	}{
		{
			name:      "string value",
			input:     `{"key": "title", "value": "My Site"}`,
			wantKey:   "title",
			wantValue: "My Site",
		},
		{
			name:      "bool value",
			input:     `{"key": "is_private", "value": true}`,
			wantKey:   "is_private",
			wantValue: true,
		},
		{
			name:      "null value",
			input:     `{"key": "empty", "value": null}`,
			wantKey:   "empty",
			wantValue: nil,
		},
		{
			name:      "number value",
			input:     `{"key": "count", "value": 42}`,
			wantKey:   "count",
			wantValue: float64(42), // JSON numbers are float64
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var setting Setting
			err := json.Unmarshal([]byte(tt.input), &setting)
			require.NoError(t, err)
			assert.Equal(t, tt.wantKey, setting.Key)
			assert.Equal(t, tt.wantValue, setting.Value)
		})
	}
}

func TestNewsletter_JSONRoundtrip(t *testing.T) {
	original := Newsletter{
		ID:                "nl123",
		Name:              "Weekly Digest",
		Description:       "Weekly updates",
		Status:            "active",
		Slug:              "weekly-digest",
		SenderName:        "Blog Team",
		SenderEmail:       "team@example.com",
		SenderReplyTo:     "newsletter",
		SubscribeOnSignup: true,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Newsletter
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original, decoded)
}

func TestWebhook_JSONRoundtrip(t *testing.T) {
	original := Webhook{
		ID:            "wh123",
		Event:         "post.published",
		TargetURL:     "https://example.com/webhook",
		Name:          "Publish Notifier",
		Status:        "active",
		IntegrationID: "int123",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Webhook
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original, decoded)
}

func TestImage_JSONRoundtrip(t *testing.T) {
	original := Image{
		URL: "https://example.com/image.jpg",
		Ref: "image-ref-123",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Image
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original, decoded)
}

func TestPagination_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Pagination
	}{
		{
			name:  "first page",
			input: `{"page": 1, "limit": 15, "pages": 5, "total": 75, "next": 2, "prev": null}`,
			want:  Pagination{Page: 1, Limit: 15, Pages: 5, Total: 75, Next: intPtr(2), Prev: nil},
		},
		{
			name:  "middle page",
			input: `{"page": 3, "limit": 15, "pages": 5, "total": 75, "next": 4, "prev": 2}`,
			want:  Pagination{Page: 3, Limit: 15, Pages: 5, Total: 75, Next: intPtr(4), Prev: intPtr(2)},
		},
		{
			name:  "last page",
			input: `{"page": 5, "limit": 15, "pages": 5, "total": 75, "next": null, "prev": 4}`,
			want:  Pagination{Page: 5, Limit: 15, Pages: 5, Total: 75, Next: nil, Prev: intPtr(4)},
		},
		{
			name:  "single page",
			input: `{"page": 1, "limit": 15, "pages": 1, "total": 5, "next": null, "prev": null}`,
			want:  Pagination{Page: 1, Limit: 15, Pages: 1, Total: 5, Next: nil, Prev: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Pagination
			err := json.Unmarshal([]byte(tt.input), &p)
			require.NoError(t, err)
			assert.Equal(t, tt.want.Page, p.Page)
			assert.Equal(t, tt.want.Limit, p.Limit)
			assert.Equal(t, tt.want.Pages, p.Pages)
			assert.Equal(t, tt.want.Total, p.Total)
			if tt.want.Next == nil {
				assert.Nil(t, p.Next)
			} else {
				require.NotNil(t, p.Next)
				assert.Equal(t, *tt.want.Next, *p.Next)
			}
			if tt.want.Prev == nil {
				assert.Nil(t, p.Prev)
			} else {
				require.NotNil(t, p.Prev)
				assert.Equal(t, *tt.want.Prev, *p.Prev)
			}
		})
	}
}

func TestAPIError_JSONUnmarshal(t *testing.T) {
	input := `{"message": "Resource not found", "context": "Post with id 123", "type": "NotFoundError"}`

	var apiErr APIError
	err := json.Unmarshal([]byte(input), &apiErr)
	require.NoError(t, err)

	assert.Equal(t, "Resource not found", apiErr.Message)
	assert.Equal(t, "Post with id 123", apiErr.Context)
	assert.Equal(t, "NotFoundError", apiErr.Type)
}

func TestErrorResponse_JSONUnmarshal(t *testing.T) {
	input := `{
		"errors": [
			{"message": "Validation failed", "context": "Title is required", "type": "ValidationError"},
			{"message": "Invalid format", "context": "Status must be draft or published", "type": "ValidationError"}
		]
	}`

	var resp ErrorResponse
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Errors, 2)
	assert.Equal(t, "Validation failed", resp.Errors[0].Message)
	assert.Equal(t, "Invalid format", resp.Errors[1].Message)
}

func TestPagesResponse_JSONUnmarshal(t *testing.T) {
	input := `{
		"pages": [
			{"id": "1", "title": "About"},
			{"id": "2", "title": "Contact"}
		]
	}`

	var resp PagesResponse
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Pages, 2)
	assert.Equal(t, "About", resp.Pages[0].Title)
	assert.Equal(t, "Contact", resp.Pages[1].Title)
}

func TestTagsResponse_JSONUnmarshal(t *testing.T) {
	input := `{"tags": [{"id": "1", "name": "News"}, {"id": "2", "name": "Tech"}]}`

	var resp TagsResponse
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Tags, 2)
}

func TestUsersResponse_JSONUnmarshal(t *testing.T) {
	input := `{"users": [{"id": "1", "name": "Admin"}, {"id": "2", "name": "Editor"}]}`

	var resp UsersResponse
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Users, 2)
}

func TestSiteResponse_JSONUnmarshal(t *testing.T) {
	input := `{"site": {"title": "My Blog", "url": "https://example.com", "version": "5.0"}}`

	var resp SiteResponse
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)

	assert.Equal(t, "My Blog", resp.Site.Title)
}

func TestSettingsResponse_JSONUnmarshal(t *testing.T) {
	input := `{"settings": [{"key": "title", "value": "Blog"}, {"key": "description", "value": "A blog"}]}`

	var resp SettingsResponse
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Settings, 2)
}

func TestNewslettersResponse_JSONUnmarshal(t *testing.T) {
	input := `{"newsletters": [{"id": "1", "name": "Weekly"}]}`

	var resp NewslettersResponse
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Newsletters, 1)
}

func TestWebhooksResponse_JSONUnmarshal(t *testing.T) {
	input := `{"webhooks": [{"id": "1", "event": "post.published"}]}`

	var resp WebhooksResponse
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Webhooks, 1)
}

func TestImagesResponse_JSONUnmarshal(t *testing.T) {
	input := `{"images": [{"url": "https://example.com/img.jpg"}]}`

	var resp ImagesResponse
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Images, 1)
}

func TestMeta_JSONRoundtrip(t *testing.T) {
	original := Meta{
		Pagination: Pagination{
			Page:  2,
			Limit: 10,
			Pages: 5,
			Total: 50,
			Next:  intPtr(3),
			Prev:  intPtr(1),
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Meta
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Pagination.Page, decoded.Pagination.Page)
	assert.Equal(t, original.Pagination.Total, decoded.Pagination.Total)
}

// Helper for pointer to int
func intPtr(i int) *int {
	return &i
}
