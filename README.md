# libecto

Go library for the Ghost Admin API.

## Installation

```bash
go get github.com/visionik/libecto
```

## Usage

### Basic Client

```go
package main

import (
    "fmt"
    "log"

    "github.com/visionik/libecto"
)

func main() {
    client := libecto.NewClient("https://mysite.ghost.io", "your-admin-api-key")

    // List published posts
    resp, err := client.ListPosts("published", 10)
    if err != nil {
        log.Fatal(err)
    }

    for _, post := range resp.Posts {
        fmt.Printf("%s: %s\n", post.ID, post.Title)
    }
}
```

### Using Config File

```go
// Load from ~/.config/ecto/config.json
config, err := libecto.LoadConfig()
if err != nil {
    log.Fatal(err)
}

client, err := config.GetActiveClient("mysite")
if err != nil {
    log.Fatal(err)
}
```

### Environment Variables

The library respects these environment variables:
- `GHOST_URL` - Ghost site URL
- `GHOST_ADMIN_KEY` - Admin API key (id:secret format)
- `GHOST_SITE` - Default site name from config

## API Reference

### Posts

```go
// List posts by status
resp, _ := client.ListPosts("draft", 20)      // "draft", "published", "scheduled", or ""

// Get a single post
post, _ := client.GetPost("post-id-or-slug")

// Create a post
newPost, _ := client.CreatePost(&libecto.Post{
    Title:  "My Post",
    HTML:   "<p>Content here</p>",
    Status: "draft",
    Tags:   []libecto.Tag{{Name: "news"}},
})

// Update a post (include UpdatedAt for conflict detection)
updated, _ := client.UpdatePost(post.ID, &libecto.Post{
    UpdatedAt: post.UpdatedAt,
    Title:     "New Title",
})

// Delete a post
client.DeletePost(post.ID)

// Publish/unpublish
client.PublishPost("post-slug")
client.UnpublishPost("post-slug")

// Schedule for future
client.SchedulePost("post-slug", "2025-02-01T09:00:00Z")
```

### Pages

```go
resp, _ := client.ListPages("published", 10)
page, _ := client.GetPage("page-slug")
newPage, _ := client.CreatePage(&libecto.Page{Title: "About", HTML: "<p>About us</p>"})
client.UpdatePage(page.ID, &libecto.Page{UpdatedAt: page.UpdatedAt, Title: "New Title"})
client.DeletePage(page.ID)
client.PublishPage("page-slug")
```

### Tags

```go
resp, _ := client.ListTags(100)
tag, _ := client.GetTag("tag-slug")
newTag, _ := client.CreateTag(&libecto.Tag{Name: "News", Description: "Latest news"})
client.UpdateTag(tag.ID, &libecto.Tag{Name: "Updated Name"})
client.DeleteTag(tag.ID)
```

### Users

```go
resp, _ := client.ListUsers()
user, _ := client.GetUser("user-slug")
```

### Site & Settings

```go
site, _ := client.GetSite()
settings, _ := client.GetSettings()
```

### Newsletters

```go
resp, _ := client.ListNewsletters()
newsletter, _ := client.GetNewsletter("newsletter-id")
```

### Webhooks

```go
resp, _ := client.ListWebhooks()
webhook, _ := client.CreateWebhook(&libecto.Webhook{
    Event:     "post.published",
    TargetURL: "https://example.com/hook",
    Name:      "My Hook",
})
client.DeleteWebhook(webhook.ID)
```

### Images

```go
resp, _ := client.UploadImage("/path/to/image.jpg")
fmt.Println(resp.Images[0].URL)
```

### Markdown Conversion

```go
// Convert markdown to HTML (uses blackfriday)
html := libecto.MarkdownToHTML([]byte("# Hello\n\nWorld"))
html := libecto.MarkdownStringToHTML("# Hello\n\nWorld")
```

### JWT Authentication

```go
// Generate a token manually (usually not needed)
token, err := libecto.GenerateToken("admin-api-key")
```

## Types

Key types include:
- `Post`, `PostsResponse` - Blog posts
- `Page`, `PagesResponse` - Static pages
- `Tag`, `TagsResponse` - Content tags
- `Author`, `UsersResponse` - Users/authors
- `Site`, `SettingsResponse` - Site configuration
- `Newsletter`, `NewslettersResponse` - Email newsletters
- `Webhook`, `WebhooksResponse` - API webhooks
- `ImageUploadResponse` - Uploaded image info

## CLI

For command-line usage, see [ecto](https://github.com/visionik/ecto).

## License

MIT
