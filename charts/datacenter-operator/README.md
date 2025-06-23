# Datacenter Operator Helm Chart

This Helm chart deploys the Datacenter Operator, which provides an API for centralizing data for datacenter infrastructure.

## Prerequisites

- Kubernetes 1.16+
- Helm 3.x

## Installation

```bash
helm install datacenter-operator ./charts/datacenter-operator
```

## Configuration

The following table lists the configurable parameters for the Datacenter Operator chart and their default values.

| Parameter                  | Description                                           | Default                                  |
| -------------------------- | ----------------------------------------------------- | ---------------------------------------- |
| `crds.install`             | Whether to install the CRDs during chart installation | `true`                                   |
| `replicaCount`             | Number of replicas to deploy                          | `1`                                      |
| `image.repository`         | Image repository                                      | `ncr.sky.nhn.no/nhn/datacenter-operator` |
| `image.pullPolicy`         | Image pull policy                                     | `IfNotPresent`                           |
| `image.tag`                | Image tag                                             | Chart's appVersion                       |
| `serviceAccount.create`    | Whether to create a service account                   | `true`                                   |
| `serviceAccount.name`      | Name of the service account                           | Generated using the fullname template    |
| `rbac.create`              | Whether to create RBAC resources                      | `true`                                   |
| `config.datacenterCrdName` | Name of the datacenter CRD to manage                  | `datacenter`                             |
| `config.configMapName`     | Name of the ConfigMap to watch for datacenter config  | `datacenter-config`                      |
| `config.development`       | Enable development mode                               | `false`                                  |
| `config.region`            | Default region for datacenters                        | `Norway`                                 |
| `config.location`          | Default location for datacenters                      | `Trøndelag`                              |

## CRDs

This chart includes the following Custom Resource Definitions (CRDs):

- `datacenters.vitistack.io`: Represents a datacenter with its associated providers
- `kubernetesproviders.vitistack.io`: Represents a Kubernetes provider
- `machineproviders.vitistack.io`: Represents a machine provider

The CRDs are installed by default when the chart is installed. You can disable CRD installation by setting `crds.install=false` in your values file.

> **Note**: CRDs are not removed when a chart is uninstalled by default. To remove the CRDs, you must delete them manually:
>
> ```bash
> kubectl delete crd datacenters.vitistack.io kubernetesproviders.vitistack.io machineproviders.vitistack.io
> ```
