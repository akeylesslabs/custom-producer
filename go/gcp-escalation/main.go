// GCP Escalation Custom Producer takes a payload of:
// {"role": "${ROLE}", "resource_type": "${RESOURCE_TYPE}", "resource_id": "${RESOURCE_ID}"
// Example:
// {"role": "roles/resourcemanager.folderAdmin","resource_type": "folder", "resource_id": "581675094958"}
// {"role": "organizations/54829595577/roles/folderAdmin","resource_type": "organization", "resource_id": "54829595577"}
// {"role": "roles/owner","resource_type": "organization", "resource_id": "54829595577"}
// and escalates the user to the proper GCP role on the resource.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gcp-privileges/producer"
	"net/http"
    "log"
)

// main routes http requests to the proper functions (Create or Revoke)
func main() {
	http.HandleFunc("/sync/create", create)
	http.HandleFunc("/sync/revoke", revoke)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

// create receives a responseWriter to write back to the akeyless server
// with, and a pointer to an http request that contains the akeyless credentials
func create(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

// Create a new producer, defined in producer/types.go
	p, err := producer.New()
	if err != nil {
		log.Fatalf("could not set up  producer %v\n", err)
	}

// Create a new CreateRequest, defined in producer/types.go which will
// contain unmarshaled json response, and contains the payload, client
// info and payload, and any user input (not used)
	var cr producer.CreateRequest
    
// This block of code defines an error type, and then unmarshals the
// http request body in the create request type.  It unmarshals the
// payload as a string, because of the way akeyless sends the user defined
// payload with escaped quotes, we have to unmarshal it first into a string,
// and then later on in the producer, we'll unmarshal that string back into
// a payload type.
	var unmarshalErr *json.UnmarshalTypeError
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&cr)
	if err != nil {
		if errors.As(err, &unmarshalErr) {
			log.Fatalf("could not unmarshal request: %v\n", err)
		}
		log.Fatal(err)
	}

// Call the create method on the producer, sending it the context, the unmarshalled
// create request, and the header, because that has the credentials we need to authenticate
	response, err := p.Create(ctx, &cr, r.Header)
	if err != nil {
		log.Fatalln(err)
	}
// Marshal the response into JSON, and send it back to the writer (w) with Fprintf, so we
// can acknowledge that we received the create and created the necessary keys, or in this
// case, escalated the user properly
	jsonResp, err := json.Marshal(*response)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprintf(w, "%v", string(jsonResp))
}

func revoke(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	p, err := producer.New()
	if err != nil {
		log.Fatalf("could not set up  producer %v\n", err)
	}

	var rr producer.RevokeRequest
	var unmarshalErr *json.UnmarshalTypeError
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err = dec.Decode(&rr)

	if err != nil {
		if errors.As(err, &unmarshalErr) {
			log.Fatalf("could not unmarshal request: %v\n", err)
		}
		log.Fatal(err)
	}

	response, err := p.Revoke(ctx, &rr, r.Header)
	if err != nil {
		log.Fatalln(err)
	}
	jsonResp, err := json.Marshal(*response)
	if err != nil {
		log.Fatalln(err)
	}
    fmt.Fprintf(w, "%v", string(jsonResp))
}
