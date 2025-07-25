[![Build Status](https://github.com/IBM/go-sdk-core/actions/workflows/build.yaml/badge.svg)](https://github.com/IBM/go-sdk-core/actions/workflows/build.yaml)
[![Release](https://img.shields.io/github/v/release/IBM/go-sdk-core)](https://github.com/IBM/go-sdk-core/releases/latest)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/IBM/go-sdk-core?filename=go.mod)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)
[![CLA assistant](https://cla-assistant.io/readme/badge/ibm/go-sdk-core)](https://cla-assistant.io/ibm/go-sdk-core)


# IBM Go SDK Core Version 5.21.0
This project contains core functionality required by Go code generated by the IBM Cloud OpenAPI SDK Generator
(openapi-sdkgen).

## Installation

Get SDK package:
```bash
go get -u github.com/IBM/go-sdk-core/...
```

## Prerequisites
- Go version 1.23 or newer

## Authentication
The go-sdk-core project supports the following types of authentication:
- Basic Authentication
- Bearer Token Authentication
- Identity and Access Management (IAM) Authentication (grant type: apikey)
- Identity and Access Management (IAM) Authentication (grant type: assume)
- Container Authentication
- VPC Instance Authentication
- Cloud Pak for Data Authentication
- Multi-Cloud Saas Platform (MCSP) Authentication
- No Authentication (for testing)

For more information about the various authentication types and how to use them with your services, click [here](Authentication.md).

## Logging
The go-sdk-core project implements a basic logging facility to log various messages.
The logger supports these logging levels: Error, Info, Warn, and Debug.

By default, the project will use a logger with log level "Error" configured, which means that
only error messages will be displayed.  A logger configured at log level "Warn" would display "Error" and "Warn" messages
(but not "Info" or "Debug"), etc.

To configure the logger to display "Info", "Warn" and "Error" messages, use the `core.SetLoggingLevel()`
method, as in this example:

```go
import (
    "github.com/IBM/go-sdk-core/v5/core"
)

// Enable Info logging.
core.SetLoggingLevel(core.LevelInfo)
```

If you configure the logger for log level "Debug", then HTTP request/response messages will be logged as well.
Here is an example that shows this, along with the steps needed to enable automatic retries:

```go
// Enable Debug logging.
core.SetLoggingLevel(core.LevelDebug)

// Construct the service client.
myService, err := exampleservicev1.NewExampleServiceV1(options)

// Enable automatic retries.
myService.EnableRetries(3, 20 * time.Second)

// Create the resource.
result, detailedResponse, err := myService.CreateResource(createResourceOptionsModel)
```

When the "CreateResource()" method is invoked, you should see a handful of debug messages
displayed on the console reporting on progress of the request, including any retries that
were performed.  Here is an example:

```
2020/10/29 10:34:57 [DEBUG] POST http://example-service.cloud.ibm.com/api/v1/resource
2020/10/29 10:34:57 [DEBUG] POST http://example-service.cloud.ibm.com/api/v1/resource (status: 429): retrying in 1s (5 left)
2020/10/29 10:34:58 [DEBUG] POST http://example-service.cloud.ibm.com/api/v1/resource (status: 429): retrying in 1s (4 left)
```

In addition to providing a basic logger implementation, the Go core library also defines
the `Logger` interface and allows users to supply their own implementation to support unique
logging requirements (perhaps you need messages logged to a file instead of the console).
To use this advanced feature, simply implement the `Logger` interface and then call the
`SetLogger(Logger)` function to set your implementation as the logger to be used by the
Go core library.

## Issues

If you encounter an issue with this project, you are welcome to submit a [bug report](https://github.com/IBM/go-sdk-core/issues).
Before opening a new issue, please search for similar issues. It's possible that someone has already reported it.

## Tests

To build, test and lint-check the project:
```bash
make all
```

Get code coverage for each test suite:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md).

## License

This library is licensed under Apache 2.0. Full license text is
available in [LICENSE](LICENSE).
