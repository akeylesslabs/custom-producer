package producer

var blankStringBytes = []byte(`""`)

// CreateRequest represents requests to /sync/create endpoint to create
// temporary credentials.
type CreateRequest struct {
	Payload    string     `json:"payload"`
	ClientInfo ClientInfo `json:"client_info"`
	Input      interface{}      `json:"input,omitempty"`
}

type Payload struct {
	Role         string `json:"role"`
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
}

// ClientInfo wraps original user information, such as Access ID or sub-claims.
type ClientInfo struct {
	AccessID  string              `json:"access_id"`
	SubClaims map[string][]string `json:"sub_claims"`
}

// CreateResponse is returned by "create" operation.
type CreateResponse struct {
	ID       string      `json:"id"`
	Response interface{} `json:"response"`
}

// RevokeRequest represents revocation requests made by Akeyless Custom
// Producer. In case of Let's Encrypt producer, revoke operation does nothing,
// but still it has to be implemented.
type RevokeRequest struct {
	Payload string   `json:"payload"`
	IDs     []string `json:"ids"`
}

// RevokeResponse is returned by revoke operation.
type RevokeResponse struct {
	Revoked []string `json:"revoked"`
	Message string   `json:"message,omitempty"`
}

type gcpUser struct {
	email string
}

func (u *gcpUser) GetEmail() string {
	return u.email
}
