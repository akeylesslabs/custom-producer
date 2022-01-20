package producer

import (
	"bytes"
	"crypto"
	"encoding/json"
	"fmt"

	"github.com/go-acme/lego/v4/registration"
)

var blankStringBytes = []byte(`""`)

// CreateRequest represents requests to /sync/create endpoint to create
// temporary credentials.
type CreateRequest struct {
	Payload    string     `json:"payload"`
	ClientInfo ClientInfo `json:"client_info"`
	Input      Input      `json:"input,omitempty"`
}

// Input includes variables specific to Let's Encrypt producer. The input
// should be provided with `get-dynamic-secret-value` operation.
type Input struct {
	UseStaging bool     `json:"use_staging"`
	Domain     []string `json:"domain"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *Input) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, blankStringBytes) {
		return nil
	}

	type inputAlias Input
	var ii *inputAlias

	if err := json.Unmarshal(data, &ii); err != nil {
		return fmt.Errorf("cannot unmarshal '%s': %w", string(data), err)
	}

	*i = Input(*ii)
	return nil
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

type leUser struct {
	email        string
	registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *leUser) GetEmail() string {
	return u.email
}

func (u leUser) GetRegistration() *registration.Resource {
	return u.registration
}

func (u *leUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

type certOutput struct {
	Domain            string `json:"domain"`
	CertURL           string `json:"cert_url"`
	CertStableURL     string `json:"cert_stable_url"`
	PrivateKey        []byte `json:"private_key"`
	Certificate       []byte `json:"certificate"`
	IssuerCertificate []byte `json:"issuer_certificate"`
	CSR               []byte `json:"csr"`
}
