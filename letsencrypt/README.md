# Akeyless Let's Encrypt producer

This is a custom producer implementation that automates Let's Encrypt
certificates via `get-dynamic-secret-value` operations.

## Installation

### Permissions

This producer must be deployed in AWS environment with sufficient permissions
to access and modify Route 53 records. These permissions are required in order
to solve DNS challenges presented by Let's Encrypt to prove that the caller has
control over the requested domain name.

### Using Docker image

Akeyless Let's Encrypt producer is available as a Docker image:
`akeyless/letsencrypt-producer`. It exposes a single port `:80` and is
stateless.

### Building from source

Clone this repository and build the binary using `letsencrypt/bin/cmd` package.
Running the binary creates a web-server listening on port `:80`.

## Configuration

This producer must be configured using the following environment variables:

### Dry-run mode

While setting up an integration with this producer, Akeyless performs a
"dry-run" session to make sure everything is configured properly. The following
environment variables must be set to support this operation:

#### `LE_DRY_RUN_EMAIL`

This email is used when generating a new certificate. Normally the email is
taken from an "email" sub-claim of a requesting user, but in case of dry-run,
no email is available, and a pre-defined email must be used instead.

#### `LE_DRY_RUN_DOMAIN`

A test certificate is issued by Let's Encrypt for this domain name. Normally,
the requesting user would specify the domain name to issue a certificate for.

### Production mode

#### `AKEYLESS_ACCESS_ID`

Each Akeyless API Gateway instance is associated with an Auth Method. Producers
running in the same API Gateway use this Auth Method to communicate with
Akeyless. In order to prove that the requests received by this producer were
issued by an authorized producer, its access ID must be specified at deployment
time. Access credentials issued by another access ID must not be accepted by
this producer to prevent abuse.

#### `AKEYLESS_ITEM_NAME`

This is an optional variable that allows to specify the name of dynamic secret
producer that is allowed to issue requests to this custom producer. For
example, if a single API Gateway manages more than a single custom producer, a
particular Let's Encrypt producer deployment may be limited to only one of
them. This name should be the full item name, including the leading `/`.
