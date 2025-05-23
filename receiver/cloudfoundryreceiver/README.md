# Cloud Foundry Receiver

<!-- status autogenerated section -->
| Status        |           |
| ------------- |-----------|
| Stability     | [development]: logs   |
|               | [beta]: metrics   |
| Distributions | [contrib] |
| Issues        | [![Open issues](https://img.shields.io/github/issues-search/open-telemetry/opentelemetry-collector-contrib?query=is%3Aissue%20is%3Aopen%20label%3Areceiver%2Fcloudfoundry%20&label=open&color=orange&logo=opentelemetry)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aopen+is%3Aissue+label%3Areceiver%2Fcloudfoundry) [![Closed issues](https://img.shields.io/github/issues-search/open-telemetry/opentelemetry-collector-contrib?query=is%3Aissue%20is%3Aclosed%20label%3Areceiver%2Fcloudfoundry%20&label=closed&color=blue&logo=opentelemetry)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aclosed+is%3Aissue+label%3Areceiver%2Fcloudfoundry) |
| Code coverage | [![codecov](https://codecov.io/github/open-telemetry/opentelemetry-collector-contrib/graph/main/badge.svg?component=receiver_cloudfoundry)](https://app.codecov.io/gh/open-telemetry/opentelemetry-collector-contrib/tree/main/?components%5B0%5D=receiver_cloudfoundry&displayType=list) |
| [Code Owners](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/CONTRIBUTING.md#becoming-a-code-owner)    | [@crobert-1](https://www.github.com/crobert-1) \| Seeking more code owners! |
| Emeritus      | [@agoallikmaa](https://www.github.com/agoallikmaa), [@pellared](https://www.github.com/pellared) |

[development]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#development
[beta]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#beta
[contrib]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
<!-- end autogenerated section -->

The Cloud Foundry receiver connects to the RLP (Reverse Log Proxy) Gateway of the Cloud Foundry installation, typically
available at the URL `https://log-stream.<cf-system-domain>`.

## Authentication

RLP Gateway authentication is performed by adding the Oauth2 token as the `Authorization` header. To obtain an OAuth2
token to use for the RLP Gateway, a request is made to the UAA component which acts as the OAuth2 provider (URL
specified by `uaa_url` configuration option, which typically is `https://uaa.<cf-system-domain>`). To authenticate with
UAA, username and password/secret combination is used (`uaa_username` and `uaa_password` configuration options). This
UAA user must have the `client_credentials` and `refresh_token` authorized grant types, and `logs.admin` authority.

The following is an example sequence of commands to create the UAA user using the `uaac` command line utility:

* `uaac target https://uaa.<cf-system-domain>`
* `uaac token client get identity -s <identity-user-secret>`
* `uaac client add <uaa_username> --name opentelemetry --secret <uaa_password> --authorized_grant_types client_credentials,refresh_token --authorities logs.admin`

The `<uaa_username>` and `<uaa_password>` above can be set to anything as long as they match the values provided to the
receiver configuration. The admin account (which is `identity` here) which has the permissions to create new clients may
have a different name on different setups. The value of `--name` is not used for receiver configuration.

## Configuration

The receiver takes the following configuration options:

| Field | Default | Description |
| --- | --- | --- |
| `rlp_gateway.endpoint` | required | URL of the RLP gateway, typically `https://log-stream.<cf-system-domain>` |
| `rlp_gateway.tls.insecure_skip_verify` | `false` | whether to skip TLS verify for the RLP gateway endpoint |
| `rlp_gateway.shard_id` | `opentelemetry` | metrics or logs are load balanced among receivers that use the same shard ID, therefore this must only be set if there are multiple receivers which must both receive all the metrics instead of them being balanced between them. This string will be a prefix used to build a different ShardID for each envelope type; for logs the final ShardID will have the `_logs` suffix, for metrics will be `_metrics` |
| `uaa.endpoint` | required | URL of the UAA provider, typically `https://uaa.<cf-system-domain>` |
| `uaa.tls.insecure_skip_verify` | `false` | whether to skip TLS verify for the UAA endpoint |
| `uaa.username` | required | name of the UAA user (required grant types/authorities described above) |
| `uaa.password` | required | password of the UAA user |

The `rlp_gateway` configuration section also inherits configuration options from the global from:

- [HTTP Client Configuration](https://github.com/open-telemetry/opentelemetry-collector/tree/main/config/confighttp#client-configuration)

Example:

```yaml
receivers:
  cloudfoundry:
    rlp_gateway:
      endpoint: "https://log-stream.sys.example.internal"
      tls:
        insecure_skip_verify: false
      shard_id: "opentelemetry"
    uaa:
      endpoint: "https://uaa.sys.example.internal"
      tls:
        insecure_skip_verify: false
      username: "otelclient"
      password: "changeit"
```

The full list of settings exposed for this receiver are documented in [config.go](./config.go)
with detailed sample configurations in [testdata/config.yaml](./testdata/config.yaml).

## Telemetry common Attributes

The receiver maps the envelope attribute tags to the following OpenTelemetry attributes:

* `origin` - origin name as documented by Cloud Foundry

For Cloud Foundry/Tanzu Application Service deployed in BOSH, the following attributes are also present, using their
canonical BOSH meanings:

* `deployment` - BOSH deployment name
* `index` - BOSH instance ID (GUID)
* `ip` - BOSH instance IP
* `job` - BOSH job name

On TAS/PCF versions 2.8.0+ and cf-deployment versions v11.1.0+, the following additional attributes are present for application metrics: `app_id`, `app_name`, `space_id`, `space_name`, `organization_id`, `organization_name` which provide the GUID and name of application, space and organization respectively.

This might not be a comprehensive list of attributes, as the receiver passes on whatever attributes the gateway
provides, which may include some that are specific to TAS and possibly new ones in future Cloud Foundry versions as
well.

## Metrics

Reported metrics are grouped under an instrumentation library named `otelcol/cloudfoundry`. Metric names are as
specified by [Cloud Foundry metrics documentation](https://docs.cloudfoundry.org/running/all_metrics.html), but the
origin name is prepended to the metric name with `.` separator. All metrics either of type `Gauge` or `Sum`.

### Attributes

The receiver maps the envelope attribute to the following OpenTelemetry attributes:
 
* `source_id` - for applications, the GUID of the application, otherwise equal to `origin`

For metrics originating with `rep` origin name (specific to applications), the following attributes are present:

* `instance_id` - numerical index of the application instance. However, also present for `bbs` origin, where it matches the value of `index`
* `process_id` - process ID (GUID). For a process of type "web" which is the main process of an application, this is equal to `source_id` and `app_id`
* `process_instance_id` - unique ID of a process instance, should be treated as an opaque string
* `process_type` - process type. Each application has exactly one process of type `web`, but many have any number of
  other processes


## Logs

The receiver maps loggregator envelopes of these types to the following OpenTelemetry log severity text and severity number:
* type `OUT` becomes `info` and severity number `9`
* type `ERR` becomes `error` and severity number `17`
* If any other log types are received, they're discarded and result in an error log message in the collector.

### Attributes

The receiver maps the envelope attribute tags to the following OpenTelemetry attributes:

* `source_id` - for applications, the GUID of the application, otherwise the GUID of the log generator
* `source_type` - The source of the log, any subset of `{API|APP|CELL|HEALTH|LGR|RTR|SSH|STG}`, for `APP` type extra labels are separated by a dash, example: `APP/PROC/WEB`
* `instance_id` - numerical index of the origin. If origin is `rep` (`source_type` is `APP`) this is the application index. However, for other cases this is the instance index. 
* `process_id` - process ID (GUID)
* `process_instance_id` - unique ID of a process instance, should be treated as an opaque string
* `process_type` - process type. Each application has exactly one process of type `web`, but many have any number of other processes

## Feature Gate

### `cloudfoundry.resourceAttributes.allow`

#### Purpose

The `cloudfoundry.resourceAttributes.allow` [feature gate](https://github.com/open-telemetry/opentelemetry-collector/blob/main/featuregate/README.md#collector-feature-gates) allows the envelope tags being copied to the metrics as resource attributes instead of datapoint attributes (default `false`).
Therefore all `org.cloudfoundry.*` datapoint attributes won't be present anymore on metrics datapoint level, but on resource level instead, since the attributes describe the resource and not the datapoints itself.

The `cloudfoundry.resourceAttributes.allow` feature gate is available since version `v0.117.0` and will be held at least for 2 versions (`v0.119.0`) until promoting to `beta` and another 2 versions (`v0.121.0`) until promoting to `stable`.

Below you can see the list of attributes that are present the resource level instead of datapoint level (when `cloudfoundry.resourceAttributes.allow` feature gate is enabled):

```
  - org.cloudfoundry.index
  - org.cloudfoundry.ip
  - org.cloudfoundry.deployment
  - org.cloudfoundry.id
  - org.cloudfoundry.job
  - org.cloudfoundry.product
  - org.cloudfoundry.instance_group
  - org.cloudfoundry.instance_id
  - org.cloudfoundry.origin
  - org.cloudfoundry.system_domain
  - org.cloudfoundry.source_id
  - org.cloudfoundry.source_type
  - org.cloudfoundry.process_type
  - org.cloudfoundry.process_id
  - org.cloudfoundry.process_instance_id
```
