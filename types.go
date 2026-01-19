package libecto

// Post represents a Ghost blog post with all standard fields.
// Posts are the primary content type in Ghost and support various statuses,
// visibility settings, and associations with tags and authors.
type Post struct {
	// ID is the unique identifier for the post.
	ID string `json:"id,omitempty"`
	// UUID is the universally unique identifier.
	UUID string `json:"uuid,omitempty"`
	// Title is the post title displayed to readers.
	Title string `json:"title,omitempty"`
	// Slug is the URL-friendly version of the title.
	Slug string `json:"slug,omitempty"`
	// HTML is the rendered HTML content of the post.
	HTML string `json:"html,omitempty"`
	// Mobiledoc is the internal document format used by Ghost.
	Mobiledoc string `json:"mobiledoc,omitempty"`
	// Status indicates the publication state: draft, published, or scheduled.
	Status string `json:"status,omitempty"`
	// Visibility controls who can see the post: public, members, paid, or tiers.
	Visibility string `json:"visibility,omitempty"`
	// PublishedAt is the publication timestamp in ISO8601 format.
	PublishedAt string `json:"published_at,omitempty"`
	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at,omitempty"`
	// UpdatedAt is the last modification timestamp.
	UpdatedAt string `json:"updated_at,omitempty"`
	// Excerpt is an auto-generated summary of the post.
	Excerpt string `json:"excerpt,omitempty"`
	// CustomExcerpt is a manually set summary.
	CustomExcerpt string `json:"custom_excerpt,omitempty"`
	// FeatureImage is the URL of the post's featured image.
	FeatureImage string `json:"feature_image,omitempty"`
	// Featured indicates whether this is a featured/pinned post.
	Featured bool `json:"featured,omitempty"`
	// Tags is the list of tags associated with the post.
	Tags []Tag `json:"tags,omitempty"`
	// Authors is the list of authors for the post.
	Authors []Author `json:"authors,omitempty"`
}

// PostsResponse is the API response structure for post listings.
// It contains an array of posts and optional pagination metadata.
type PostsResponse struct {
	// Posts is the array of returned posts.
	Posts []Post `json:"posts"`
	// Meta contains pagination information when available.
	Meta *Meta `json:"meta,omitempty"`
}

// Page represents a Ghost page, which is a static content type.
// Pages have similar fields to posts but are typically used for
// non-chronological content like About or Contact pages.
type Page struct {
	// ID is the unique identifier for the page.
	ID string `json:"id,omitempty"`
	// UUID is the universally unique identifier.
	UUID string `json:"uuid,omitempty"`
	// Title is the page title.
	Title string `json:"title,omitempty"`
	// Slug is the URL-friendly version of the title.
	Slug string `json:"slug,omitempty"`
	// HTML is the rendered HTML content.
	HTML string `json:"html,omitempty"`
	// Mobiledoc is the internal document format.
	Mobiledoc string `json:"mobiledoc,omitempty"`
	// Status indicates the publication state.
	Status string `json:"status,omitempty"`
	// Visibility controls who can see the page.
	Visibility string `json:"visibility,omitempty"`
	// PublishedAt is the publication timestamp.
	PublishedAt string `json:"published_at,omitempty"`
	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at,omitempty"`
	// UpdatedAt is the last modification timestamp.
	UpdatedAt string `json:"updated_at,omitempty"`
	// FeatureImage is the URL of the featured image.
	FeatureImage string `json:"feature_image,omitempty"`
	// Tags is the list of associated tags.
	Tags []Tag `json:"tags,omitempty"`
	// Authors is the list of authors.
	Authors []Author `json:"authors,omitempty"`
}

// PagesResponse is the API response structure for page listings.
type PagesResponse struct {
	// Pages is the array of returned pages.
	Pages []Page `json:"pages"`
	// Meta contains pagination information.
	Meta *Meta `json:"meta,omitempty"`
}

// Tag represents a Ghost tag used to categorize content.
// Tags can be public or internal (starting with #).
type Tag struct {
	// ID is the unique identifier.
	ID string `json:"id,omitempty"`
	// Name is the display name of the tag.
	Name string `json:"name,omitempty"`
	// Slug is the URL-friendly version.
	Slug string `json:"slug,omitempty"`
	// Description provides additional information about the tag.
	Description string `json:"description,omitempty"`
	// FeatureImage is the URL of the tag's image.
	FeatureImage string `json:"feature_image,omitempty"`
	// Visibility controls whether the tag is public or internal.
	Visibility string `json:"visibility,omitempty"`
	// PostCount is the number of posts using this tag.
	PostCount int `json:"count.posts,omitempty"`
}

// TagsResponse is the API response structure for tag listings.
type TagsResponse struct {
	// Tags is the array of returned tags.
	Tags []Tag `json:"tags"`
	// Meta contains pagination information.
	Meta *Meta `json:"meta,omitempty"`
}

// Author represents a Ghost user who can create content.
// Authors have profiles with optional social links and biographical information.
type Author struct {
	// ID is the unique identifier.
	ID string `json:"id,omitempty"`
	// Name is the author's display name.
	Name string `json:"name,omitempty"`
	// Slug is the URL-friendly version of the name.
	Slug string `json:"slug,omitempty"`
	// Email is the author's email address.
	Email string `json:"email,omitempty"`
	// Bio is a short biographical description.
	Bio string `json:"bio,omitempty"`
	// Location is the author's location.
	Location string `json:"location,omitempty"`
	// Website is the author's personal website URL.
	Website string `json:"website,omitempty"`
	// Twitter is the author's Twitter username.
	Twitter string `json:"twitter,omitempty"`
	// Facebook is the author's Facebook profile.
	Facebook string `json:"facebook,omitempty"`
	// ProfileImage is the URL of the author's profile picture.
	ProfileImage string `json:"profile_image,omitempty"`
}

// UsersResponse is the API response structure for user listings.
type UsersResponse struct {
	// Users is the array of returned users.
	Users []Author `json:"users"`
	// Meta contains pagination information.
	Meta *Meta `json:"meta,omitempty"`
}

// Site represents Ghost site configuration and metadata.
// It contains general information about the Ghost installation.
type Site struct {
	// Title is the site name.
	Title string `json:"title"`
	// Description is the site tagline or description.
	Description string `json:"description"`
	// Logo is the URL of the site logo.
	Logo string `json:"logo"`
	// Icon is the URL of the site favicon.
	Icon string `json:"icon"`
	// URL is the site's public URL.
	URL string `json:"url"`
	// Version is the Ghost version number.
	Version string `json:"version"`
}

// SiteResponse is the API response structure for site information.
type SiteResponse struct {
	// Site contains the site information.
	Site Site `json:"site"`
}

// Setting represents a Ghost configuration setting as a key-value pair.
// Settings control various aspects of site behavior and appearance.
type Setting struct {
	// Key is the setting identifier.
	Key string `json:"key"`
	// Value is the setting value, which can be various types.
	Value interface{} `json:"value"`
}

// SettingsResponse is the API response structure for site settings.
type SettingsResponse struct {
	// Settings is the array of setting key-value pairs.
	Settings []Setting `json:"settings"`
}

// Newsletter represents a Ghost newsletter configuration.
// Newsletters are used for email distribution to subscribers.
type Newsletter struct {
	// ID is the unique identifier.
	ID string `json:"id"`
	// Name is the newsletter display name.
	Name string `json:"name"`
	// Description provides information about the newsletter.
	Description string `json:"description"`
	// Status indicates whether the newsletter is active or archived.
	Status string `json:"status"`
	// Slug is the URL-friendly identifier.
	Slug string `json:"slug"`
	// SenderName is the name shown in sent emails.
	SenderName string `json:"sender_name,omitempty"`
	// SenderEmail is the reply-to email address.
	SenderEmail string `json:"sender_email,omitempty"`
	// SenderReplyTo configures reply behavior.
	SenderReplyTo string `json:"sender_reply_to,omitempty"`
	// SubscribeOnSignup determines if new members auto-subscribe.
	SubscribeOnSignup bool `json:"subscribe_on_signup,omitempty"`
}

// NewslettersResponse is the API response for newsletter listings.
type NewslettersResponse struct {
	// Newsletters is the array of newsletters.
	Newsletters []Newsletter `json:"newsletters"`
}

// Webhook represents a Ghost webhook configuration.
// Webhooks send HTTP requests when events occur in Ghost.
type Webhook struct {
	// ID is the unique identifier.
	ID string `json:"id,omitempty"`
	// Event is the trigger event (e.g., "post.published").
	Event string `json:"event,omitempty"`
	// TargetURL is where the webhook sends requests.
	TargetURL string `json:"target_url,omitempty"`
	// Name is an optional friendly name.
	Name string `json:"name,omitempty"`
	// Secret is used for request signing.
	Secret string `json:"secret,omitempty"`
	// Status indicates if the webhook is active.
	Status string `json:"status,omitempty"`
	// LastTriggeredAt is when the webhook last fired.
	LastTriggeredAt string `json:"last_triggered_at,omitempty"`
	// IntegrationID links to the parent integration.
	IntegrationID string `json:"integration_id,omitempty"`
}

// WebhooksResponse is the API response for webhook listings.
type WebhooksResponse struct {
	// Webhooks is the array of webhooks.
	Webhooks []Webhook `json:"webhooks"`
}

// Image represents an uploaded image with its URL.
type Image struct {
	// URL is the public URL of the uploaded image.
	URL string `json:"url"`
	// Ref is an optional reference identifier.
	Ref string `json:"ref,omitempty"`
}

// ImagesResponse is the API response for image uploads.
type ImagesResponse struct {
	// Images is the array of uploaded images.
	Images []Image `json:"images"`
}

// Meta contains pagination information for list responses.
type Meta struct {
	// Pagination contains the pagination details.
	Pagination Pagination `json:"pagination"`
}

// Pagination contains pagination state and navigation.
type Pagination struct {
	// Page is the current page number (1-indexed).
	Page int `json:"page"`
	// Limit is the number of items per page.
	Limit int `json:"limit"`
	// Pages is the total number of pages.
	Pages int `json:"pages"`
	// Total is the total number of items.
	Total int `json:"total"`
	// Next is the next page number, or nil if on the last page.
	Next *int `json:"next"`
	// Prev is the previous page number, or nil if on the first page.
	Prev *int `json:"prev"`
}

// APIError represents a single error from the Ghost API.
type APIError struct {
	// Message is the human-readable error description.
	Message string `json:"message"`
	// Context provides additional error details.
	Context string `json:"context"`
	// Type categorizes the error (e.g., "ValidationError").
	Type string `json:"type"`
}

// ErrorResponse is the API error response structure.
// Ghost returns errors in this format for all error status codes.
type ErrorResponse struct {
	// Errors is the array of error details.
	Errors []APIError `json:"errors"`
}
