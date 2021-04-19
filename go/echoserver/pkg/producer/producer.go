// Package producer is a dummy implementation of Akeyless Custom Producer
// compatible producer implementation. It essentially is an echo server, that
// returns request data as response value.
package producer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

var blankStringBytes = []byte(`""`)

// Producer is a simple custom producer implementation that can be deployed
// anywhere and used for tests. It doesn't include authentication!
type Producer struct{}

// Create sends back the incoming request as a "Response", and uses current
// timestamp (nano-second resolution) as an ID.
func (p *Producer) Create(r *CreateRequest) (*CreateResponse, error) {
	return &CreateResponse{
		ID:       fmt.Sprintf("%d", time.Now().UnixNano()),
		Response: r,
	}, nil
}

// Revoke sends back all the received IDs.
func (p *Producer) Revoke(r *RevokeRequest) (*RevokeResponse, error) {
	return &RevokeResponse{
		Revoked: r.IDs,
	}, nil
}

// CreateRequest represents requests to /sync/create endpoint to create
// temporary credentials.
type CreateRequest struct {
	Payload    string     `json:"payload"`
	ClientInfo ClientInfo `json:"client_info"`
	Input      Input      `json:"input,omitempty"`
}

// Input is the input that this producer accepts. It can be any JSON object.
type Input map[string]interface{}

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
// Producer.
type RevokeRequest struct {
	Payload string   `json:"payload"`
	IDs     []string `json:"ids"`
}

// RevokeResponse is returned by revoke operation.
type RevokeResponse struct {
	Revoked []string `json:"revoked"`
	Message string   `json:"message,omitempty"`
}
