# Authentication
The go-sdk-core project supports the following types of authentication:
- Basic Authentication
- Bearer Token Authentication
- Identity and Access Management (IAM) Authentication
- Container Authentication
- VPC Instance Authentication
- Cloud Pak for Data Authentication
- No Authentication

The SDK user configures the appropriate type of authentication for use with service instances.  
The authentication types that are appropriate for a particular service may vary from service to service,
so it is important for the SDK user to consult with the appropriate service documentation to understand
which authentication types are supported for that service.

The go-sdk-core allows an authenticator to be specified in one of two ways:
1. programmatically - the SDK user invokes the appropriate function(s) to create an instance of the 
desired authenticator and then passes the authenticator instance when constructing an instance of the service client.
2. configuration - the SDK user provides external configuration information (in the form of environment variables
or a credentials file) to indicate the type of authenticator, along with the configuration of the necessary properties
for that authenticator.
The SDK user then invokes the configuration-based service client constructor method
to construct an instance of the authenticator and service client that reflect the external configuration information.

The sections below will provide detailed information for each authenticator
which will include the following:
- A description of the authenticator
- The properties associated with the authenticator
- An example of how to construct the authenticator programmatically
- An example of how to configure the authenticator through the use of external
configuration information.  The configuration examples below will use
environment variables, although the same properties could be specified in a
credentials file instead.


## Basic Authentication
The `BasicAuthenticator` is used to add Basic Authentication information to
each outbound request in the `Authorization` header in the form:
```
   Authorization: Basic <encoded username and password>
```

### Properties

- Username: (required) the basic auth username

- Password: (required) the basic auth password

### Programming example
```go
import {
    "github.com/IBM/go-sdk-core/v5/core"
    "<appropriate-git-repo-url>/exampleservicev1"
}
...
// Create the authenticator.
authenticator, err := core.NewBasicAuthenticator("myuser", "mypassword")
if err != nil {
    panic(err)
}

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    Authenticator: authenticator,
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```

### Configuration example
External configuration:
```
export EXAMPLE_SERVICE_AUTH_TYPE=basic
export EXAMPLE_SERVICE_USERNAME=myuser
export EXAMPLE_SERVICE_PASSWORD=mypassword
```
Application code:
```go
import {
    "<appropriate-git-repo-url>/exampleservicev1"
}
...

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    ServiceName:   "example_service",
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1UsingExternalConfig(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```


## Bearer Token Authentication
The `BearerTokenAuthenticator` will add a user-supplied bearer token to
each outbound request in the `Authorization` header in the form:
```
    Authorization: Bearer <bearer-token>
```

### Properties

- BearerToken: (required) the bearer token value

### Programming example
```go
import {
    "github.com/IBM/go-sdk-core/v5/core"
    "<appropriate-git-repo-url>/exampleservicev1"
}
...
// Create the authenticator.
bearerToken := // ... obtain bearer token value ...
authenticator := core.NewBearerTokenAuthenticator(bearerToken)
if err != nil {
    panic(err)
}


// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    Authenticator: authenticator,
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
...
// Later, if your bearer token value expires, you can set a new one like this:
newToken := // ... obtain new bearer token value
authenticator.BearerToken = newToken
```

### Configuration example
External configuration:
```
export EXAMPLE_SERVICE_AUTH_TYPE=bearertoken
export EXAMPLE_SERVICE_BEARER_TOKEN=<the bearer token value>
```
Application code:
```go
import {
    "<appropriate-git-repo-url>/exampleservicev1"
}
...

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    ServiceName:   "example_service",
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1UsingExternalConfig(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```

Note that the use of external configuration is not as useful with the `BearerTokenAuthenticator` as it
is for other authenticator types because bearer tokens typically need to be obtained and refreshed
programmatically since they normally have a relatively short lifespan before they expire.  This
authenticator type is intended for situations in which the application will be managing the bearer 
token itself in terms of initial acquisition and refreshing as needed.


## Identity and Access Management Authentication (IAM)
The `IamAuthenticator` will accept a user-supplied apikey or refresh token and will perform
the necessary interactions with the IAM token service to obtain a suitable
bearer token for the specified apikey or refresh token.  The authenticator will also obtain 
a new bearer token when the current token expires.  The bearer token is 
then added to each outbound request in the `Authorization` header in the
form:
```
   Authorization: Bearer <bearer-token>
```

### Properties

- ApiKey: (optional) the IAM apikey to be used to obtain an IAM access token.
One of ApiKey or RefreshToken must be specified.

- RefreshToken: (optional) a refresh token to be used to obtain an IAM access token.
One of ApiKey or RefreshToken must be specified. If RefreshToken is specified, then
the ClientId and ClientSecret properties must also be specified, using the same values that were
used to obtain the refresh token value.

- URL: (optional) The base endpoint URL of the IAM token service.
The default value of this property is the "prod" IAM token service endpoint
(`https://iam.cloud.ibm.com`).
Make sure that you use an IAM token service endpoint that is appropriate for the
location of the service being used by your application.
For example, if you are using an instance of a service in the "production" environment
(e.g. `https://resource-controller.cloud.ibm.com`),
then the default "prod" IAM token service endpoint should suffice.
However, if your application is using an instance of a service in the "staging" environment
(e.g. `https://resource-controller.test.cloud.ibm.com`),
then you would also need to configure the authenticator to use the IAM token service "staging"
endpoint as well (`https://iam.test.cloud.ibm.com`).

- ClientId/ClientSecret: (optional) The `ClientId` and `ClientSecret` fields are used to form a 
"basic auth" Authorization header for interactions with the IAM token server. If neither field 
is specified, then no Authorization header will be sent with token server requests.  These fields 
are optional, but must be specified together.

- Scope: (optional) the scope to be associated with the IAM access token.
If not specified, then no scope will be associated with the access token.

- DisableSSLVerification: (optional) A flag that indicates whether verificaton of the server's SSL 
certificate should be disabled or not. The default value is `false`.

- Headers: (optional) A set of key/value pairs that will be sent as HTTP headers in requests
made to the IAM token service.

- Client: (Optional) The `http.Client` object used to invoke token servive requests. If not specified
by the user, a suitable default Client will be constructed.

### Usage Notes
- The IamAuthenticator is used to obtain an access token (a bearer token) from the IAM token service.

- When constructing an IamAuthenticator instance, you must specify exactly one of ApiKey or RefreshToken.

- If you specify the ApiKey property, the authenticator will use the 
IAM token service's `POST /identity/token` operation
with grant_type `urn:ibm:params:oauth:grant-type:apikey` to exchange the apikey value for an access token.

- If you specify the RefreshToken property, the authenticator will use the 
IAM token service's `POST /identity/token` operation
with grant_type `refresh_token` to exchange the refresh token value for an access token.
In this scenario, you must also specify the ClientId and ClientSecret properties, using the same values
that were used when initially obtaining the refresh token value from the IAM token service.

### Programming example
```go
import {
    "github.com/IBM/go-sdk-core/v5/core"
    "<appropriate-git-repo-url>/exampleservicev1"
}
...
// Create the authenticator.
authenticator, err := core.NewIamAuthenticatorBuilder().
    SetApiKey("myapikey").
    Build()
if err != nil {
    panic(err)
}

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    Authenticator: authenticator,
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```

### Configuration example
External configuration:
```
export EXAMPLE_SERVICE_AUTH_TYPE=iam
export EXAMPLE_SERVICE_APIKEY=myapikey
```
Application code:
```go
import {
    "<appropriate-git-repo-url>/exampleservicev1"
}
...

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    ServiceName:   "example_service",
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1UsingExternalConfig(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```


## Container Authentication
The `ContainerAuthenticator` is intended to be used by application code
running inside a compute resource managed by the IBM Kubernetes Service (IKS)
in which a secure compute resource token (CR token) has been stored in a file
within the compute resource's local file system.
The CR token is similar to an IAM apikey except that it is managed automatically by
the compute resource provider (IKS).
This allows the application developer to:
- avoid storing credentials in application code, configuraton files or a password vault
- avoid managing or rotating credentials

The `ContainerAuthenticator` will retrieve the CR token from
the compute resource in which the application is running, and will then perform
the necessary interactions with the IAM token service to obtain an IAM access token
using the IAM "get token" operation with grant-type `cr-token`.
The authenticator will repeat these steps to obtain a new IAM access token when the
current access token expires.
The IAM access token is added to each outbound request in the `Authorization` header in the form:
```
   Authorization: Bearer <IAM-access-token>
```

### Properties

- CRTokenFilename: (optional) the name of the file containing the injected CR token value.
If not specified, then `/var/run/secrets/tokens/vault-token` is used as the default value.
The application must have `read` permissions on the file containing the CR token value.

- IAMProfileName: (optional) the name of the linked trusted IAM profile to be used when obtaining the
IAM access token (a CR token might map to multiple IAM profiles).
One of `IAMProfileName` or `IAMProfileID` must be specified.

- IAMProfileID: (optional) the id of the linked trusted IAM profile to be used when obtaining the
IAM access token (a CR token might map to multiple IAM profiles).
One of `IAMProfileName` or `IAMProfileID` must be specified.

- URL: (optional) The base endpoint URL of the IAM token service.
The default value of this property is the "prod" IAM token service endpoint
(`https://iam.cloud.ibm.com`).
Make sure that you use an IAM token service endpoint that is appropriate for the
location of the service being used by your application.
For example, if you are using an instance of a service in the "production" environment
(e.g. `https://resource-controller.cloud.ibm.com`),
then the default "prod" IAM token service endpoint should suffice.
However, if your application is using an instance of a service in the "staging" environment
(e.g. `https://resource-controller.test.cloud.ibm.com`),
then you would also need to configure the authenticator to use the IAM token service "staging"
endpoint as well (`https://iam.test.cloud.ibm.com`).

- ClientId/ClientSecret: (optional) The `ClientId` and `ClientSecret` fields are used to form a 
"basic auth" Authorization header for interactions with the IAM token service. If neither field 
is specified, then no Authorization header will be sent with token server requests.  These fields 
are optional, but must be specified together.

- Scope: (optional) the scope to be associated with the IAM access token.
If not specified, then no scope will be associated with the access token.

- DisableSSLVerification: (optional) A flag that indicates whether verificaton of the server's SSL 
certificate should be disabled or not. The default value is `false`.

- Headers: (optional) A set of key/value pairs that will be sent as HTTP headers in requests
made to the IAM token service.

- Client: (optional) The `http.Client` object used to invoke token servive requests. If not specified
by the user, a suitable default Client will be constructed.

### Programming example
```go
import {
    "github.com/IBM/go-sdk-core/v5/core"
    "<appropriate-git-repo-url>/exampleservicev1"
}
...
// Create the authenticator.
authenticator, err := core.NewContainerAuthenticatorBuilder().
	SetIAMProfileName("iam-user123").
	Build()
if err != nil {
    panic(err)
}

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    Authenticator: authenticator,
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```

### Configuration example
External configuration:
```
export EXAMPLE_SERVICE_AUTH_TYPE=container
export EXAMPLE_SERVICE_IAM_PROFILE_NAME=iam-user123
```
Application code:
```go
import {
    "<appropriate-git-repo-url>/exampleservicev1"
}
...

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    ServiceName:   "example_service",
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1UsingExternalConfig(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```


## VPC Instance Authentication
The `VpcInstanceAuthenticator` is intended to be used by application code
running inside a VPC-managed compute resource (virtual server instance) that has been configured
to use the "compute resource identity" feature.
The compute resource identity feature allows you to assign a trusted IAM profile to the compute resource as its "identity".
This, in turn, allows applications running within the compute resource to take on this identity when interacting with
IAM-secured IBM Cloud services.
This results in a simplified security model that allows the application developer to:
- avoid storing credentials in application code, configuraton files or a password vault
- avoid managing or rotating credentials

The `VpcInstanceAuthenticator` will invoke the appropriate operations on the compute resource's locally-available
VPC Instance Metadata Service to (1) retrieve an instance identity token
and then (2) exchange that instance identity token for an IAM access token.
The authenticator will repeat these steps to obtain a new IAM access token whenever the current access token expires.
The IAM access token is added to each outbound request in the `Authorization` header in the form:
```
   Authorization: Bearer <IAM-access-token>
```

### Properties

- IAMProfileCRN: (optional) the crn of the linked trusted IAM profile to be used when obtaining the IAM access token. 

- IAMProfileID: (optional) the id of the linked trusted IAM profile to be used when obtaining the IAM access token.

- URL: (optional) The base endpoint URL of the VPC Instance Metadata Service.  
The default value of this property is `http://169.254.169.254`, and should not need to be specified in normal situations.

- Client: (optional) The `http.Client` object used to interact with the VPC Instance Metadata Service.
If not specified by the user, a suitable default Client will be constructed.

Usage Notes:
1. At most one of `IAMProfileCRN` or `IAMProfileID` may be specified.  The specified value must map
to a trusted IAM profile that has been linked to the compute resource (virtual server instance).

2. If both `IAMProfileCRN` and `IAMProfileID` are specified, then an error occurs.

3. If neither `IAMProfileCRN` nor `IAMProfileID` are specified, then the default trusted profile linked to the 
compute resource will be used to perform the IAM token exchange.
If no default trusted profile is defined for the compute resource, then an error occurs.


### Programming example
```go
import {
    "github.com/IBM/go-sdk-core/v5/core"
    "<appropriate-git-repo-url>/exampleservicev1"
}
...
// Create the authenticator.
authenticator, err := core.NewVpcInstanceAuthenticatorBuilder().
	SetIAMProfileCRN("crn:iam-profile-123").
	Build()
if err != nil {
    panic(err)
}

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    Authenticator: authenticator,
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```

### Configuration example
External configuration:
```
export EXAMPLE_SERVICE_AUTH_TYPE=vpc
export EXAMPLE_SERVICE_IAM_PROFILE_CRN=crn:iam-profile-123
```
Application code:
```go
import {
    "<appropriate-git-repo-url>/exampleservicev1"
}
...

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    ServiceName:   "example_service",
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1UsingExternalConfig(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```


##  Cloud Pak for Data Authentication
The `CloudPakForDataAuthenticator` will accept a user-supplied username value, along with either a
password or apikey, and will 
perform the necessary interactions with the Cloud Pak for Data token service to obtain a suitable
bearer token.  The authenticator will also obtain a new bearer token when the current token expires.
The bearer token is then added to each outbound request in the `Authorization` header in the
form:
```
   Authorization: Bearer <bearer-token>
```
### Properties

- Username: (required) the username used to obtain a bearer token.

- Password: (required if APIKey is not specified) the user's password used to obtain a bearer token.
Exactly one of Password or APIKey should be specified.

- APIKey: (required if Password is not specified) the user's apikey used to obtain a bearer token.
Exactly one of Password or APIKey should be specified.

- URL: (required) The URL representing the Cloud Pak for Data token service endpoint's base URL string.
This value should not include the `/v1/authorize` path portion.

- DisableSSLVerification: (optional) A flag that indicates whether verificaton of the server's SSL 
certificate should be disabled or not. The default value is `false`.

- Headers: (optional) A set of key/value pairs that will be sent as HTTP headers in requests
made to the Cloud Pak for Data token service.

- Client: (Optional) The `http.Client` object used to invoke token servive requests. If not specified
by the user, a suitable default Client will be constructed.

### Programming example
```go
import {
    "github.com/IBM/go-sdk-core/v5/core"
    "<appropriate-git-repo-url>/exampleservicev1"
}
...
// Create the authenticator using username/apikey.
authenticator, err := core.NewCloudPakForDataAuthenticatorUsingAPIKey(
    "https://mycp4dhost.com", "myuser", "myapikey", false, nil)
if err != nil {
    panic(err)
}

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    Authenticator: authenticator,
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```

### Configuration example
External configuration:
```
# Configure "example_service" with username/apikey.
export EXAMPLE_SERVICE_AUTH_TYPE=cp4d
export EXAMPLE_SERVICE_USERNAME=myuser
export EXAMPLE_SERVICE_APIKEY=myapikey
export EXAMPLE_SERVICE_URL=https://mycp4dhost.com
```
Application code:
```go
import {
    "<appropriate-git-repo-url>/exampleservicev1"
}
...

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    ServiceName:   "example_service1",
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1UsingExternalConfig(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```


## No Auth Authentication
The `NoAuthAuthenticator` is a placeholder authenticator which performs no actual authentication function.
It can be used in situations where authentication needs to be bypassed, perhaps while developing
or debugging an application or service.

### Properties
None

### Programming example
```go
import {
    "github.com/IBM/go-sdk-core/v5/core"
    "<appropriate-git-repo-url>/exampleservicev1"
}
...
// Create the authenticator.
authenticator := &core.NoAuthAuthenticator{}

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    Authenticator: authenticator,
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```

### Configuration example
External configuration:
```
export EXAMPLE_SERVICE_AUTH_TYPE=noauth
```
Application code:
```go
import {
    "<appropriate-git-repo-url>/exampleservicev1"
}
...

// Create the service options struct.
options := &exampleservicev1.ExampleServiceV1Options{
    ServiceName:   "example_service",
}

// Construct the service instance.
service, err := exampleservicev1.NewExampleServiceV1UsingExternalConfig(options)
if err != nil {
    panic(err)
}

// 'service' can now be used to invoke operations.
```
