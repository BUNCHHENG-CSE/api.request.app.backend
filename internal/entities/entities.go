package entities

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// User represents a user account
type User struct {
	// 24-byte fields (Slices, Times)
	CreatedAt  time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
	Workspaces []*Workspace `gorm:"many2many:workspace_members" json:"workspaces,omitempty"`

	// 16-byte fields (Strings)
	ID       string `gorm:"primaryKey" json:"id"`
	Email    string `gorm:"uniqueIndex" json:"email"`
	Password string `json:"-"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
	Color    string `json:"color"`
	Theme    string `json:"theme"` // "light" or "dark"

	// 8-byte fields (Pointers)
	Profile *Profile `gorm:"constraint:OnDelete:CASCADE;" json:"profile,omitempty"`
}

// Profile stores additional user settings
type Profile struct {
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	ID               string `gorm:"primaryKey" json:"id"`
	UserID           string `gorm:"uniqueIndex" json:"user_id"`
	Theme            string `json:"theme"`
	DefaultWorkspace string `json:"default_workspace"`

	// 1-byte fields (Bools)
	NotificationsOn bool `json:"notifications_on"`
}

// Workspace represents a user's workspace
type Workspace struct {
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	Members      []*User        `gorm:"many2many:workspace_members" json:"members,omitempty"`
	Collections  []*Collection  `gorm:"constraint:OnDelete:CASCADE;" json:"collections,omitempty"`
	Environments []*Environment `gorm:"constraint:OnDelete:CASCADE;" json:"environments,omitempty"`
	Flows        []*Flow        `gorm:"constraint:OnDelete:CASCADE;" json:"flows,omitempty"`
	Specs        []*Spec        `gorm:"constraint:OnDelete:CASCADE;" json:"specs,omitempty"`

	ID          string `gorm:"primaryKey" json:"id"`
	OwnerID     string `json:"owner_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`

	Owner *User `json:"owner,omitempty"`
}

// Collection groups related requests
type Collection struct {
	CreatedAt time.Time           `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time           `gorm:"autoUpdateTime" json:"updated_at"`
	Requests  []*RequestTab       `gorm:"constraint:OnDelete:CASCADE;" json:"requests,omitempty"`
	Folders   []*CollectionFolder `gorm:"constraint:OnDelete:CASCADE;" json:"folders,omitempty"`

	ID          string `gorm:"primaryKey" json:"id"`
	WorkspaceID string `gorm:"index" json:"workspace_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`

	Workspace *Workspace `json:"workspace,omitempty"`
	Order     int        `json:"order"`
}

// CollectionFolder for organizing requests within a collection
type CollectionFolder struct {
	CreatedAt time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	Requests  []*RequestTab `gorm:"constraint:OnDelete:SET NULL;" json:"requests,omitempty"`

	ID           string `gorm:"primaryKey" json:"id"`
	CollectionID string `gorm:"index" json:"collection_id"`
	Name         string `json:"name"`

	Collection *Collection `json:"collection,omitempty"`
	Order      int         `json:"order"`
}

// RequestTab represents a saved API request
type RequestTab struct {
	CreatedAt       time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
	Headers         []byte            `gorm:"type:jsonb" json:"headers"`
	Params          []byte            `gorm:"type:jsonb" json:"params"`
	FormDataRows    []byte            `gorm:"type:jsonb" json:"form_data_rows"`
	FormEncodedRows []byte            `gorm:"type:jsonb" json:"form_encoded_rows"`
	Responses       []*Response       `gorm:"constraint:OnDelete:CASCADE;" json:"responses,omitempty"`
	History         []*RequestHistory `gorm:"constraint:OnDelete:CASCADE;" json:"history,omitempty"`

	ID               string `gorm:"primaryKey" json:"id"`
	CollectionID     string `gorm:"index" json:"collection_id"`
	WorkspaceID      string `gorm:"index" json:"workspace_id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Method           string `json:"method"`
	URL              string `json:"url"`
	Body             string `json:"body"`
	BodyType         string `json:"body_type"`
	GraphqlQuery     string `json:"graphql_query"`
	GraphqlVariables string `json:"graphql_variables"`

	FolderID   *string            `json:"folder_id,omitempty"`
	Auth       *AuthorizationData `gorm:"type:jsonb;serializer:json" json:"auth"`
	Scripts    *Scripts           `gorm:"type:jsonb;serializer:json" json:"scripts"`
	Settings   *RequestSettings   `gorm:"type:jsonb;serializer:json" json:"settings"`
	Collection *Collection        `json:"collection,omitempty"`
	Folder     *CollectionFolder  `json:"folder,omitempty"`
	Workspace  *Workspace         `json:"workspace,omitempty"`
}

// AuthorizationData for different auth types
type AuthorizationData struct {
	Type          string `json:"type"`
	BearerToken   string `json:"bearer_token,omitempty"`
	BasicUsername string `json:"basic_username,omitempty"`
	BasicPassword string `json:"basic_password,omitempty"`
	APIKeyKey     string `json:"api_key_key,omitempty"`
	APIKeyValue   string `json:"api_key_value,omitempty"`
	APIKeyIn      string `json:"api_key_in,omitempty"`

	OAuth2 *OAuth2Config `json:"oauth2,omitempty"`
}

// OAuth2Config for OAuth2 authentication
type OAuth2Config struct {
	GrantType    string `json:"grant_type"`
	AuthURL      string `json:"auth_url"`
	TokenURL     string `json:"token_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope"`
	RedirectURI  string `json:"redirect_uri"`
	State        string `json:"state,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
}

// Scripts for pre and post-request execution
type Scripts struct {
	PreRequest   string `json:"pre_request"`
	PostResponse string `json:"post_response"`
}

// RequestSettings for HTTP behavior
type RequestSettings struct {
	DisabledTLSProtocols []string `json:"disabled_tls_protocols"`
	CipherSuites         []string `json:"cipher_suites"`

	HTTPVersion string `json:"http_version"`

	MaxRedirects int `json:"max_redirects"`

	StrictSSL                 bool `json:"strict_ssl"`
	FollowRedirects           bool `json:"follow_redirects"`
	FollowAuthorizationHeader bool `json:"follow_authorization_header"`
	RemoveRefererOnRedirect   bool `json:"remove_referer_on_redirect"`
	StrictHTTPParser          bool `json:"strict_http_parser"`
	EncodeURLAutomatically    bool `json:"encode_url_automatically"`
	DisableCookieJar          bool `json:"disable_cookie_jar"`
	UseServerCipherSuite      bool `json:"use_server_cipher_suite"`
}

// KeyValuePair for headers, params, form data
type KeyValuePair struct {
	Key         string `json:"key"`
	Val         string `json:"value"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// Scan implements sql.Scanner for JSONB (Optimized to handle NULLs)
func (kvp *KeyValuePair) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: expected byte slice")
	}
	return json.Unmarshal(bytes, &kvp)
}

// Value implements driver.Valuer for JSONB
func (kvp KeyValuePair) Value() (driver.Value, error) {
	return json.Marshal(kvp)
}

// Response stores API response data
type Response struct {
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	Headers   []byte    `gorm:"type:jsonb" json:"headers"`

	ID         string `gorm:"primaryKey" json:"id"`
	RequestID  string `gorm:"index" json:"request_id"`
	StatusText string `json:"status_text"`
	Body       string `json:"body"`

	StatusCode int `json:"status_code"`
	Size       int `json:"size"`
	Time       int `json:"time"`

	Request *RequestTab `json:"request,omitempty"`
}

// RequestHistory tracks request executions
type RequestHistory struct {
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`

	ID        string `gorm:"primaryKey" json:"id"`
	RequestID string `gorm:"index" json:"request_id"`
	UserID    string `json:"user_id"`
	Method    string `json:"method"`
	URL       string `json:"url"`

	Status   int `json:"status"`
	Duration int `json:"duration"`

	Request *RequestTab `json:"request,omitempty"`
	User    *User       `json:"user,omitempty"`
}

// Environment stores environment variables
type Environment struct {
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Variables []byte    `gorm:"type:jsonb" json:"variables"`

	ID          string `gorm:"primaryKey" json:"id"`
	WorkspaceID string `gorm:"index" json:"workspace_id"`
	Name        string `json:"name"`

	Active bool `json:"active"`

	Workspace *Workspace `json:"workspace,omitempty"`
}

// EnvironmentVariable represents a single env var
type EnvironmentVariable struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Secret bool   `json:"secret"`
}

// Flow represents a visual request orchestration workflow
type Flow struct {
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Nodes     []byte    `gorm:"type:jsonb" json:"nodes"`
	Edges     []byte    `gorm:"type:jsonb" json:"edges"`

	ID          string `gorm:"primaryKey" json:"id"`
	WorkspaceID string `gorm:"index" json:"workspace_id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Workspace *Workspace `json:"workspace,omitempty"`
}

// FlowNode represents a node in a flow
type FlowNode struct {
	Position map[string]interface{} `json:"position"`
	Data     map[string]interface{} `json:"data"`

	ID        string  `json:"id"`
	Type      string  `json:"type"`
	RequestID *string `json:"request_id,omitempty"`
}

// FlowEdge represents a connection between nodes
type FlowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// Spec stores OpenAPI specifications
type Spec struct {
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	ID          string `gorm:"primaryKey" json:"id"`
	WorkspaceID string `gorm:"index" json:"workspace_id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Content     string `gorm:"type:text" json:"content"`

	Workspace *Workspace `json:"workspace,omitempty"`
}

// WorkspaceMember tracks membership and roles
type WorkspaceMember struct {
	JoinedAt  time.Time `gorm:"autoCreateTime" json:"joined_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	WorkspaceID string `gorm:"primaryKey" json:"workspace_id"`
	UserID      string `gorm:"primaryKey" json:"user_id"`
	Role        string `json:"role"`
}
