# Kubernetes Deployment

This directory contains manifests for deploying the `akeyless/custom-server` image to Kubernetes.

## Prerequisites

- A running Kubernetes cluster
- `kubectl` configured against your target cluster
- A `custom_logic.sh` file with your producer/rotator logic (see [`../custom_logic/`](../custom_logic/))

## Available Images

| Tag | Base OS | Use case |
|-----|---------|----------|
| `akeyless/custom-server:latest` | Ubuntu 24.04 | Default |
| `akeyless/custom-server:alpine-latest` | Alpine 3.23 | Smaller footprint, security-hardened |

> **Note for Alpine users:** Your `custom_logic.sh` must use POSIX sh syntax.
> Avoid the bash-only `function name()` syntax — use `name()` instead.
> See the [example files](../custom_logic/) for reference.

## Steps

### 1. Create the ConfigMap from your logic file

```bash
kubectl create configmap custom-server-logic \
  --from-file=custom_logic.sh=./custom_logic.sh \
  -n <YOUR NAMESPACE>
```

To update the logic after initial deployment:
```bash
kubectl create configmap custom-server-logic \
  --from-file=custom_logic.sh=./custom_logic.sh \
  -n <YOUR NAMESPACE> \
  --dry-run=client -o yaml | kubectl apply -f -
```

> **Note:** Kubernetes does not automatically restart pods when a ConfigMap changes.
> After updating the ConfigMap, restart the deployment to pick up the new logic:
> ```bash
> kubectl rollout restart deployment/custom-server -n <YOUR NAMESPACE>
> ```

### 2. Edit the manifest

Open `deployment.yaml` and replace:

| Placeholder | Value |
|-------------|-------|
| `<INSERT NAMESPACE>` | Your Kubernetes namespace |
| `<INSERT IMAGE>` | e.g. `akeyless/custom-server:alpine-latest` |
| `<INSERT ACCESS-ID>` | Your Akeyless Gateway Admin Access ID |

### 3. Apply

```bash
kubectl apply -f deployment.yaml
```

### 4. Register the custom producer in Akeyless

Point the webhook(s) at the Service's in-cluster DNS, base URL:

```
http://custom-server.<NAMESPACE>.svc.cluster.local:2608
```

**For a Custom Rotated Secret**, the web-target should be:

```
http://custom-server.<NAMESPACE>.svc.cluster.local:2608/sync/rotate
```

**For a Custom Dynamic Secret**, set:

| Webhook | URL |
|---------|-----|
| Create | `http://custom-server.<NAMESPACE>.svc.cluster.local:2608/sync/create` |
| Revoke | `http://custom-server.<NAMESPACE>.svc.cluster.local:2608/sync/revoke` |

To also rotate the power user's credentials, enable rotation and set:

- **Rotation period (days)** — number of days between auto-rotations
- **Rotate Webhook URL** — `http://custom-server.<NAMESPACE>.svc.cluster.local:2608/sync/rotate`

## Manifest overview

```
Deployment (1 replica)
  └── custom-server container
        ├── port 2608
        ├── env: GW_ACCESS_ID
        └── volume: custom-logic (ConfigMap → /custom_logic.sh, mode 0755, readOnly)

Service (ClusterIP)
  └── port 2608 → targetPort 2608
```

The `custom_logic.sh` is mounted read-only from a ConfigMap with execute permissions (`defaultMode: 0755`), so it can be invoked by the server without being writable at runtime.
