# AWS X-Ray Receiver

<!-- status autogenerated section -->
| Status        |           |
| ------------- |-----------|
| Stability     | [beta]: traces   |
| Distributions | [contrib] |
| Issues        | [![Open issues](https://img.shields.io/github/issues-search/open-telemetry/opentelemetry-collector-contrib?query=is%3Aissue%20is%3Aopen%20label%3Areceiver%2Fawsxray%20&label=open&color=orange&logo=opentelemetry)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aopen+is%3Aissue+label%3Areceiver%2Fawsxray) [![Closed issues](https://img.shields.io/github/issues-search/open-telemetry/opentelemetry-collector-contrib?query=is%3Aissue%20is%3Aclosed%20label%3Areceiver%2Fawsxray%20&label=closed&color=blue&logo=opentelemetry)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aclosed+is%3Aissue+label%3Areceiver%2Fawsxray) |
| Code coverage | [![codecov](https://codecov.io/github/open-telemetry/opentelemetry-collector-contrib/graph/main/badge.svg?component=receiver_awsxray)](https://app.codecov.io/gh/open-telemetry/opentelemetry-collector-contrib/tree/main/?components%5B0%5D=receiver_awsxray&displayType=list) |
| [Code Owners](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/CONTRIBUTING.md#becoming-a-code-owner)    | [@wangzlei](https://www.github.com/wangzlei), [@srprash](https://www.github.com/srprash) |

[beta]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#beta
[contrib]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
<!-- end autogenerated section -->
## Overview
The AWS X-Ray receiver accepts segments (i.e. spans) in the [X-Ray Segment format](https://docs.aws.amazon.com/xray/latest/devguide/xray-api-segmentdocuments.html).
This enables the collector to receive spans emitted by the existing X-Ray SDK. [Centralized sampling](https://github.com/aws/aws-xray-daemon/blob/master/CHANGELOG.md#300-2018-08-28) is also supported via a local TCP port.

The requests sent to AWS using the [go sdk's default authentication mechanism](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials).

## Configuration

Example:

```yaml
receivers:
  awsxray:
    endpoint: 0.0.0.0:2000
    transport: udp
    proxy_server:
      endpoint: 0.0.0.0:2000
      proxy_address: ""
      tls:
        insecure: false
        server_name_override: ""
      region: ""
      role_arn: ""
      aws_endpoint: ""
      local_mode: false
```

The default configurations below are based on the [default configurations](https://github.com/aws/aws-xray-daemon/blob/master/pkg/cfg/cfg.go#L99) of the existing X-Ray Daemon.

### endpoint (Optional)
The UDP address and port on which this receiver listens for X-Ray segment documents emitted by the X-Ray SDK.

Default: `localhost:2000`

See our [security best practices doc](https://opentelemetry.io/docs/security/config-best-practices/#protect-against-denial-of-service-attacks) to understand how to set the endpoint in different environments.

### transport (Optional)
This should always be "udp" as X-Ray SDKs only send segments using UDP.

Default: `udp`

### proxy_server (Optional)
Defines configurations related to the local TCP proxy server.

### endpoint (Optional)
The TCP address and port on which this receiver listens for calls from the X-Ray SDK and relays them to the AWS X-Ray backend to get sampling rules and report sampling statistics.

Default: `0.0.0.0:2000`

See our [security best practices doc](https://opentelemetry.io/docs/security/config-best-practices/#protect-against-denial-of-service-attacks) to understand how to set the endpoint in different environments.

### proxy_address (Optional)
Defines the proxy address that the local TCP server forwards HTTP requests to AWS X-Ray backend through. If left unconfigured, requests will be sent directly.

### insecure (Optional)
Enables or disables TLS certificate verification when the local TCP server forwards HTTP requests to the AWS X-Ray backend. This sets the `InsecureSkipVerify` in the [TLSConfig](https://godoc.org/crypto/tls#Config). When setting to true, TLS is susceptible to man-in-the-middle attacks so it should be used only for testing.

Default: `false`

### server_name_override (Optional)
This sets the ``ServerName` in the [TLSConfig](https://godoc.org/crypto/tls#Config).

### region (Optional)
The AWS region the local TCP server forwards requests to. When missing, we will try to retrieve this value through environment variables or optionally ECS/EC2 metadata endpoint (depends on `local_mode` below).

### role_arn (Optional)
The IAM role used by the local TCP server when communicating with the AWS X-Ray service. If non-empty, the receiver will attempt to call STS to retrieve temporary credentials, otherwise the standard AWS credential [lookup](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) will be performed.

### aws_endpoint (Optional)
The X-Ray service endpoint which the local TCP server forwards requests to.

### local_mode (Optional)
Determines whether the ECS/EC2 instance metadata endpoint will be called to fetch the AWS region to send requests to. Set to `true` to skip metadata check.

Default: `false`
