package models

import (
	"time"
)

// ===== Request/Response DTOs =====

// CreateRequestDTO for creating a new request
type CreateRequestDTO struct {
	CollectionID     string              `json:"collection_id" binding:"required"`
	FolderID         *string             `json:"folder_id"`
	Name             string              `json:"name" binding:"required"`
	Description      string              `json:"description"`
	Method           string              `json:"method" binding:"required,oneof=GET POST PUT PATCH DELETE HEAD OPTIONS"`
	URL              string              `json:"url" binding:"required,url"`
	Headers          []KeyValuePairDTO   `json:"headers"`
	Params           []KeyValuePairDTO   `json:"params"`
	Body             string              `json:"body"`
	BodyType         string              `json:"body_type"`
	FormDataRows     []KeyValuePairDTO   `json:"form_data_rows"`
	FormEncodedRows  []KeyValuePairDTO   `json:"form_encoded_rows"`
	GraphqlQuery     string              `json:"graphql_query"`
	GraphqlVariables string              `json:"graphql_variables"`
	Auth             *AuthorizationDTO   `json:"auth"`
	Scripts          *ScriptsDTO         `json:"scripts"`
	Settings         *RequestSettingsDTO `json:"settings"`
}

// UpdateRequestDTO for updating a request
type UpdateRequestDTO struct {
	Name             *string             `json:"name"`
	Description      *string             `json:"description"`
	Method           *string             `json:"method"`
	URL              *string             `json:"url"`
	Headers          []KeyValuePairDTO   `json:"headers"`
	Params           []KeyValuePairDTO   `json:"params"`
	Body             *string             `json:"body"`
	BodyType         *string             `json:"body_type"`
	FormDataRows     []KeyValuePairDTO   `json:"form_data_rows"`
	FormEncodedRows  []KeyValuePairDTO   `json:"form_encoded_rows"`
	GraphqlQuery     *string             `json:"graphql_query"`
	GraphqlVariables *string             `json:"graphql_variables"`
	Auth             *AuthorizationDTO   `json:"auth"`
	Scripts          *ScriptsDTO         `json:"scripts"`
	Settings         *RequestSettingsDTO `json:"settings"`
}

// RequestResponseDTO returns a request
type RequestResponseDTO struct {
	ID               string              `json:"id"`
	CollectionID     string              `json:"collection_id"`
	FolderID         *string             `json:"folder_id"`
	WorkspaceID      string              `json:"workspace_id"`
	Name             string              `json:"name"`
	Description      string              `json:"description"`
	Method           string              `json:"method"`
	URL              string              `json:"url"`
	Headers          []KeyValuePairDTO   `json:"headers"`
	Params           []KeyValuePairDTO   `json:"params"`
	Body             string              `json:"body"`
	BodyType         string              `json:"body_type"`
	FormDataRows     []KeyValuePairDTO   `json:"form_data_rows"`
	FormEncodedRows  []KeyValuePairDTO   `json:"form_encoded_rows"`
	GraphqlQuery     string              `json:"graphql_query"`
	GraphqlVariables string              `json:"graphql_variables"`
	Auth             *AuthorizationDTO   `json:"auth"`
	Scripts          *ScriptsDTO         `json:"scripts"`
	Settings         *RequestSettingsDTO `json:"settings"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

// KeyValuePairDTO for headers, params, form data
type KeyValuePairDTO struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// AuthorizationDTO for different auth types
type AuthorizationDTO struct {
	Type          string           `json:"type"`
	BearerToken   string           `json:"bearer_token,omitempty"`
	BasicUsername string           `json:"basic_username,omitempty"`
	BasicPassword string           `json:"basic_password,omitempty"`
	APIKeyKey     string           `json:"api_key_key,omitempty"`
	APIKeyValue   string           `json:"api_key_value,omitempty"`
	APIKeyIn      string           `json:"api_key_in,omitempty"`
	OAuth2        *OAuth2ConfigDTO `json:"oauth2,omitempty"`
}

// OAuth2ConfigDTO for OAuth2
type OAuth2ConfigDTO struct {
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

// ScriptsDTO for pre- and post-request
type ScriptsDTO struct {
	PreRequest   string `json:"pre_request"`
	PostResponse string `json:"post_response"`
}

// RequestSettingsDTO for HTTP behavior
type RequestSettingsDTO struct {
	HTTPVersion               string   `json:"http_version"`
	StrictSSL                 bool     `json:"strict_ssl"`
	FollowRedirects           bool     `json:"follow_redirects"`
	FollowAuthorizationHeader bool     `json:"follow_authorization_header"`
	RemoveRefererOnRedirect   bool     `json:"remove_referer_on_redirect"`
	StrictHTTPParser          bool     `json:"strict_http_parser"`
	EncodeURLAutomatically    bool     `json:"encode_url_automatically"`
	DisableCookieJar          bool     `json:"disable_cookie_jar"`
	UseServerCipherSuite      bool     `json:"use_server_cipher_suite"`
	MaxRedirects              int      `json:"max_redirects"`
	DisabledTLSProtocols      []string `json:"disabled_tls_protocols"`
	CipherSuites              []string `json:"cipher_suites"`
}

// ExecuteRequestDTO for running a request
type ExecuteRequestDTO struct {
	RequestID     string            `json:"request_id" binding:"required"`
	EnvironmentID *string           `json:"environment_id"`
	Override      *UpdateRequestDTO `json:"override"`
}

// ExecuteResponseDTO returns execution result
type ExecuteResponseDTO struct {
	StatusCode int               `json:"status_code"`
	StatusText string            `json:"status_text"`
	Headers    []KeyValuePairDTO `json:"headers"`
	Body       string            `json:"body"`
	Size       int               `json:"size"`
	Time       int               `json:"time"`
	Cookies    []CookieDTO       `json:"cookies"`
	Tests      map[string]bool   `json:"tests,omitempty"`
}

// CookieDTO for response cookies
type CookieDTO struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Expires  string `json:"expires,omitempty"`
	MaxAge   int    `json:"max_age,omitempty"`
	Secure   bool   `json:"secure"`
	HttpOnly bool   `json:"http_only"`
}

// ===== Collection DTOs =====

type CreateCollectionDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type UpdateCollectionDTO struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
}

type CollectionResponseDTO struct {
	ID          string               `json:"id"`
	WorkspaceID string               `json:"workspace_id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Icon        string               `json:"icon"`
	Order       int                  `json:"order"`
	Requests    []RequestResponseDTO `json:"requests,omitempty"`
	Folders     []FolderResponseDTO  `json:"folders,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type FolderResponseDTO struct {
	ID           string               `json:"id"`
	CollectionID string               `json:"collection_id"`
	Name         string               `json:"name"`
	Order        int                  `json:"order"`
	Requests     []RequestResponseDTO `json:"requests,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

// ===== Environment DTOs =====

type CreateEnvironmentDTO struct {
	Name      string                   `json:"name" binding:"required"`
	Active    bool                     `json:"active"`
	Variables []EnvironmentVariableDTO `json:"variables"`
}

type UpdateEnvironmentDTO struct {
	Name      *string                  `json:"name"`
	Active    *bool                    `json:"active"`
	Variables []EnvironmentVariableDTO `json:"variables"`
}

type EnvironmentVariableDTO struct {
	Name   string `json:"name" binding:"required"`
	Value  string `json:"value" binding:"required"`
	Secret bool   `json:"secret"`
}

type EnvironmentResponseDTO struct {
	ID        string                   `json:"id"`
	Name      string                   `json:"name"`
	Active    bool                     `json:"active"`
	Variables []EnvironmentVariableDTO `json:"variables"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
}

// ===== Workspace DTOs =====

type CreateWorkspaceDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type UpdateWorkspaceDTO struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
}

type WorkspaceResponseDTO struct {
	ID           string                   `json:"id"`
	OwnerID      string                   `json:"owner_id"`
	Name         string                   `json:"name"`
	Description  string                   `json:"description"`
	Icon         string                   `json:"icon"`
	Members      []WorkspaceMemberDTO     `json:"members,omitempty"`
	Collections  []CollectionResponseDTO  `json:"collections,omitempty"`
	Environments []EnvironmentResponseDTO `json:"environments,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

type WorkspaceMemberDTO struct {
	UserID   string    `json:"user_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Avatar   string    `json:"avatar"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// ===== User DTOs =====

type RegisterDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type LoginDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileDTO struct {
	Name   *string `json:"name"`
	Bio    *string `json:"bio"`
	Avatar *string `json:"avatar"`
	Color  *string `json:"color"`
	Theme  *string `json:"theme"`
}

type UserResponseDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	Bio       string    `json:"bio"`
	Color     string    `json:"color"`
	Theme     string    `json:"theme"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthTokenDTO struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token,omitempty"`
	ExpiresIn    int             `json:"expires_in"`
	User         UserResponseDTO `json:"user"`
}

// ===== Flow DTOs =====

type CreateFlowDTO struct {
	Name        string        `json:"name" binding:"required"`
	Description string        `json:"description"`
	Nodes       []FlowNodeDTO `json:"nodes"`
	Edges       []FlowEdgeDTO `json:"edges"`
}

type UpdateFlowDTO struct {
	Name        *string       `json:"name"`
	Description *string       `json:"description"`
	Nodes       []FlowNodeDTO `json:"nodes"`
	Edges       []FlowEdgeDTO `json:"edges"`
}

type FlowResponseDTO struct {
	ID          string        `json:"id"`
	WorkspaceID string        `json:"workspace_id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Nodes       []FlowNodeDTO `json:"nodes"`
	Edges       []FlowEdgeDTO `json:"edges"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type FlowNodeDTO struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	RequestID *string                `json:"request_id,omitempty"`
	Position  map[string]interface{} `json:"position"`
	Data      map[string]interface{} `json:"data"`
}

type FlowEdgeDTO struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// ===== Spec DTOs =====

type CreateSpecDTO struct {
	Name    string `json:"name" binding:"required"`
	Version string `json:"version"`
	Content string `json:"content" binding:"required"`
}

type UpdateSpecDTO struct {
	Name    *string `json:"name"`
	Version *string `json:"version"`
	Content *string `json:"content"`
}

type SpecResponseDTO struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ===== Pagination =====

type PaginationDTO struct {
	Page  int `json:"page" binding:"min=1"`
	Limit int `json:"limit" binding:"min=1,max=100"`
}

type PaginatedResponseDTO struct {
	Data      interface{} `json:"data"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
	Limit     int         `json:"limit"`
	TotalPage int         `json:"total_page"`
}

// ===== Error Response =====

type ErrorResponseDTO struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
