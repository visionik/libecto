package libecto

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAPIKey = "test123:7365637265746b6579313233"

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		apiKey  string
		wantURL string
	}{
		{
			name:    "basic URL",
			url:     "https://example.ghost.io",
			apiKey:  testAPIKey,
			wantURL: "https://example.ghost.io/ghost/api/admin",
		},
		{
			name:    "URL with trailing slash",
			url:     "https://example.ghost.io/",
			apiKey:  testAPIKey,
			wantURL: "https://example.ghost.io/ghost/api/admin",
		},
		{
			name:    "URL with multiple trailing slashes",
			url:     "https://example.ghost.io///",
			apiKey:  testAPIKey,
			wantURL: "https://example.ghost.io///ghost/api/admin", // only removes one
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.url, tt.apiKey)
			assert.Equal(t, tt.wantURL, client.BaseURL())
		})
	}
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	customClient := &http.Client{}
	client := NewClient("https://example.com", testAPIKey, WithHTTPClient(customClient))
	assert.Same(t, customClient, client.httpClient)
}

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := NewClient(strings.TrimSuffix(server.URL, "/ghost/api/admin"), testAPIKey)
	return server, client
}

// Posts tests

func TestClient_ListPosts(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		limit      int
		response   PostsResponse
		statusCode int
		wantErr    bool
		checkPath  func(t *testing.T, path string)
	}{
		{
			name:       "list all posts",
			status:     "",
			limit:      0,
			response:   PostsResponse{Posts: []Post{{ID: "1", Title: "Test"}}},
			statusCode: 200,
			checkPath: func(t *testing.T, path string) {
				assert.Contains(t, path, "/posts/")
				assert.Contains(t, path, "formats=html")
				assert.NotContains(t, path, "filter=status")
			},
		},
		{
			name:       "filter by published",
			status:     "published",
			limit:      0,
			response:   PostsResponse{Posts: []Post{{ID: "2", Status: "published"}}},
			statusCode: 200,
			checkPath: func(t *testing.T, path string) {
				assert.Contains(t, path, "filter=status:published")
			},
		},
		{
			name:       "with limit",
			status:     "",
			limit:      10,
			response:   PostsResponse{Posts: []Post{}},
			statusCode: 200,
			checkPath: func(t *testing.T, path string) {
				assert.Contains(t, path, "limit=10")
			},
		},
		{
			name:       "status all",
			status:     "all",
			limit:      0,
			response:   PostsResponse{Posts: []Post{{ID: "1"}, {ID: "2"}}},
			statusCode: 200,
			checkPath: func(t *testing.T, path string) {
				assert.NotContains(t, path, "filter=status")
			},
		},
		{
			name:       "server error",
			statusCode: 500,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.Header.Get("Authorization"), "Ghost ")
				if tt.checkPath != nil {
					tt.checkPath(t, r.URL.String())
				}
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					json.NewEncoder(w).Encode(tt.response)
				} else {
					w.Write([]byte("error"))
				}
			})
			defer server.Close()

			resp, err := client.ListPosts(tt.status, tt.limit)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, len(tt.response.Posts), len(resp.Posts))
		})
	}
}

func TestClient_GetPost(t *testing.T) {
	tests := []struct {
		name       string
		idOrSlug   string
		response   PostsResponse
		statusCode int
		wantErr    bool
		errContain string
	}{
		{
			name:       "get by ID",
			idOrSlug:   "abc123",
			response:   PostsResponse{Posts: []Post{{ID: "abc123", Title: "Test"}}},
			statusCode: 200,
		},
		{
			name:       "not found",
			idOrSlug:   "nonexistent",
			statusCode: 404,
			wantErr:    true,
		},
		{
			name:       "empty response",
			idOrSlug:   "empty",
			response:   PostsResponse{Posts: []Post{}},
			statusCode: 200,
			wantErr:    true,
			errContain: "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					json.NewEncoder(w).Encode(tt.response)
				} else {
					json.NewEncoder(w).Encode(ErrorResponse{Errors: []APIError{{Message: "Not found"}}})
				}
			})
			defer server.Close()

			post, err := client.GetPost(tt.idOrSlug)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.response.Posts[0].ID, post.ID)
		})
	}
}

func TestClient_CreatePost(t *testing.T) {
	tests := []struct {
		name       string
		post       *Post
		response   PostsResponse
		statusCode int
		wantErr    bool
	}{
		{
			name:       "create post",
			post:       &Post{Title: "New Post", Status: "draft"},
			response:   PostsResponse{Posts: []Post{{ID: "new123", Title: "New Post"}}},
			statusCode: 201,
		},
		{
			name:       "validation error",
			post:       &Post{},
			statusCode: 422,
			wantErr:    true,
		},
		{
			name:       "empty response",
			post:       &Post{Title: "Test"},
			response:   PostsResponse{Posts: []Post{}},
			statusCode: 201,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				w.WriteHeader(tt.statusCode)
				if tt.statusCode < 400 {
					json.NewEncoder(w).Encode(tt.response)
				}
			})
			defer server.Close()

			created, err := client.CreatePost(tt.post)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.response.Posts[0].ID, created.ID)
		})
	}
}

func TestClient_UpdatePost(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Contains(t, r.URL.Path, "/posts/123/")
		json.NewEncoder(w).Encode(PostsResponse{Posts: []Post{{ID: "123", Title: "Updated"}}})
	})
	defer server.Close()

	updated, err := client.UpdatePost("123", &Post{Title: "Updated"})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)
}

func TestClient_DeletePost(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Contains(t, r.URL.Path, "/posts/123/")
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeletePost("123")
	require.NoError(t, err)
}

func TestClient_PublishPost(t *testing.T) {
	callCount := 0
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First call: GetPost
			json.NewEncoder(w).Encode(PostsResponse{Posts: []Post{{ID: "123", UpdatedAt: "2025-01-15"}}})
		} else {
			// Second call: UpdatePost
			assert.Equal(t, "PUT", r.Method)
			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), `"status":"published"`)
			json.NewEncoder(w).Encode(PostsResponse{Posts: []Post{{ID: "123", Status: "published"}}})
		}
	})
	defer server.Close()

	post, err := client.PublishPost("123")
	require.NoError(t, err)
	assert.Equal(t, "published", post.Status)
}

func TestClient_UnpublishPost(t *testing.T) {
	callCount := 0
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(PostsResponse{Posts: []Post{{ID: "123", UpdatedAt: "2025-01-15"}}})
		} else {
			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), `"status":"draft"`)
			json.NewEncoder(w).Encode(PostsResponse{Posts: []Post{{ID: "123", Status: "draft"}}})
		}
	})
	defer server.Close()

	post, err := client.UnpublishPost("123")
	require.NoError(t, err)
	assert.Equal(t, "draft", post.Status)
}

func TestClient_SchedulePost(t *testing.T) {
	callCount := 0
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(PostsResponse{Posts: []Post{{ID: "123", UpdatedAt: "2025-01-15"}}})
		} else {
			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), `"status":"scheduled"`)
			assert.Contains(t, string(body), `"published_at":"2025-02-01T12:00:00Z"`)
			json.NewEncoder(w).Encode(PostsResponse{Posts: []Post{{ID: "123", Status: "scheduled"}}})
		}
	})
	defer server.Close()

	post, err := client.SchedulePost("123", "2025-02-01T12:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, "scheduled", post.Status)
}

// Pages tests

func TestClient_ListPages(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		json.NewEncoder(w).Encode(PagesResponse{Pages: []Page{{ID: "1", Title: "About"}}})
	})
	defer server.Close()

	resp, err := client.ListPages("", 0)
	require.NoError(t, err)
	assert.Len(t, resp.Pages, 1)
}

func TestClient_GetPage(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PagesResponse{Pages: []Page{{ID: "123", Title: "About"}}})
	})
	defer server.Close()

	page, err := client.GetPage("123")
	require.NoError(t, err)
	assert.Equal(t, "About", page.Title)
}

func TestClient_CreatePage(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(PagesResponse{Pages: []Page{{ID: "new", Title: "New Page"}}})
	})
	defer server.Close()

	page, err := client.CreatePage(&Page{Title: "New Page"})
	require.NoError(t, err)
	assert.Equal(t, "new", page.ID)
}

func TestClient_UpdatePage(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		json.NewEncoder(w).Encode(PagesResponse{Pages: []Page{{ID: "123", Title: "Updated"}}})
	})
	defer server.Close()

	page, err := client.UpdatePage("123", &Page{Title: "Updated"})
	require.NoError(t, err)
	assert.Equal(t, "Updated", page.Title)
}

func TestClient_DeletePage(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeletePage("123")
	require.NoError(t, err)
}

func TestClient_PublishPage(t *testing.T) {
	callCount := 0
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(PagesResponse{Pages: []Page{{ID: "123", UpdatedAt: "2025-01-15"}}})
		} else {
			json.NewEncoder(w).Encode(PagesResponse{Pages: []Page{{ID: "123", Status: "published"}}})
		}
	})
	defer server.Close()

	page, err := client.PublishPage("123")
	require.NoError(t, err)
	assert.Equal(t, "published", page.Status)
}

// Tags tests

func TestClient_ListTags(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "include=count.posts")
		json.NewEncoder(w).Encode(TagsResponse{Tags: []Tag{{ID: "1", Name: "News"}}})
	})
	defer server.Close()

	resp, err := client.ListTags(0)
	require.NoError(t, err)
	assert.Len(t, resp.Tags, 1)
}

func TestClient_GetTag(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TagsResponse{Tags: []Tag{{ID: "123", Name: "Tech"}}})
	})
	defer server.Close()

	tag, err := client.GetTag("123")
	require.NoError(t, err)
	assert.Equal(t, "Tech", tag.Name)
}

func TestClient_CreateTag(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(TagsResponse{Tags: []Tag{{ID: "new", Name: "New Tag"}}})
	})
	defer server.Close()

	tag, err := client.CreateTag(&Tag{Name: "New Tag"})
	require.NoError(t, err)
	assert.Equal(t, "new", tag.ID)
}

func TestClient_UpdateTag(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		json.NewEncoder(w).Encode(TagsResponse{Tags: []Tag{{ID: "123", Name: "Updated"}}})
	})
	defer server.Close()

	tag, err := client.UpdateTag("123", &Tag{Name: "Updated"})
	require.NoError(t, err)
	assert.Equal(t, "Updated", tag.Name)
}

func TestClient_DeleteTag(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteTag("123")
	require.NoError(t, err)
}

// Users tests

func TestClient_ListUsers(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(UsersResponse{Users: []Author{{ID: "1", Name: "Admin"}}})
	})
	defer server.Close()

	resp, err := client.ListUsers()
	require.NoError(t, err)
	assert.Len(t, resp.Users, 1)
}

func TestClient_GetUser(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(UsersResponse{Users: []Author{{ID: "123", Name: "John"}}})
	})
	defer server.Close()

	user, err := client.GetUser("123")
	require.NoError(t, err)
	assert.Equal(t, "John", user.Name)
}

// Site tests

func TestClient_GetSite(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SiteResponse{Site: Site{Title: "My Blog", Version: "5.0"}})
	})
	defer server.Close()

	site, err := client.GetSite()
	require.NoError(t, err)
	assert.Equal(t, "My Blog", site.Title)
}

// Settings tests

func TestClient_GetSettings(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SettingsResponse{Settings: []Setting{{Key: "title", Value: "Blog"}}})
	})
	defer server.Close()

	resp, err := client.GetSettings()
	require.NoError(t, err)
	assert.Len(t, resp.Settings, 1)
}

// Newsletters tests

func TestClient_ListNewsletters(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(NewslettersResponse{Newsletters: []Newsletter{{ID: "1", Name: "Weekly"}}})
	})
	defer server.Close()

	resp, err := client.ListNewsletters()
	require.NoError(t, err)
	assert.Len(t, resp.Newsletters, 1)
}

func TestClient_GetNewsletter(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(NewslettersResponse{Newsletters: []Newsletter{{ID: "123", Name: "Weekly"}}})
	})
	defer server.Close()

	nl, err := client.GetNewsletter("123")
	require.NoError(t, err)
	assert.Equal(t, "Weekly", nl.Name)
}

func TestClient_GetNewsletter_NotFound(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(NewslettersResponse{Newsletters: []Newsletter{}})
	})
	defer server.Close()

	_, err := client.GetNewsletter("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "newsletter not found")
}

// Webhooks tests

func TestClient_ListWebhooks(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(WebhooksResponse{Webhooks: []Webhook{{ID: "1", Event: "post.published"}}})
	})
	defer server.Close()

	resp, err := client.ListWebhooks()
	require.NoError(t, err)
	assert.Len(t, resp.Webhooks, 1)
}

func TestClient_CreateWebhook(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(WebhooksResponse{Webhooks: []Webhook{{ID: "new", Event: "post.published"}}})
	})
	defer server.Close()

	wh, err := client.CreateWebhook(&Webhook{Event: "post.published", TargetURL: "http://example.com"})
	require.NoError(t, err)
	assert.Equal(t, "new", wh.ID)
}

func TestClient_DeleteWebhook(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(204)
	})
	defer server.Close()

	err := client.DeleteWebhook("123")
	require.NoError(t, err)
}

// Images tests

func TestClient_UploadImage(t *testing.T) {
	// Create a temp file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.jpg")
	err := os.WriteFile(tmpFile, []byte("fake image data"), 0644)
	require.NoError(t, err)

	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
		json.NewEncoder(w).Encode(ImagesResponse{Images: []Image{{URL: "https://example.com/test.jpg"}}})
	})
	defer server.Close()

	resp, err := client.UploadImage(tmpFile)
	require.NoError(t, err)
	assert.Len(t, resp.Images, 1)
	assert.Equal(t, "https://example.com/test.jpg", resp.Images[0].URL)
}

func TestClient_UploadImage_FileNotFound(t *testing.T) {
	client := NewClient("http://localhost", testAPIKey)
	_, err := client.UploadImage("/nonexistent/file.jpg")
	require.Error(t, err)
}

func TestClient_UploadImageReader(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		json.NewEncoder(w).Encode(ImagesResponse{Images: []Image{{URL: "https://example.com/upload.jpg"}}})
	})
	defer server.Close()

	reader := strings.NewReader("fake image data")
	resp, err := client.UploadImageReader(reader, "upload.jpg")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/upload.jpg", resp.Images[0].URL)
}

// Error handling tests

func TestClient_APIError(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(ErrorResponse{
			Errors: []APIError{{Message: "Validation failed", Context: "Title is required"}},
		})
	})
	defer server.Close()

	_, err := client.ListPosts("", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Validation failed")
	assert.Contains(t, err.Error(), "Title is required")
}

func TestClient_APIError_NoContext(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(ErrorResponse{
			Errors: []APIError{{Message: "Bad request"}},
		})
	})
	defer server.Close()

	_, err := client.ListPosts("", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Bad request")
}

func TestClient_APIError_NonJSON(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
	})
	defer server.Close()

	_, err := client.ListPosts("", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestClient_InvalidAPIKey(t *testing.T) {
	client := NewClient("http://localhost", "invalid-key")
	_, err := client.ListPosts("", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "generating token")
}

// Edge cases for empty responses

func TestClient_CreatePost_EmptyResponse(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(PostsResponse{Posts: []Post{}})
	})
	defer server.Close()

	_, err := client.CreatePost(&Post{Title: "Test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no post returned")
}

func TestClient_CreatePage_EmptyResponse(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(PagesResponse{Pages: []Page{}})
	})
	defer server.Close()

	_, err := client.CreatePage(&Page{Title: "Test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no page returned")
}

func TestClient_CreateTag_EmptyResponse(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(TagsResponse{Tags: []Tag{}})
	})
	defer server.Close()

	_, err := client.CreateTag(&Tag{Name: "Test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no tag returned")
}

func TestClient_CreateWebhook_EmptyResponse(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(WebhooksResponse{Webhooks: []Webhook{}})
	})
	defer server.Close()

	_, err := client.CreateWebhook(&Webhook{Event: "test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no webhook returned")
}

func TestClient_UpdatePost_EmptyResponse(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PostsResponse{Posts: []Post{}})
	})
	defer server.Close()

	_, err := client.UpdatePost("123", &Post{Title: "Test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no post returned")
}

func TestClient_UpdatePage_EmptyResponse(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PagesResponse{Pages: []Page{}})
	})
	defer server.Close()

	_, err := client.UpdatePage("123", &Page{Title: "Test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no page returned")
}

func TestClient_UpdateTag_EmptyResponse(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TagsResponse{Tags: []Tag{}})
	})
	defer server.Close()

	_, err := client.UpdateTag("123", &Tag{Name: "Test"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no tag returned")
}

func TestClient_GetPage_NotFound(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(ErrorResponse{Errors: []APIError{{Message: "Not found"}}})
	})
	defer server.Close()

	_, err := client.GetPage("nonexistent")
	require.Error(t, err)
}

func TestClient_GetTag_NotFound(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TagsResponse{Tags: []Tag{}})
	})
	defer server.Close()

	_, err := client.GetTag("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tag not found")
}

func TestClient_GetUser_NotFound(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(UsersResponse{Users: []Author{}})
	})
	defer server.Close()

	_, err := client.GetUser("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestClient_UploadImage_ServerError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.jpg")
	os.WriteFile(tmpFile, []byte("data"), 0644)

	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("Server error"))
	})
	defer server.Close()

	_, err := client.UploadImage(tmpFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upload failed")
}

func TestClient_UploadImageReader_ServerError(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("Server error"))
	})
	defer server.Close()

	_, err := client.UploadImageReader(strings.NewReader("data"), "test.jpg")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upload failed")
}

// Fallback path tests (ID -> slug)

func TestClient_GetPost_FallbackToSlug(t *testing.T) {
	callCount := 0
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First try by ID fails
			w.WriteHeader(404)
			json.NewEncoder(w).Encode(ErrorResponse{Errors: []APIError{{Message: "Not found"}}})
		} else {
			// Second try by slug succeeds
			json.NewEncoder(w).Encode(PostsResponse{Posts: []Post{{ID: "123", Slug: "my-post"}}})
		}
	})
	defer server.Close()

	post, err := client.GetPost("my-post")
	require.NoError(t, err)
	assert.Equal(t, "123", post.ID)
	assert.Equal(t, 2, callCount)
}

func TestClient_GetPage_FallbackToSlug(t *testing.T) {
	callCount := 0
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(404)
			json.NewEncoder(w).Encode(ErrorResponse{Errors: []APIError{{Message: "Not found"}}})
		} else {
			json.NewEncoder(w).Encode(PagesResponse{Pages: []Page{{ID: "123", Slug: "about"}}})
		}
	})
	defer server.Close()

	page, err := client.GetPage("about")
	require.NoError(t, err)
	assert.Equal(t, "123", page.ID)
}

func TestClient_GetTag_FallbackToSlug(t *testing.T) {
	callCount := 0
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(404)
			json.NewEncoder(w).Encode(ErrorResponse{Errors: []APIError{{Message: "Not found"}}})
		} else {
			json.NewEncoder(w).Encode(TagsResponse{Tags: []Tag{{ID: "123", Slug: "tech"}}})
		}
	})
	defer server.Close()

	tag, err := client.GetTag("tech")
	require.NoError(t, err)
	assert.Equal(t, "123", tag.ID)
}

func TestClient_GetUser_FallbackToSlug(t *testing.T) {
	callCount := 0
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(404)
			json.NewEncoder(w).Encode(ErrorResponse{Errors: []APIError{{Message: "Not found"}}})
		} else {
			json.NewEncoder(w).Encode(UsersResponse{Users: []Author{{ID: "123", Slug: "john"}}})
		}
	})
	defer server.Close()

	user, err := client.GetUser("john")
	require.NoError(t, err)
	assert.Equal(t, "123", user.ID)
}
