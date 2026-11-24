# Kubernetes Deployment

This directory contains a Deployment and Service for running Akeyless Custom Server on Kubernetes. The manifest uses the Alpine image by default.

## Prerequisites

- A Kubernetes cluster with a `linux/amd64` node for the published Alpine image.
- `kubectl` configured for the target cluster and an existing namespace for the deployment.
- Your Akeyless Gateway Admin Access ID.
- A `custom_logic.sh` file with your producer or rotator logic. See the [custom logic examples](../custom_logic/). When adapting an example for Alpine, use POSIX `sh` syntax, including function declarations such as `run_rotate() { ... }`.

## Available images

| Image | Base OS | Use |
| --- | --- | --- |
| `akeyless/custom-server:alpine-latest` | Alpine 3.23 | Default in `deployment.yaml` |
| `akeyless/custom-server:latest` | Ubuntu 24.04 | Set this explicitly in `deployment.yaml` to use Ubuntu |

## 1. Create the ConfigMap

Run the following command from the directory containing your `custom_logic.sh` file. Replace `your-namespace` with the namespace where you will deploy Custom Server.

```bash
kubectl -n your-namespace create configmap custom-server-logic \
  --from-file=custom_logic.sh=./custom_logic.sh
```

## 2. Edit the manifest

Open `deployment.yaml` and replace these placeholders.

| Placeholder | Value |
| --- | --- |
| `<INSERT NAMESPACE>` | The namespace from step 1. Replace it in both the Deployment and the Service. |
| `<INSERT ACCESS-ID>` | Your Akeyless Gateway Admin Access ID. The Deployment passes this value to the container as `GW_ACCESS_ID`. |

## 3. Deploy Custom Server

Run the following commands from the directory containing `deployment.yaml`. Replace `your-namespace` with the namespace from step 1.

```bash
kubectl apply -f deployment.yaml
kubectl -n your-namespace rollout status deployment/custom-server --timeout=120s
```

## 4. Register the webhooks in Akeyless

If the Akeyless Gateway runs in the same cluster, use the Service DNS name as the webhook base URL. Replace `your-namespace` with the Custom Server namespace. If your cluster uses a different DNS suffix, replace `cluster.local` with that suffix.

```text
http://custom-server.your-namespace.svc.cluster.local:2608
```

The Gateway must be able to resolve the Service name and reach port `2608`. If the Gateway runs outside the cluster and cannot reach the ClusterIP Service, use a private HTTPS endpoint backed by an internal Ingress or load balancer that forwards to `custom-server` on port `2608`. Confirm that the Gateway can resolve and reach that endpoint before registering the webhooks.

Append the appropriate path to the base URL the Gateway can reach.

| Akeyless configuration | Webhook path |
| --- | --- |
| Custom Rotated Secret web target | `/sync/rotate` |
| Custom Dynamic Secret create webhook | `/sync/create` |
| Custom Dynamic Secret revoke webhook | `/sync/revoke` |
| Optional Custom Dynamic Secret webhook for rotating the power user's credentials | `/sync/rotate` |

For example, with the in-cluster Service, the Custom Rotated Secret web target would be `http://custom-server.your-namespace.svc.cluster.local:2608/sync/rotate`.

To enable automatic rotation of the power user's credentials for a Custom Dynamic Secret, set the rotation period in days and configure the rotate webhook.

## Update the custom logic

Run the following commands from the directory containing the updated `custom_logic.sh`. Replace `your-namespace` with the namespace from step 1.

```bash
(
  set -o pipefail
  kubectl -n your-namespace create configmap custom-server-logic \
    --from-file=custom_logic.sh=./custom_logic.sh \
    --dry-run=client -o yaml | kubectl -n your-namespace apply -f -
) && kubectl -n your-namespace rollout restart deployment/custom-server \
  && kubectl -n your-namespace rollout status deployment/custom-server --timeout=120s
```

The Deployment mounts `custom_logic.sh` at `/custom_logic.sh` as a read-only file with mode `0755`. Because the mount uses `subPath`, existing pods need to be replaced to pick up changes to the file.
