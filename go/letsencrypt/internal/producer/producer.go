// Package producer implements Let's Encrypt certificate automation using
// Akeyless Dynamic Secrets.
package producer

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"os"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/route53"
	"github.com/go-acme/lego/v4/registration"
)

const dryRynAccessID = "p-custom"
const envLEEmail = "LE_EMAIL"

// ErrMissingSubClaim is returned when the original user doesn't have an
// "email" sub-claim in their access credentials.
var ErrMissingSubClaim = fmt.Errorf("email sub-claim is required")

// Producer is an implementation of Akeyless Custom Producer.
type Producer interface {
	Create(*CreateRequest) (*CreateResponse, error)
	Revoke(*RevokeRequest) (*RevokeResponse, error)
}

// New creates a new Producer with the provided options.
func New(opts ...Option) (Producer, error) {
	p := &producer{}

	for _, opt := range opts {
		opt(p)
	}

	return p, nil
}

type producer struct {
	dryRunEmail  string
	dryRunDomain string
}

func (p *producer) Create(r *CreateRequest) (*CreateResponse, error) {
	var email string

	// Akeyless sends a "dry-run" request to make sure producer configuration
	// is valid, but this request doesn't include all the required information;
	// for example, this request doesn't have sub-claims, so we can't know the
	// email of the user that initiated the request. In fact, we don't even
	// know the user, so we use a special dry-run access ID and email.
	//
	// In case of Let's Encrypt, it's also important that staging environment
	// is used to prevent rate-limits in production environment.
	switch r.ClientInfo.AccessID {
	case dryRynAccessID:
		email = p.dryRunEmail
		r.Input = Input{
			UseStaging: true,
			Domain:     p.dryRunDomain,
		}
	default:
		emailClaims := r.ClientInfo.SubClaims["email"]
		if len(emailClaims) == 0 {
			if envEmail := os.Getenv(envLEEmail); envEmail != "" {
				email = envEmail
			} else {
				return nil, ErrMissingSubClaim
			}
		}
	}

	certOut, err := obtainCertificate(email, r.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a new certificate: %w", err)
	}

	return &CreateResponse{
		ID:       "",
		Response: certOut,
	}, nil
}

func (p *producer) Revoke(r *RevokeRequest) (*RevokeResponse, error) {
	// This producer doesn't allow to revoke temporary credentials.
	// Here we only return the same user ids that we received even though we do
	// nothing with them.

	return &RevokeResponse{
		Revoked: r.IDs,
	}, nil
}

// obtainCertificate requests a new certificate from Let's Encrypt and attempts
// to solve the challenge to prove our identity.
//
// Currently, only DNS challenge can be solved, and only Route53 (AWS) is
// supported. To support other cloud environments, this function and its input
// should be modified.
//
// The environment that runs this producer must be able to authenticate
// seamlessly with the cloud provider and have sufficient permissions to manage
// DNS records.
func obtainCertificate(email string, inp Input) (*certOutput, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("can't generate private key: %w", err)
	}

	user := leUser{email: email, key: privateKey}
	config := lego.NewConfig(&user)

	if inp.UseStaging {
		config.CADirURL = lego.LEDirectoryStaging
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("can't crate lets encrypt client: %w", err)
	}

	r53, err := route53.NewDNSProvider()
	if err != nil {
		return nil, fmt.Errorf("can't create a new route53 dns provider: %w", err)
	}

	if err := client.Challenge.SetDNS01Provider(r53); err != nil {
		return nil, fmt.Errorf("can't setup a new dns challenge using route53 provider: %w", err)
	}

	user.registration, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("can't obtain lets encrypt registration for %s: %w", email, err)
	}

	out, err := client.Certificate.Obtain(certificate.ObtainRequest{Domains: []string{inp.Domain}})
	if err != nil {
		return nil, fmt.Errorf("can't obtain certificates for domain %v: %w", inp.Domain, err)
	}

	// we use an internal type instead of reusing lego type since we want to
	// include every field in json while lego type skips some of them
	return &certOutput{
		Domain:            out.Domain,
		CertURL:           out.CertURL,
		CertStableURL:     out.CertStableURL,
		PrivateKey:        out.PrivateKey,
		Certificate:       out.Certificate,
		IssuerCertificate: out.IssuerCertificate,
		CSR:               out.CSR,
	}, nil
}
