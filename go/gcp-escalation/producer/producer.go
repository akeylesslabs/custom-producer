package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/cloudresourcemanager/v3"
	"strings"
	"time"
    "log"
    "net/http"
	"github.com/akeylesslabs/custom-producer/go/pkg/auth"
    "os"
    "regexp"
    "github.com/golang-jwt/jwt/v4"
    "errors"
)

const dryRunAccessID = "p-custom"
//const envGCPEmail = "GCP_EMAIL"

// ErrMissingSubClaim is returned when the original user doesn't have an
// "email" sub-claim in their access credentials.
var ErrMissingSubClaim = fmt.Errorf("email sub-claim is required")
var RoleNotExist = errors.New("user role does not exist")

// Producer is an implementation of Akeyless Custom Producer.
type Producer interface {
	Create(context.Context, *CreateRequest, http.Header) (*CreateResponse, error)
	Revoke(context.Context, *RevokeRequest, http.Header) (*RevokeResponse, error)
}

type producer struct{}
func New() (Producer, error) {
	p := producer{}
	return &p, nil
}

// contains searches a slice for a string and returns true if it finds it, returns false if it doesn't.
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func (p *producer) Create(ctx context.Context, rb *CreateRequest, rh http.Header) (*CreateResponse, error) {
	// dry run mode only makes sure that the producer configuration is valid,
	// not that the implementation is correct, so it's enough to return a valid
	// without doing anything
    

    // Akeyless sends the credentials in the first element of a list in the header, so this is assigning it
    // to a variable.
    akeylesscreds := rh["Akeylesscreds"][0]
    log.Printf("Create: received client info: %+v", rb.ClientInfo)

    // Here we check to see if the access ID is p-custom, which means its a dry run. If it is, we want to just
    // acknowledge and return a create response, so we let the akeyless gateway  know we're receiving their input
	if rb.ClientInfo.AccessID == dryRunAccessID {
        log.Printf("Create: received dry run, returning empty response")
		return &CreateResponse{}, nil
	}

    // Here we authenticate that the request we received was signed by the akeyless gateway's ID.
    // We set the akeyless gateway ID that it should come from as AKEYLESS_ACCESS_ID env variable
    // so we can check the headers for it.
  	if err := auth.Authenticate(ctx, akeylesscreds, os.Getenv("AKEYLESS_ACCESS_ID")); err != nil {
  		log.Fatalf("Create: authentication failed: %v\n", err)
   	}

    // this block will prase the JWT token and put all subclaims into claims.  This is because we want
    // to check the directory in the item_name to make sure it matches a proper area regular users
    // don't have access to.  If we don't do this, then anyone with permission to create a secret could
    // create their own payload, and send it to google.  Since we're authenticating to google based on
    // service account with cloud run, and there's no api key in the payload here, there would be no
    // other way to stop a user from giving themselves whatever GCP permissions they want.  With this,
    // we can use RBAC in akeyless to control who has access to this directory, and if they create
    // a custom secret in another directory, it will be rejected.
    claims := jwt.MapClaims{}
    _, err := jwt.ParseWithClaims(akeylesscreds, claims, nil)
    if err != nil {
        log.Printf("Create: %v", err)
    }
    if claims["attaches"] == nil {
        return nil, fmt.Errorf("Create: attaches claim is empty, can't verify path")
    }
    var attaches  map[string]interface{}
	json.Unmarshal([]byte(claims["attaches"].(string)), &attaches)
    item_name := attaches["item_name"]
    // checks against regex '/dynamic-secrets/cloud-user-access/gcp
    re := regexp.MustCompile(`^\/dynamic-secrets\/cloud-user-access\/gcp.*`)
	if re.MatchString(item_name.(string)) == false {
        return nil, fmt.Errorf("Create permission denied, %v does not match allowed folder", item_name)
    }

    // We need to check if there is an email subclaim, because we'll be giving access to the email
    // we receive here.
	var email string
	switch emailClaims := rb.ClientInfo.SubClaims["email"]; len(emailClaims) {
	case 0:
		return nil, ErrMissingSubClaim
	default:
		email = emailClaims[0]
	}
	log.Printf("Create: received Payload %+v for user %v", rb.Payload, email)

    // Unmarshal the payload string into the payload structure type, to use easier.
	payload := Payload{}
	err = json.Unmarshal([]byte(rb.Payload), &payload)
	if err != nil {
		return nil, fmt.Errorf("Create: could not unmarshal payload: %w", err)
	}

    // Create the cloud resource manager service needed for the google cloud api
	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("Create: cloudresourcemanager.NewService: %v", err)
	}

    // We set user: in front of the email because google cloud requires it when talking about GCP principals,
    // and then we call addBinding to add the permissions specified in the payload to the user.  We need to
    // specify the resource(resourceID), type of resource(project, folder, or organization), the principal (user),
    // and the gcp role to give
	user := "user:" + email
	err = addBinding(crmService, payload.ResourceID, payload.ResourceType, user, payload.Role)
	if err != nil {
		return nil, fmt.Errorf("Create: failed to escalate user: %w", err)
	}
    log.Printf("Create: escalated user %v to role %v on %v: %v", user, payload.Role, payload.ResourceType, payload.ResourceID)

    // Return the create response to acknowledge back to the gateway we received and processed the create request
	responseMessage := fmt.Sprintf("Create: escalated %s to %s on %s %s", user, payload.Role, payload.ResourceType, payload.ResourceID)
	return &CreateResponse{
		ID:       user,
		Response: responseMessage,
	}, nil
}

func (p *producer) Revoke(ctx context.Context, rb *RevokeRequest, rh http.Header) (*RevokeResponse, error) {

	// if the user starts with tmp thats from the dry run and just return now so we don't pass anything weird to gcp iam
	if strings.HasPrefix(rb.IDs[0], "tmp") {
        log.Printf("Revoke: received dry run, returning IDs as revoked")
		return &RevokeResponse{
			Revoked: rb.IDs,
		}, nil
	}

    if err := auth.Authenticate(ctx, rh["Akeylesscreds"][0], os.Getenv("AKEYLESS_ACCESS_ID")); err != nil {
		log.Fatalf("Revoke: authentication failed: %v\n", err)
	}

	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("Revoke: could not create cloudresourcemanager.NewService: %v", err)
	}

	log.Printf("Received Payload %+v\n", rb.Payload)
	log.Printf("Revoking User IDs %+v\n", rb.IDs)

	payload := Payload{}
	err = json.Unmarshal([]byte(rb.Payload), &payload)
	if err != nil {
		return nil, fmt.Errorf("Revoke: could not unmarshal payload: %w", err)
	}



	for _, user := range rb.IDs {
        for _ = range [3]int{} {
		    err := removeMember(crmService, payload.ResourceID, payload.ResourceType, user, payload.Role)
            if errors.Is(err, RoleNotExist) {
                log.Printf("Revoke exiting, %v", err)
                break
            }
            if err == nil {
                log.Printf("Revoke: removed access for user %v to %v: %v", user, payload.ResourceType, payload.Role)
                break
            }
            log.Printf("Revoke: could not remove access for user %v to %v: %v, retrying",  user, payload.ResourceType, payload.Role)
            time.Sleep(3 * time.Second)
        } 
    	if err != nil {
		    log.Printf("Revoke: could not revoke access for user %v to %v: %v, giving up",  user, payload.ResourceType, payload.Role)
	    }
	}

	return &RevokeResponse{
		Revoked: rb.IDs,
	}, nil

}

// addBinding adds the member to the project's IAM policy
func addBinding(crmService *cloudresourcemanager.Service, resourceID, resourceType, member, role string) error {

   
    // get the policy for the resource we're going to be adding permissions to
	policy, err := getPolicy(crmService, resourceID, resourceType)
	if err != nil {
		return fmt.Errorf("could not escalate user: %s to role %s for %s: %s \n%w\n", member, role, resourceType, resourceID, err)
	}

	// Find the policy binding for role. Only one binding can have the role.
	var binding *cloudresourcemanager.Binding
	for _, b := range policy.Bindings {
		if b.Role == role {
			binding = b
			break
		}
	}
    
    // check to make sure binding isn't nil, so we don't receive nil pointer
    // errors if the permission has already been revoked.
	if binding != nil {
		// If the binding exists, adds the member to the binding
		binding.Members = append(binding.Members, member)
	} else {
		// If the binding does not exist, adds a new binding to the policy
		binding = &cloudresourcemanager.Binding{
			Role:    role,
			Members: []string{member},
		}
		policy.Bindings = append(policy.Bindings, binding)
	}

    // set the policy from above on the resource type 
	err = setPolicy(crmService, resourceID, resourceType, policy)
	if err != nil {
		return fmt.Errorf("could not escalate user: %s to role %s for %s: %s \n%w\n", member, role, resourceType, resourceID, err)
	}
	return nil
}


// removeMember removes the member from the project's IAM policy
func removeMember(crmService *cloudresourcemanager.Service, resourceID, resourceType, member, role string) error {

    // Here we get the policy for the resource that we're going to be removing permissions from
	policy, err := getPolicy(crmService, resourceID, resourceType)
	if err != nil {
		return fmt.Errorf("could not fetch policy %w", err)
	}
    // Even if the error isn't nil, the policy could be empty if it doesn't exist,
    // so we want to check to make sure its not nil
    if policy == nil {
		return fmt.Errorf("could not fetch policy %w", err)
    }

	// Find the policy binding for role. Only one binding can have the role.
	var binding *cloudresourcemanager.Binding
	var bindingIndex int
    // we want to check to make sure that the bindings on the policy aren't
    // nil so we don't get nil pointer errors if there are none.
    if policy.Bindings == nil {
        return fmt.Errorf("could not fetch bindings")
    }

    // go through each policy, and if the role on the policy matches the role we used in the
    // the function call, then we set that binding(b) to the binding variable, and set the
    // index to bindingIndex, so we know where it was, we'll use that to shrink the slice below.
	for i, b := range policy.Bindings {
		if b.Role == role {
			binding = b
			bindingIndex = i
			break
		}
	}

    // its possible here that when it checked b.Role == role that none matched, and if so, then
    // binding will be nil, which causes a nil pointer crash, so make sure its not nil
    if binding == nil {
        return fmt.Errorf("%w", RoleNotExist)
    }

    // Provide an extra check to make sure that the binding members slice has the member we're trying
    // to remove permissions for, so we don't accidentally remove them for others.  If the user doesn't
    // exist and binding is nil, then when we shrink the array below, it will use the zero value for
    // binding and bindingIndex, and since bindindexIndex would be  0, it will always delete the first member
    // in the binding.  We don't want to remove extra permissions, so these two checks are important.
    if contains(binding.Members, member) == false {
        return fmt.Errorf("%w", RoleNotExist)
    }

	// Order doesn't matter for bindings or members, so to remove, move the last item
	// into the removed spot and shrink the slice.
	if len(binding.Members) == 1 {
		// If the member is the only member in the binding, removes the binding
		last := len(policy.Bindings) - 1
        // Here we overwrite the binding we want with the last one in the array,
        // then shrink it down, this effectively removes the element we don't want
		policy.Bindings[bindingIndex] = policy.Bindings[last]
		policy.Bindings = policy.Bindings[:last]
    } else {
    // Here we do the same copy and shrink method we used just above, but with members
    // on the binding.
        var memberIndex int
        for i, mm := range binding.Members {
            if mm == member {
                memberIndex = i
            }
        }
        last := len(policy.Bindings[bindingIndex].Members) - 1
        binding.Members[memberIndex] = binding.Members[last]
        binding.Members = binding.Members[:last]
    }

    // now we need to set the policy that we created
	err = setPolicy(crmService, resourceID, resourceType, policy)
	if err != nil {
		return fmt.Errorf("could not remove role: %w", err)
	}
	return nil
}

// getPolicy gets the project's IAM policy
func getPolicy(crmService *cloudresourcemanager.Service, resourceID, resourceType string) (*cloudresourcemanager.Policy, error) {

	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	request := new(cloudresourcemanager.GetIamPolicyRequest)

	if resourceType == "project" {
        project := "projects/" + resourceID
		policy, err := crmService.Projects.GetIamPolicy(project, request).Do()
		if err != nil {
			return nil, fmt.Errorf("Projects.GetIamPolicy: %w", err)
		}
		return policy, nil
	}

	if resourceType == "folder" {
		folder := "folders/" + resourceID
		policy, err := crmService.Folders.GetIamPolicy(folder, request).Do()
		if err != nil {
			return nil, fmt.Errorf("Folders.GetIamPolicy: %w", err)
		}
		return policy, nil
	}

	if resourceType == "organization" {
		organization := "organizations/" + resourceID
		policy, err := crmService.Organizations.GetIamPolicy(organization, request).Do()
		if err != nil {
			return nil, fmt.Errorf("Organizations.GetIamPolicy: %w", err)
		}
		return policy, nil
	}
	return nil, fmt.Errorf("resource not found")
}

// setPolicy sets the project's IAM policy
func setPolicy(crmService *cloudresourcemanager.Service, resourceID, resourceType string, policy *cloudresourcemanager.Policy) error {

	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	request := new(cloudresourcemanager.SetIamPolicyRequest)
	request.Policy = policy
	var err error

	if resourceType == "project" {
        project := "projects/" + resourceID
		_, err = crmService.Projects.SetIamPolicy(project, request).Do()
		if err != nil {
			return fmt.Errorf("Projects.SetIamPolicy: %w", err)
		}
	}

	if resourceType == "folder" {
		folder := "folders/" + resourceID
		_, err = crmService.Folders.SetIamPolicy(folder, request).Do()
		if err != nil {
			return fmt.Errorf("Folders.SetIamPolicy: %w", err)
		}
	}

	if resourceType == "organization" {
        organization := "organizations/" + resourceID
		_, err = crmService.Organizations.SetIamPolicy(organization, request).Do()
		if err != nil {
			return fmt.Errorf("Projects.SetIamPolicy: %w", err)
		}
	}

	return nil
}
