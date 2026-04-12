# Custom Server

A Docker image that runs your custom logic as an Akeyless [Custom Producer](https://docs.akeyless.io/docs/custom-producer) or [Custom Rotated Secret](https://docs.akeyless.io/docs/create-a-custom-rotated-secret).

You provide a `custom_logic.sh` file that implements one or more of `run_create`, `run_revoke`, and `run_rotate`. The server handles authentication, JWT validation, and dispatching.

## Available Images

| Tag | Base OS |
|-----|---------|
| `akeyless/custom-server:latest` | Ubuntu 24.04 |
| `akeyless/custom-server:alpine-latest` | Alpine 3.23 (smaller, non-root) |

> **Alpine users:** `custom_logic.sh` is sourced by POSIX `sh`. Use `run_rotate()` syntax, not `function run_rotate()`.

## Custom Logic Examples

See [`custom_logic/`](./custom_logic/) for ready-to-use examples:

| Example | Type |
|---------|------|
| [`random_rotated_password`](./custom_logic/random_rotated_password/) | Rotated Secret |
| [`okta_rotated_password`](./custom_logic/okta_rotated_password/) | Rotated Secret |
| [`heroku_push_rotated_creds`](./custom_logic/heroku_push_rotated_creds/) | Rotated Secret |
| [`keycloak_rotate_admin_password`](./custom_logic/keycloak_rotate_admin_password/) | Rotated Secret |
| [`splunk_temp_user`](./custom_logic/splunk_temp_user/) | Dynamic Producer |

## Deployment

### Docker

```bash
# Set your Gateway's Admin Access ID
export GW_ACCESS_ID="p-xxxx"

bash docker/run_custom_server.sh
```

See [`docker/`](./docker/).

### Kubernetes

See [`kubernetes/`](./kubernetes/) for the Deployment + Service manifest and step-by-step instructions.
