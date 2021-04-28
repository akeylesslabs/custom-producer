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
"dry-run" session to make sure everything is configured properly. This mode
doesn't do anything with Let's Encrypt. Instead, it returns an empty "Success"
response immediately.

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

### `LE_EMAIL`

This is an optional variable to set a default email address to use to access
Let's Encrypt. It will be used if no "email" sub-claim exists in the request.

## Usage

This producer accepts the following arguments:

| Field name | Description |
|-|-|
| `domain` | Required: Issue Let's Encrypt certificate for this domain. |
| `use_staging` | Use Let's Encrypt staging environment. Useful during testing or integration to avoid rate limits. |

For example:

```
akeyless get-dynamic-secret-value \
    --name /letsencrypt \
    --profile saml-profile \
    --args='{"domain":"foo.example.com","use_staging":true}' \
    --timeout 300
```

> Please make sure you don't exceed Let's Encrypt [rate
> limits](https://letsencrypt.org/docs/rate-limits/).
