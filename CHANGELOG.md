# [5.21.0](https://github.com/IBM/go-sdk-core/compare/v5.20.1...v5.21.0) (2025-07-25)


### Features

* enable proxy support when SSL verification is disabled ([#254](https://github.com/IBM/go-sdk-core/issues/254)) ([28c72e3](https://github.com/IBM/go-sdk-core/commit/28c72e3eb0d6b41cd427753fd5e0e9abfa2313a3))

## [5.20.1](https://github.com/IBM/go-sdk-core/compare/v5.20.0...v5.20.1) (2025-06-11)


### Bug Fixes

* **BaseService:** allow User-Agent to be set via service-level headers ([ae8e642](https://github.com/IBM/go-sdk-core/commit/ae8e6422bcd1ad5d50d152ecfc0aa2d95167f77a))

# [5.20.0](https://github.com/IBM/go-sdk-core/compare/v5.19.1...v5.20.0) (2025-05-30)


### Features

* **auth:** add support for MCSP V2 authentication ([c7328f9](https://github.com/IBM/go-sdk-core/commit/c7328f9ce855882def836c80d363bb192287f61e))

## [5.19.1](https://github.com/IBM/go-sdk-core/compare/v5.19.0...v5.19.1) (2025-04-17)


### Bug Fixes

* **build:** migrate to new yaml package ([476dc47](https://github.com/IBM/go-sdk-core/commit/476dc472e68919b9bc341a7034cfeda456e3bbd8)), closes [#247](https://github.com/IBM/go-sdk-core/issues/247)

# [5.19.0](https://github.com/IBM/go-sdk-core/compare/v5.18.6...v5.19.0) (2025-03-07)


### Features

* **ContainerAuthenticator:** add support for code engine workload ([#244](https://github.com/IBM/go-sdk-core/issues/244)) ([80518d2](https://github.com/IBM/go-sdk-core/commit/80518d281b258af9f7fd4da3555f06fa0b48bfdd))

## [5.18.6](https://github.com/IBM/go-sdk-core/compare/v5.18.5...v5.18.6) (2025-03-07)


### Bug Fixes

* **build:** bump crypto dependency to avoid CVE ([22d98e3](https://github.com/IBM/go-sdk-core/commit/22d98e30c32f14284ead14fb13b46625a2c6707f))

## [5.18.5](https://github.com/IBM/go-sdk-core/compare/v5.18.4...v5.18.5) (2025-01-07)


### Bug Fixes

* avoid runtime panic when the `errors` field of the response is `nil` ([9c104b3](https://github.com/IBM/go-sdk-core/commit/9c104b3c2d304f01bead1ba515d18e461b9c980b))

## [5.18.4](https://github.com/IBM/go-sdk-core/compare/v5.18.3...v5.18.4) (2025-01-07)


### Bug Fixes

* bump x/net to v0.34.0 to avoid CVE-2024-24338 ([#235](https://github.com/IBM/go-sdk-core/issues/235)) ([3b75e84](https://github.com/IBM/go-sdk-core/commit/3b75e8401ca1b0ae8ff600f9ea3a43220a55424b))
* fix semantic-release config ([9d9e0be](https://github.com/IBM/go-sdk-core/commit/9d9e0be08acd00ccf26ab8e8c6bbc268fc719b8d))

## [5.18.3](https://github.com/IBM/go-sdk-core/compare/v5.18.2...v5.18.3) (2024-12-13)


### Bug Fixes

* bump golang.org/x/crypto to avoid CVE-2024-45337 ([#234](https://github.com/IBM/go-sdk-core/issues/234)) ([4b6901f](https://github.com/IBM/go-sdk-core/commit/4b6901f9a3592d55a446d7264192a5108d1a8e30))

## [5.18.2](https://github.com/IBM/go-sdk-core/compare/v5.18.1...v5.18.2) (2024-12-12)


### Bug Fixes

* wrap full validation errors instead of just their message ([#232](https://github.com/IBM/go-sdk-core/issues/232)) ([1bfb146](https://github.com/IBM/go-sdk-core/commit/1bfb146dfb85a470a7f0c70854252551853f89bb))

## [5.18.1](https://github.com/IBM/go-sdk-core/compare/v5.18.0...v5.18.1) (2024-10-21)


### Bug Fixes

* fix minor problem with debug message ([#231](https://github.com/IBM/go-sdk-core/issues/231)) ([cf84c9d](https://github.com/IBM/go-sdk-core/commit/cf84c9d32b34d0d1132ab4deb27b998d182e2c58))

# [5.18.0](https://github.com/IBM/go-sdk-core/compare/v5.17.5...v5.18.0) (2024-10-15)


### Features

* **IamAssumeAuthenticator:** introduce new authenticator type ([#229](https://github.com/IBM/go-sdk-core/issues/229)) ([df85f15](https://github.com/IBM/go-sdk-core/commit/df85f1516a6c1d5c7884f2614dc70404709c4423))

## [5.17.5](https://github.com/IBM/go-sdk-core/compare/v5.17.4...v5.17.5) (2024-09-03)


### Bug Fixes

* **logging:** improve go core's debug logging ([#227](https://github.com/IBM/go-sdk-core/issues/227)) ([f0f3237](https://github.com/IBM/go-sdk-core/commit/f0f3237dd13567a394adaf5a5091c64bedec0c59))

## [5.17.4](https://github.com/IBM/go-sdk-core/compare/v5.17.3...v5.17.4) (2024-06-25)


### Bug Fixes

* tidy Go modules ([#221](https://github.com/IBM/go-sdk-core/issues/221)) ([5da6409](https://github.com/IBM/go-sdk-core/commit/5da6409a41ef0e6a57f6a1e1fb4cfced8ef6df55))
* This also includes the change to bump go-retryablehttp from v0.7.5 to v0.7.7 to avoid a vulnerability.

## [5.17.3](https://github.com/IBM/go-sdk-core/compare/v5.17.2...v5.17.3) (2024-05-16)


### Bug Fixes

* **request_builder:** encode unresolved path string prior to path parameter insertion ([#219](https://github.com/IBM/go-sdk-core/issues/219)) ([5f567da](https://github.com/IBM/go-sdk-core/commit/5f567da5e3cad27e5d3ea66d5bb5bc0b7b3cc602))

## [5.17.2](https://github.com/IBM/go-sdk-core/compare/v5.17.1...v5.17.2) (2024-05-08)


### Bug Fixes

* **errors:** flatten sdk problem chains to reduce hash complexity ([#218](https://github.com/IBM/go-sdk-core/issues/218)) ([9fc1ebc](https://github.com/IBM/go-sdk-core/commit/9fc1ebcf8c0b6068da13307adca2cfbac5fd5cea))

## [5.17.1](https://github.com/IBM/go-sdk-core/compare/v5.17.0...v5.17.1) (2024-05-06)


### Bug Fixes

* use correct error variable in external auth case ([#217](https://github.com/IBM/go-sdk-core/issues/217)) ([0b70d7d](https://github.com/IBM/go-sdk-core/commit/0b70d7d0ceae54ded6379627a180b86273880add))

# [5.17.0](https://github.com/IBM/go-sdk-core/compare/v5.16.5...v5.17.0) (2024-04-17)


### Features

* send user-agent header with auth token requests ([#216](https://github.com/IBM/go-sdk-core/issues/216)) ([90f0ba5](https://github.com/IBM/go-sdk-core/commit/90f0ba5b67107f1b40eb82646b2a3493b20891aa))

## [5.16.5](https://github.com/IBM/go-sdk-core/compare/v5.16.4...v5.16.5) (2024-04-10)


### Bug Fixes

* bump golang.org/x/net to avoid  CVE-2023-45288 ([#215](https://github.com/IBM/go-sdk-core/issues/215)) ([b9287d2](https://github.com/IBM/go-sdk-core/commit/b9287d298dd4f95148601a4f5bcac5fc5b790209))

## [5.16.4](https://github.com/IBM/go-sdk-core/compare/v5.16.3...v5.16.4) (2024-04-08)


### Bug Fixes

* **errors:** wrap errors from external authenticators to use new system ([#214](https://github.com/IBM/go-sdk-core/issues/214)) ([c74778e](https://github.com/IBM/go-sdk-core/commit/c74778eb5607c37e280afad749d803ab1d05f9a2))

## [5.16.3](https://github.com/IBM/go-sdk-core/compare/v5.16.2...v5.16.3) (2024-03-21)


### Bug Fixes

* **errors:** restore matching against all original errors ([#213](https://github.com/IBM/go-sdk-core/issues/213)) ([005fdb8](https://github.com/IBM/go-sdk-core/commit/005fdb81f15b02816ed9164dc1ebcccd08b3254d))

## [5.16.2](https://github.com/IBM/go-sdk-core/compare/v5.16.1...v5.16.2) (2024-03-20)


### Bug Fixes

* **errors:** support error checking methods for new problem structures ([#212](https://github.com/IBM/go-sdk-core/issues/212)) ([65eb01d](https://github.com/IBM/go-sdk-core/commit/65eb01df2c2030704eaa980e26b249c91069eaf5))

## [5.16.1](https://github.com/IBM/go-sdk-core/compare/v5.16.0...v5.16.1) (2024-03-14)


### Bug Fixes

* **deps:** upgrade yaml package to avoid vulnerability ([#211](https://github.com/IBM/go-sdk-core/issues/211)) ([0ee3e33](https://github.com/IBM/go-sdk-core/commit/0ee3e33415564ccd5a43b3440b6bb8e3758368fd))

# [5.16.0](https://github.com/IBM/go-sdk-core/compare/v5.15.3...v5.16.0) (2024-03-13)


### Features

* **errors:** augment errors with new "Problem" types ([#199](https://github.com/IBM/go-sdk-core/issues/199)) ([6ec86dd](https://github.com/IBM/go-sdk-core/commit/6ec86ddee70f7f7d0da8fe2ec23081e409ce53fa))

## [5.15.3](https://github.com/IBM/go-sdk-core/compare/v5.15.2...v5.15.3) (2024-02-27)


### Bug Fixes

* adjust IAM token expiration time ([#209](https://github.com/IBM/go-sdk-core/issues/209)) ([0fdd924](https://github.com/IBM/go-sdk-core/commit/0fdd924889e63775c58169ed7fb8e70a524d789c))

## [5.15.2](https://github.com/IBM/go-sdk-core/compare/v5.15.1...v5.15.2) (2024-02-23)


### Bug Fixes

* update go.sum to enable tests ([#208](https://github.com/IBM/go-sdk-core/issues/208)) ([fac6515](https://github.com/IBM/go-sdk-core/commit/fac6515179586a7facc317ad19f59aed167aa16f))

## [5.15.1](https://github.com/IBM/go-sdk-core/compare/v5.15.0...v5.15.1) (2024-01-29)


### Bug Fixes

* **deps:** bump dependencies to the latest versions ([#205](https://github.com/IBM/go-sdk-core/issues/205)) ([9f23f44](https://github.com/IBM/go-sdk-core/commit/9f23f44ef8b1bf130950b80560227c31ca4fc63d))

# [5.15.0](https://github.com/IBM/go-sdk-core/compare/v5.14.1...v5.15.0) (2023-11-15)


### Features

* **MCSPAuthenticator:** add new authenticator for Multi-Cloud Saas Platform ([#198](https://github.com/IBM/go-sdk-core/issues/198)) ([8987085](https://github.com/IBM/go-sdk-core/commit/898708506dd28fdc5992a05d99e894302b87d21b))

## [5.14.1](https://github.com/IBM/go-sdk-core/compare/v5.14.0...v5.14.1) (2023-08-22)


### Bug Fixes

* **RedactSecrets:** add additional keywords to be redacted ([#191](https://github.com/IBM/go-sdk-core/issues/191)) ([d176568](https://github.com/IBM/go-sdk-core/commit/d1765683a46a4994f91bb70e51f95396c115c4c7)), closes [#190](https://github.com/IBM/go-sdk-core/issues/190)

# [5.14.0](https://github.com/IBM/go-sdk-core/compare/v5.13.4...v5.14.0) (2023-08-18)


### Features

* bump Go versions ([#188](https://github.com/IBM/go-sdk-core/issues/188)) ([d1496b1](https://github.com/IBM/go-sdk-core/commit/d1496b17d2b6d49787079ae0cd1d94d0f78d81c4))

## [5.13.4](https://github.com/IBM/go-sdk-core/compare/v5.13.3...v5.13.4) (2023-05-22)


### Bug Fixes

* **ContainerAuthenticator:** add sa-token as default CR token filename ([#183](https://github.com/IBM/go-sdk-core/issues/183)) ([25472f3](https://github.com/IBM/go-sdk-core/commit/25472f3909b31d939cef45f21edf7d11bcbb0a9d))

## [5.13.3](https://github.com/IBM/go-sdk-core/compare/v5.13.2...v5.13.3) (2023-05-18)


### Bug Fixes

* tweak JSON parsing error message ([#185](https://github.com/IBM/go-sdk-core/issues/185)) ([7474f1b](https://github.com/IBM/go-sdk-core/commit/7474f1b73bf19b4d52fbe680fef04c3932d1fc48))

## [5.13.2](https://github.com/IBM/go-sdk-core/compare/v5.13.1...v5.13.2) (2023-05-05)


### Bug Fixes

* **deps:** bump dependencies to recent versions ([#184](https://github.com/IBM/go-sdk-core/issues/184)) ([a33f767](https://github.com/IBM/go-sdk-core/commit/a33f767ce419513721eed96f6115956424039ec0))

## [5.13.1](https://github.com/IBM/go-sdk-core/compare/v5.13.0...v5.13.1) (2023-02-28)


### Bug Fixes

* **utils:** recognize content types with whitespace ([#181](https://github.com/IBM/go-sdk-core/issues/181)) ([09fe5cf](https://github.com/IBM/go-sdk-core/commit/09fe5cfc2590b0b2cfacc95b3736e3bdb71db442))

# [5.13.0](https://github.com/IBM/go-sdk-core/compare/v5.12.2...v5.13.0) (2023-02-24)


### Bug Fixes

* fix v5-related errors ([#178](https://github.com/IBM/go-sdk-core/issues/178)) ([e8e20a3](https://github.com/IBM/go-sdk-core/commit/e8e20a35d9f86224aa19e46726aa599a04fe29e2))
* remove CONTRIBUTING.md from bump2version config ([#179](https://github.com/IBM/go-sdk-core/issues/179)) ([4d610bb](https://github.com/IBM/go-sdk-core/commit/4d610bb9655ac898384c9e7ab1ae591009f3f845))


### Features

* bump minimum go version to 1.18 ([#177](https://github.com/IBM/go-sdk-core/issues/177)) ([4648b07](https://github.com/IBM/go-sdk-core/commit/4648b07558e6bc80d7bd1c32484a7f8642708087))

## [5.12.2](https://github.com/IBM/go-sdk-core/compare/v5.12.1...v5.12.2) (2023-02-23)


### Bug Fixes

* BaseService.Request invoked without result does not close http.Response ([#176](https://github.com/IBM/go-sdk-core/issues/176)) ([a8c0324](https://github.com/IBM/go-sdk-core/commit/a8c0324701ed96381a47807fe03e6ae245d53205))

## [5.12.1](https://github.com/IBM/go-sdk-core/compare/v5.12.0...v5.12.1) (2023-02-08)


### Bug Fixes

* avoid warnings with new gosec ([a2b536c](https://github.com/IBM/go-sdk-core/commit/a2b536c19ab8e66912ad8588227c6fd6e257bdca))

# [5.12.0](https://github.com/IBM/go-sdk-core/compare/v5.11.0...v5.12.0) (2023-01-10)


### Features

* **utils:** add helper for creating a byte array pointer ([#173](https://github.com/IBM/go-sdk-core/issues/173)) ([df9dcc6](https://github.com/IBM/go-sdk-core/commit/df9dcc6094b29fda152637810f749d9f035f6cae))

# [5.11.0](https://github.com/IBM/go-sdk-core/compare/v5.10.3...v5.11.0) (2023-01-09)


### Features

* create multi-part form as a streamed request body ([#169](https://github.com/IBM/go-sdk-core/issues/169)) ([7df8c71](https://github.com/IBM/go-sdk-core/commit/7df8c71c9a91dc0086ac61deec8e98513df6e6c3))

## [5.10.3](https://github.com/IBM/go-sdk-core/compare/v5.10.2...v5.10.3) (2023-01-09)


### Bug Fixes

* pin build to semantic-release v19 ([11c0810](https://github.com/IBM/go-sdk-core/commit/11c08103974f7baf2235b0f7b186389e3054dd6d))
* use node v18 with semantic-release ([07a57ef](https://github.com/IBM/go-sdk-core/commit/07a57ef1c30a1a64ec2c61a3bfef31de66a483cb))
* **VpcInstanceAuthenticator:** use correct version string ([b4c7377](https://github.com/IBM/go-sdk-core/commit/b4c737774fea6900e7abd0c439d45681e8692f9f))

## [5.10.2](https://github.com/IBM/go-sdk-core/compare/v5.10.1...v5.10.2) (2022-08-01)


### Bug Fixes

* bump deps to avoid yaml.v3 vulnerability ([#164](https://github.com/IBM/go-sdk-core/issues/164)) ([2885864](https://github.com/IBM/go-sdk-core/commit/28858640bba89ee8d706ddb87533928481afaa5c))

## [5.10.1](https://github.com/IBM/go-sdk-core/compare/v5.10.0...v5.10.1) (2022-05-27)


### Bug Fixes

* **deps:** refresh some build dependencies ([#163](https://github.com/IBM/go-sdk-core/issues/163)) ([b66932a](https://github.com/IBM/go-sdk-core/commit/b66932a6936ee796490315433f5207d04f229554))

# [5.10.0](https://github.com/IBM/go-sdk-core/compare/v5.9.5...v5.10.0) (2022-05-09)


### Features

* add function GetQueryParamAsInt ([#162](https://github.com/IBM/go-sdk-core/issues/162)) ([2b4d018](https://github.com/IBM/go-sdk-core/commit/2b4d018c6dfd50d340958f6152ff1c17181fe8dd))

## [5.9.5](https://github.com/IBM/go-sdk-core/compare/v5.9.4...v5.9.5) (2022-03-23)


### Bug Fixes

* **IamAuthenticator:** tweak Validate() method to be more lenient ([#158](https://github.com/IBM/go-sdk-core/issues/158)) ([8f002d6](https://github.com/IBM/go-sdk-core/commit/8f002d6102a2f8d0eeed6d73eb59a2cd98ad8f65))

## [5.9.4](https://github.com/IBM/go-sdk-core/compare/v5.9.3...v5.9.4) (2022-03-22)


### Bug Fixes

* retain http.Client config when retries are enabled ([#157](https://github.com/IBM/go-sdk-core/issues/157)) ([fe093da](https://github.com/IBM/go-sdk-core/commit/fe093da7e039a0fc0cfcf5d2ae9d642323561dd4))

## [5.9.3](https://github.com/IBM/go-sdk-core/compare/v5.9.2...v5.9.3) (2022-03-16)


### Bug Fixes

* set the minimum TLS version in the client to v1.2 ([#156](https://github.com/IBM/go-sdk-core/issues/156)) ([0188990](https://github.com/IBM/go-sdk-core/commit/01889905767f6d8315e27fea539a134620806120))

## [5.9.2](https://github.com/IBM/go-sdk-core/compare/v5.9.1...v5.9.2) (2022-02-02)


### Bug Fixes

* allow retries and disable ssl to co-exist ([#154](https://github.com/IBM/go-sdk-core/issues/154)) ([b16fe8d](https://github.com/IBM/go-sdk-core/commit/b16fe8df7e6f90a794c4ecdebf6a48c7949cb2a7))

## [5.9.1](https://github.com/IBM/go-sdk-core/compare/v5.9.0...v5.9.1) (2021-12-10)


### Bug Fixes

* avoid false positive gosec errors ([#149](https://github.com/IBM/go-sdk-core/issues/149)) ([b3da5ed](https://github.com/IBM/go-sdk-core/commit/b3da5ed4d2ceaf703c77a553d28dfd9726b1a44d))

# [5.9.0](https://github.com/IBM/go-sdk-core/compare/v5.8.2...v5.9.0) (2021-11-29)


### Features

* **IamAuthenticator:** support refresh token flow in IamAuthenticator ([#146](https://github.com/IBM/go-sdk-core/issues/146)) ([97f89dd](https://github.com/IBM/go-sdk-core/commit/97f89dd9a1e8dd268993c03151bac7e8e5db00f3))

## [5.8.2](https://github.com/IBM/go-sdk-core/compare/v5.8.1...v5.8.2) (2021-11-25)


### Bug Fixes

* bump go-openapi/strfmt to avoid vulnerability alert ([#147](https://github.com/IBM/go-sdk-core/issues/147)) ([7d61715](https://github.com/IBM/go-sdk-core/commit/7d61715a7f0b3eea82ca07e3eb814a5429e3d623))

## [5.8.1](https://github.com/IBM/go-sdk-core/compare/v5.8.0...v5.8.1) (2021-11-19)


### Bug Fixes

* add .cveignore ([#144](https://github.com/IBM/go-sdk-core/issues/144)) ([e903a2f](https://github.com/IBM/go-sdk-core/commit/e903a2fdd101db700fe0e6ac96e7d1f5301a49a9))

# [5.8.0](https://github.com/IBM/go-sdk-core/compare/v5.7.2...v5.8.0) (2021-11-08)


### Features

* **VpcInstanceAuthenticator:** add support for new VPC authentication flow ([#139](https://github.com/IBM/go-sdk-core/issues/139)) ([9906ab3](https://github.com/IBM/go-sdk-core/commit/9906ab382ea206312498f636777c43205c9b1be8))

## [5.7.2](https://github.com/IBM/go-sdk-core/compare/v5.7.1...v5.7.2) (2021-10-26)


### Bug Fixes

* use consistent retry behavior for 5xx status codes ([ee5f62d](https://github.com/IBM/go-sdk-core/commit/ee5f62d58fd7da52380b3bc7c1a7155bb93b833a))

## [5.7.1](https://github.com/IBM/go-sdk-core/compare/v5.7.0...v5.7.1) (2021-10-25)


### Bug Fixes

* redact secrets when logging requests/responses ([8693f6a](https://github.com/IBM/go-sdk-core/commit/8693f6a484c4a45634d11a7b5992034a7de0612c))

# [5.7.0](https://github.com/IBM/go-sdk-core/compare/v5.6.5...v5.7.0) (2021-10-07)


### Features

* **build:** bump min go version to 1.14 ([#140](https://github.com/IBM/go-sdk-core/issues/140)) ([eb86886](https://github.com/IBM/go-sdk-core/commit/eb86886ef0385752f12f88a8aa5a09ee74afc185))

## [5.6.5](https://github.com/IBM/go-sdk-core/compare/v5.6.4...v5.6.5) (2021-09-15)


### Bug Fixes

* recognize vendor-specific JSON mimetypes ([#138](https://github.com/IBM/go-sdk-core/issues/138)) ([fb2c14a](https://github.com/IBM/go-sdk-core/commit/fb2c14a12eed98fc1d92dc8db8b746243757eb1d))

## [5.6.4](https://github.com/IBM/go-sdk-core/compare/v5.6.3...v5.6.4) (2021-08-31)


### Bug Fixes

* handle the error during gzip compression instead of panic ([#137](https://github.com/IBM/go-sdk-core/issues/137)) ([15bc45b](https://github.com/IBM/go-sdk-core/commit/15bc45b26efc113f3b32328cca32f4627f2d5141))

## [5.6.3](https://github.com/IBM/go-sdk-core/compare/v5.6.2...v5.6.3) (2021-08-13)


### Bug Fixes

* support 'AUTHTYPE' as alias for 'AUTH_TYPE' config property ([#133](https://github.com/IBM/go-sdk-core/issues/133)) ([6795484](https://github.com/IBM/go-sdk-core/commit/6795484cf8a7df70808a4342d7dba8a780ef287a))

## [5.6.2](https://github.com/IBM/go-sdk-core/compare/v5.6.1...v5.6.2) (2021-08-04)


### Bug Fixes

* refactor container authenticator with recent design changes ([#129](https://github.com/IBM/go-sdk-core/issues/129)) ([58d4475](https://github.com/IBM/go-sdk-core/commit/58d4475f394cd5bcf1d4802534780a7815a1dc77))

## [5.6.1](https://github.com/IBM/go-sdk-core/compare/v5.6.0...v5.6.1) (2021-07-27)


### Bug Fixes

* error message used by CR Authenticator ([#126](https://github.com/IBM/go-sdk-core/issues/126)) ([3632ce6](https://github.com/IBM/go-sdk-core/commit/3632ce65e98981fe02e864ca4b39430ecaf1deeb))

# [5.6.0](https://github.com/IBM/go-sdk-core/compare/v5.5.1...v5.6.0) (2021-07-26)


### Features

* add support for new ComputeResourceAuthenticator ([#123](https://github.com/IBM/go-sdk-core/issues/123)) ([c7631e3](https://github.com/IBM/go-sdk-core/commit/c7631e392f99c703aaeafee04c1dae177ab56bd2))

## [5.5.1](https://github.com/IBM/go-sdk-core/compare/v5.5.0...v5.5.1) (2021-06-22)


### Bug Fixes

* make the get token method exported ([#120](https://github.com/IBM/go-sdk-core/issues/120)) ([658327c](https://github.com/IBM/go-sdk-core/commit/658327c27eecfccda4933c18fbb76b04284f1b3e))

# [5.5.0](https://github.com/IBM/go-sdk-core/compare/v5.4.5...v5.5.0) (2021-06-02)


### Features

* add `constructServiceURL` function ([#119](https://github.com/IBM/go-sdk-core/issues/119)) ([8213faa](https://github.com/IBM/go-sdk-core/commit/8213faaed484fe22c73a9fbb13eac37e992aab46))

## [5.4.5](https://github.com/IBM/go-sdk-core/compare/v5.4.4...v5.4.5) (2021-05-28)


### Bug Fixes

* allow user to set "Host" request header ([#118](https://github.com/IBM/go-sdk-core/issues/118)) ([efd7fe3](https://github.com/IBM/go-sdk-core/commit/efd7fe36930f794aad1c2a14a1a414611afae340))

## [5.4.4](https://github.com/IBM/go-sdk-core/compare/v5.4.3...v5.4.4) (2021-05-21)


### Bug Fixes

* add check for empty string in ParseDate with tests ([#116](https://github.com/IBM/go-sdk-core/issues/116)) ([35cd647](https://github.com/IBM/go-sdk-core/commit/35cd64717c0920cddab0720abac6dc450d9b9099))

## [5.4.3](https://github.com/IBM/go-sdk-core/compare/v5.4.2...v5.4.3) (2021-05-14)


### Bug Fixes

* **build:** prevent semantic-release from committing package-lock.json ([#115](https://github.com/IBM/go-sdk-core/issues/115)) ([7fa259f](https://github.com/IBM/go-sdk-core/commit/7fa259f0da0926f043961de45243b4e40643bb12))

## [5.4.2](https://github.com/IBM/go-sdk-core/compare/v5.4.1...v5.4.2) (2021-04-29)


### Bug Fixes

* switch to a fork of the original JWT package ([#114](https://github.com/IBM/go-sdk-core/issues/114)) ([18d04ad](https://github.com/IBM/go-sdk-core/commit/18d04ad2f6e4fa32386898c39a4580eb4bca7910))

## [5.4.1](https://github.com/IBM/go-sdk-core/compare/v5.4.0...v5.4.1) (2021-04-27)


### Bug Fixes

* support expected (but empty) response body ([#111](https://github.com/IBM/go-sdk-core/issues/111)) ([2f857c2](https://github.com/IBM/go-sdk-core/commit/2f857c2c62d8df6df9d6b3b954131fc95ac73e73))

# [5.4.0](https://github.com/IBM/go-sdk-core/compare/v5.3.0...v5.4.0) (2021-04-23)


### Bug Fixes

* eliminate goroutine leak in the authenticators ([#109](https://github.com/IBM/go-sdk-core/issues/109)) ([e5d921a](https://github.com/IBM/go-sdk-core/commit/e5d921afe4d792354ce334ea1f6a35ffe7db041a))


### Features

* add FileWithMetadata type to the core ([#110](https://github.com/IBM/go-sdk-core/issues/110)) ([c1a4884](https://github.com/IBM/go-sdk-core/commit/c1a48844690488a6efce1c1ecb53a520c0ae1d9c))

# [5.3.0](https://github.com/IBM/go-sdk-core/compare/v5.2.1...v5.3.0) (2021-03-30)


### Features

* add support for unmarshalling two-dimensional slices of model instances ([#103](https://github.com/IBM/go-sdk-core/issues/103)) ([1438a2c](https://github.com/IBM/go-sdk-core/commit/1438a2c964a2d101dfaee1ca321801b4c06c9ccd))

## [5.2.1](https://github.com/IBM/go-sdk-core/compare/v5.2.0...v5.2.1) (2021-03-30)


### Bug Fixes

* avoid data race warnings ([#102](https://github.com/IBM/go-sdk-core/issues/102)) ([9e0fcc3](https://github.com/IBM/go-sdk-core/commit/9e0fcc35175f99d4fec5a12a4ccd9b8bcb7e9737))
* update go-openapi/strfmt dependency ([#104](https://github.com/IBM/go-sdk-core/issues/104)) ([018a475](https://github.com/IBM/go-sdk-core/commit/018a47562400d58525359d6f7c93d2cb26a0f313))

# [5.2.0](https://github.com/IBM/go-sdk-core/compare/v5.1.0...v5.2.0) (2021-03-11)


### Features

* add GetQueryParam method to support pagination ([e6528df](https://github.com/IBM/go-sdk-core/commit/e6528df40260e9b391dd345f77f116dfcb9f1cee))

# [5.1.0](https://github.com/IBM/go-sdk-core/compare/v5.0.3...v5.1.0) (2021-03-04)


### Features

* add UUID, date, and datetime helpers for terraform usage ([#96](https://github.com/IBM/go-sdk-core/issues/96)) ([e651369](https://github.com/IBM/go-sdk-core/commit/e6513692bd8188e3fd628bb46eb7bbddfae94428))

## [5.0.3](https://github.com/IBM/go-sdk-core/compare/v5.0.2...v5.0.3) (2021-02-25)


### Bug Fixes

* **IAM Authenticator:** canonicalize iam url & improve iam error reporting ([835ba17](https://github.com/IBM/go-sdk-core/commit/835ba17001294802d4bbb8e19612ac2f7ae39b98))

## [5.0.2](https://github.com/IBM/go-sdk-core/compare/v5.0.1...v5.0.2) (2021-02-18)


### Bug Fixes

* ensure result value is set if err is nil ([c80dc2f](https://github.com/IBM/go-sdk-core/commit/c80dc2f43afd62a38716eb3adff7ac1cd958ee0e))

## [5.0.1](https://github.com/IBM/go-sdk-core/compare/v5.0.0...v5.0.1) (2021-02-10)


### Bug Fixes

* **build:** main migration ([#93](https://github.com/IBM/go-sdk-core/issues/93)) ([903dbae](https://github.com/IBM/go-sdk-core/commit/903dbae6d596782ca78cfee56d022a65dce6ba41))
* **build:** main migration release ([#94](https://github.com/IBM/go-sdk-core/issues/94)) ([1ec22e0](https://github.com/IBM/go-sdk-core/commit/1ec22e034356f1bec55a7158133813ac460dbfba))

# [5.0.0](https://github.com/IBM/go-sdk-core/compare/v4.10.0...v5.0.0) (2021-01-20)


### Features

* add debug logging of requests/responses ([37e6597](https://github.com/IBM/go-sdk-core/commit/37e65976c10d9371794646030fb7905ba3a495f4))


### BREAKING CHANGES

* several methods added to Logger interface

Several methods were added to the Go core's Logger interface:
SetLogLevel(), GetLogLevel(), and IsLogLevelEnabled().
These additional methods will need to be added to any
user implementations of the Logger interface.
* additional parameter added to NewLogger() signature

The NewLogger() function has a new parameter "errorLogger".
Any calls to NewLogger() will need to be modified to include the
new parameter.
* deprecated unmarshal-related methods have been removed

Several deprecated unmarshal-related methods were removed from the Go core:
- UnmarshalString, UnmarshalStringSlice, UnmarshalStringMap, UnmarshalStringMapSlice
- UnmarshalByteArray, UnmarshalByteArraySlice, UnmarshalByteArrayMap, UnmarshalByteArrayMapSlice
- UnmarshalBool, UnmarshalBoolSlice, UnmarshalBoolMap, UnmarshalBoolMapSlice
- UnmarshalInt64, UnmarshalInt64Slice, UnmarshalInt64Map, UnmarshalInt64MapSlice
- UnmarshalFloat32, UnmarshalFloat32Slice, UnmarshalFloat32Map, UnmarshalFloat32MapSlice
- UnmarshalFloat64, UnmarshalFloat64Slice, UnmarshalFloat64Map, UnmarshalFloat64MapSlice
- UnmarshalUUID, UnmarshalUUIDSlice, UnmarshalUUIDMap, UnmarshalUUIDMapSlice
- UnmarshalDate, UnmarshalDateSlice, UnmarshalDateMap, UnmarshalDateMapSlice
- UnmarshalDateTime, UnmarshalDateTimeSlice, UnmarshalDateTimeMap, UnmarshalDateTimeMapSlice
- UnmarshalObject, UnmarshalObjectSlice
- UnmarshalAny, UnmarshalAnySlice, UnmarshalAnyMap, UnmarshalAnyMapSlice
These methods are no longer used by code emitted by the Go generator.  If you
have old generated Go code that still uses these methods, then you should continue
using version 4 of the Go core, or regenerate your SDK code using a new version of the
SDK generator.

# [4.10.0](https://github.com/IBM/go-sdk-core/compare/v4.9.0...v4.10.0) (2021-01-15)


### Features

* support username/apikey use-case in CloudPakForDataAuthenticator ([4e72735](https://github.com/IBM/go-sdk-core/commit/4e72735ec034d9993a22b462e2d116c984ac6cfb)), closes [arf/planning-sdk-squad#2344](https://github.com/arf/planning-sdk-squad/issues/2344)

# [4.9.0](https://github.com/IBM/go-sdk-core/compare/v4.8.2...v4.9.0) (2020-12-03)


### Features

* **BaseService:** add Clone() method to clone a BaseService instance ([45b40ee](https://github.com/IBM/go-sdk-core/commit/45b40eeeebfbbab858e079d61d6b2219d45ef75b))

## [4.8.2](https://github.com/IBM/go-sdk-core/compare/v4.8.1...v4.8.2) (2020-11-17)


### Bug Fixes

* improve serialization of DateTime values ([410fdae](https://github.com/IBM/go-sdk-core/commit/410fdaeafef209b8c0ab3c954b1c886fedcf1bca)), closes [arf/planning-sdk-squad#2313](https://github.com/arf/planning-sdk-squad/issues/2313)

## [4.8.1](https://github.com/IBM/go-sdk-core/compare/v4.8.0...v4.8.1) (2020-10-30)


### Bug Fixes

* support enable-retries via external configuration ([2f88b9f](https://github.com/IBM/go-sdk-core/commit/2f88b9f5962c7acb20ef6335522a5a0c3fec90ce))

# [4.8.0](https://github.com/IBM/go-sdk-core/compare/v4.7.1...v4.8.0) (2020-10-27)


### Features

* introduce support for automatic retries ([39bc64c](https://github.com/IBM/go-sdk-core/commit/39bc64c933fce5961382099367e0de875657e223)), closes [arf/planning-sdk-squad#2229](https://github.com/arf/planning-sdk-squad/issues/2229)

## [4.7.1](https://github.com/IBM/go-sdk-core/compare/v4.7.0...v4.7.1) (2020-10-26)


### Bug Fixes

* jwt dependency upgrade ([#81](https://github.com/IBM/go-sdk-core/issues/81)) ([ba2780c](https://github.com/IBM/go-sdk-core/commit/ba2780cf773fcfbaa5ff3bc005d53441e89bdc21))

# [4.7.0](https://github.com/IBM/go-sdk-core/compare/v4.6.1...v4.7.0) (2020-10-15)


### Features

* support use of Context with RequestBuilder ([d8e3f71](https://github.com/IBM/go-sdk-core/commit/d8e3f71f4296364478bc613e163e3c7d73c379da)), closes [arf/planning-sdk-squad#2230](https://github.com/arf/planning-sdk-squad/issues/2230) [#77](https://github.com/IBM/go-sdk-core/issues/77)

## [4.6.1](https://github.com/IBM/go-sdk-core/compare/v4.6.0...v4.6.1) (2020-10-12)


### Bug Fixes

* expose IamAuthenticator.requestToken as a public method ([c7f4cfd](https://github.com/IBM/go-sdk-core/commit/c7f4cfdbba3d94647aa9a823982f075c13112ad6))

# [4.6.0](https://github.com/IBM/go-sdk-core/compare/v4.5.1...v4.6.0) (2020-10-05)


### Features

* add support for gzip compression of request bodies ([397cbaa](https://github.com/IBM/go-sdk-core/commit/397cbaad5429b8810840fa82f8a1b187bd405c42)), closes [arf/planning-sdk-squad#2185](https://github.com/arf/planning-sdk-squad/issues/2185)

## [4.5.1](https://github.com/IBM/go-sdk-core/compare/v4.5.0...v4.5.1) (2020-09-25)


### Bug Fixes

* dont panic in SetBodyContent when nonJSONContent is nil ([#75](https://github.com/IBM/go-sdk-core/issues/75)) ([23dfbd4](https://github.com/IBM/go-sdk-core/commit/23dfbd4202069f325e62d948d0135d06fcbde0fc))

# [4.5.0](https://github.com/IBM/go-sdk-core/compare/v4.4.1...v4.5.0) (2020-09-17)


### Features

* **IAM Authenticator:** add support for optional 'scope' property ([3aa18d5](https://github.com/IBM/go-sdk-core/commit/3aa18d5fa075e9df7687da0962ee70bf44bcdad5))

## [4.4.1](https://github.com/IBM/go-sdk-core/compare/v4.4.0...v4.4.1) (2020-09-14)


### Bug Fixes

* iam/cp4d token refresh logic ([#73](https://github.com/IBM/go-sdk-core/issues/73)) ([8d4685e](https://github.com/IBM/go-sdk-core/commit/8d4685e881c3f4806f971ab98f26bba64cb7b40f))

# [4.4.0](https://github.com/IBM/go-sdk-core/compare/v4.3.1...v4.4.0) (2020-09-09)


### Features

* add new RequestBuilder.ResolveRequestURL ([5739af8](https://github.com/IBM/go-sdk-core/commit/5739af8ab0627d3a060a7d6a9636fdf25358e626)), closes [arf/planning-sdk-squad#2152](https://github.com/arf/planning-sdk-squad/issues/2152)

## [4.3.1](https://github.com/IBM/go-sdk-core/compare/v4.3.0...v4.3.1) (2020-09-08)


### Bug Fixes

* expose AuthenticationError.Err field and add ctor ([3005687](https://github.com/IBM/go-sdk-core/commit/3005687a9087000c177dcada0e8ca14ccab25971))

# [4.3.0](https://github.com/IBM/go-sdk-core/compare/v4.2.0...v4.3.0) (2020-08-19)


### Features

* add utility function to convert primitive type slices to string slices ([#68](https://github.com/IBM/go-sdk-core/issues/68)) ([136d316](https://github.com/IBM/go-sdk-core/commit/136d31608d13a88dfbfb1257611bc3367b9d4821))

# [4.2.0](https://github.com/IBM/go-sdk-core/compare/v4.1.0...v4.2.0) (2020-08-14)


### Features

* add detailed error response to iam/cp4d authenticators ([#66](https://github.com/IBM/go-sdk-core/issues/66)) ([3485263](https://github.com/IBM/go-sdk-core/commit/3485263179566e258883cc8ce55144b5b99fa308))

# [4.1.0](https://github.com/IBM/go-sdk-core/compare/v4.0.8...v4.1.0) (2020-08-07)


### Features

* rename isNil to be IsNil (public) ([1698f78](https://github.com/IBM/go-sdk-core/commit/1698f787b6e518525f2360864a672faa0b04a17f))

## [4.0.8](https://github.com/IBM/go-sdk-core/compare/v4.0.7...v4.0.8) (2020-07-29)


### Bug Fixes

* improve error paths in BaseService.Request() ([c5dd77f](https://github.com/IBM/go-sdk-core/commit/c5dd77f04ccc7440fbb25c17ed687fb6c85cb3c1))

## [4.0.7](https://github.com/IBM/go-sdk-core/compare/v4.0.6...v4.0.7) (2020-07-28)


### Bug Fixes

* use isNil() for interface{} values ([e1c27a0](https://github.com/IBM/go-sdk-core/commit/e1c27a00aecba6550c82ead865e1e1b9b5423fe6))

## [4.0.6](https://github.com/IBM/go-sdk-core/compare/v4.0.5...v4.0.6) (2020-07-14)


### Bug Fixes

* avoid linter error in DetailedResponse.String() ([ad41174](https://github.com/IBM/go-sdk-core/commit/ad4117405502843001e17310f8323af9a3568ae7))

## [4.0.5](https://github.com/IBM/go-sdk-core/compare/v4.0.4...v4.0.5) (2020-07-13)


### Bug Fixes

* correctly handle nil pointer interfaces ([2734d50](https://github.com/IBM/go-sdk-core/commit/2734d50de6e6b075359260b94df2344d9b9bc088))

## [4.0.4](https://github.com/IBM/go-sdk-core/compare/v4.0.3...v4.0.4) (2020-06-02)


### Bug Fixes

* correctly unmarshal JSON 'null' value for maps and slices ([0117461](https://github.com/IBM/go-sdk-core/commit/0117461d47c1a734c8726ecfe4ad5cbd1c971af2))

## [4.0.3](https://github.com/IBM/go-sdk-core/compare/v4.0.2...v4.0.3) (2020-05-29)


### Bug Fixes

* support applications that use 'dep' ([70e852a](https://github.com/IBM/go-sdk-core/commit/70e852a54c2acb1724c1188fe2e50cb2466888e9))

## [4.0.2](https://github.com/IBM/go-sdk-core/compare/v4.0.1...v4.0.2) (2020-05-09)


### Bug Fixes

* expose GetServiceProperties function ([b908d82](https://github.com/IBM/go-sdk-core/commit/b908d82d59301ffd06c9049c8266b3ee6900d679))

## [4.0.1](https://github.com/IBM/go-sdk-core/compare/v4.0.0...v4.0.1) (2020-05-08)


### Bug Fixes

* allow = in config property values ([13beaae](https://github.com/IBM/go-sdk-core/commit/13beaaebd10564886d87f8b7b516e1907358a776))

# [4.0.0](https://github.com/IBM/go-sdk-core/compare/v3.3.1...v4.0.0) (2020-05-04)


### Features

* **BaseService:** return non-JSON responses via 'result' and DetailedResponse.Result ([6fd7194](https://github.com/IBM/go-sdk-core/commit/6fd7194a83150f6737ef47c5e62ea6c4df4a595c))
* **BaseService:** return non-JSON responses via 'result' and DetailedResponse.Result ([e46d8c2](https://github.com/IBM/go-sdk-core/commit/e46d8c251c645846e8e61f08fd162c5cefb1d7fa))
* **unmarshal:** introduce new unmarshal functions for primitives and models ([1a033d6](https://github.com/IBM/go-sdk-core/commit/1a033d6018dfa552caa2f8be45d6b10cd34accc0))


### BREAKING CHANGES

* **BaseService:** This change to the BaseService.Request method introduces
an incompatibility with respect to the 'result' parameter.
Projects generated with the SDK generator v3.5.0 and below should
continue using version 3.x of the Go core.
Any code generated with the SDK generator version 3.6.0 or above, should use
this new version 4.0.0 of the Go core.

Note: this commit contains only a trivial change and the
* **BaseService:** message actually applies to the previous commit
with the same commit message.

## [3.3.1](https://github.com/IBM/go-sdk-core/compare/v3.3.0...v3.3.1) (2020-04-30)


### Bug Fixes

* Pace requests to token server for new auth tokens ([#55](https://github.com/IBM/go-sdk-core/issues/55)) ([578399b](https://github.com/IBM/go-sdk-core/commit/578399b1c8294de8f9e87d516264b864d711ef8e))

# [3.3.0](https://github.com/IBM/go-sdk-core/compare/v3.2.4...v3.3.0) (2020-03-29)


### Features

* add unmarshal methods for maps of primitive types ([0afd3f7](https://github.com/IBM/go-sdk-core/commit/0afd3f7cc650ca9fdf868d6a2276c940cdb52651))

## [3.2.4](https://github.com/IBM/go-sdk-core/compare/v3.2.3...v3.2.4) (2020-02-24)


### Bug Fixes

* tolerate explicit JSON null values in UnmarshalXXX() methods ([3967601](https://github.com/IBM/go-sdk-core/commit/39676013711af6cb685c8c5ec7c631e226b266df))

## [3.2.3](https://github.com/IBM/go-sdk-core/compare/v3.2.2...v3.2.3) (2020-02-19)


### Bug Fixes

* Fix token caching ([880b0be](https://github.com/IBM/go-sdk-core/commit/880b0bed51187332f26ba140d01b47e079f8df0c))

## [3.2.2](https://github.com/IBM/go-sdk-core/compare/v3.2.1...v3.2.2) (2020-02-13)


### Bug Fixes

* correct go.mod ([64ff92d](https://github.com/IBM/go-sdk-core/commit/64ff92decff6e1595f3f1f7764b5839864bcca20))

## [3.2.1](https://github.com/IBM/go-sdk-core/compare/v3.2.0...v3.2.1) (2020-02-13)


### Bug Fixes

* tolerate non-compliant error response bodies ([f0e3a13](https://github.com/IBM/go-sdk-core/commit/f0e3a1301c028df05ddd315cda687fd6295e39ab))

# [3.2.0](https://github.com/IBM/go-sdk-core/compare/v3.1.1...v3.2.0) (2020-02-07)


### Features

* add unmarshal functions for 'any' ([55c1eee](https://github.com/IBM/go-sdk-core/commit/55c1eee879932086061c9d5849b972caf5d31094))

## [3.1.1](https://github.com/IBM/go-sdk-core/compare/v3.1.0...v3.1.1) (2020-01-09)


### Bug Fixes

* ensure version # is updated in go.mod ([8fdc596](https://github.com/IBM/go-sdk-core/commit/8fdc5961b6951cc8f2769fbad241f749cc983d9c))
* fixup version #'s to 3.1.0 ([ecdafe1](https://github.com/IBM/go-sdk-core/commit/ecdafe11762d060ff08fb56ab5bd3b37ca870bbc))

# [3.1.0](https://github.com/IBM/go-sdk-core/compare/v3.0.0...v3.1.0) (2020-01-06)


### Features

* add unmarshal utility functions for primitive types ([3f7299b](https://github.com/IBM/go-sdk-core/commit/3f7299b0203f0fec5f6a6ede6bd23f63568388c3))

# [3.0.0](https://github.com/IBM/go-sdk-core/compare/v2.1.0...v3.0.0) (2019-12-09)

### Features

* created new major version 3.0.0 in v3 directory ([1595df4](https://github.com/IBM/go-sdk-core/commit/1595df486aba57dd5b965354376f5590d435ecfb))

### BREAKING CHANGES

* created new major version 3.0.0 in v3 directory

# [2.1.0](https://github.com/IBM/go-sdk-core/compare/v2.0.1...v2.1.0) (2019-12-04)


### Features

* allow JSON response body to be streamed ([d1345d7](https://github.com/IBM/go-sdk-core/commit/d1345d7d5d7dc91959eafc0d8c1ddd79a6f31450))

## [2.0.1](https://github.com/IBM/go-sdk-core/compare/v2.0.0...v2.0.1) (2019-11-21)


### Bug Fixes

* add HEAD operation constant ([#41](https://github.com/IBM/go-sdk-core/issues/41)) ([47b5dc9](https://github.com/IBM/go-sdk-core/commit/47b5dc9e46c4fa25b3e93e2b1ff15136c16e1877))

# [2.0.0](https://github.com/IBM/go-sdk-core/compare/v1.1.0...v2.0.0) (2019-11-06)


### Features

* **loadFromVCAPServices:** Service configuration factory. ([87ac493](https://github.com/IBM/go-sdk-core/commit/87ac49304e600a4bac9e52f2a0a0b529e26f0db1))


### BREAKING CHANGES

* **loadFromVCAPServices:** NewBaseService constructor changes. `displayname`, and `serviceName` removed from construction function signature, since they are no longer used.

# [1.1.0](https://github.com/IBM/go-sdk-core/compare/v1.0.1...v1.1.0) (2019-11-06)


### Features

* **BaseService:** add new method ConfigureService() to BaseService struct ([27192a7](https://github.com/IBM/go-sdk-core/commit/27192a7a796038d172af5a579a7535f91973990f))

# [1.0.1](https://github.com/IBM/go-sdk-core/compare/v1.0.0...v1.0.1) (2019-10-18)
    
### Bug Fixes
    
* fixed ConstructHTTPURL to avoid '//' when path segment is empty string ([e618205](https://github.com/IBM/go-sdk-core/commit/e61820596fbab3d475f4c2ba1d4417d755b78557))
* use go module instead of dep for dependency management; use golangci-lint for linting ([58a9cf6](https://github.com/IBM/go-sdk-core/commit/58a9cf666216ab4a420b686347f5e050e78ef975))

# [1.0.0](https://github.com/IBM/go-sdk-core/compare/v0.8.0...v1.0.0) (2019-10-04)


### Bug Fixes

* use correct error message for SSL verification errors ([a7fe39e](https://github.com/IBM/go-sdk-core/commit/a7fe39e))


### Documentation

* Updated README with information about the authentticators ([7ef7f24](https://github.com/IBM/go-sdk-core/commit/7ef7f24))


### Features

* **creds:** Search creds in working dir before home ([bf4a377](https://github.com/IBM/go-sdk-core/commit/bf4a377))
* **creds:** Search creds in working dir before home ([fbb900b](https://github.com/IBM/go-sdk-core/commit/fbb900b))
* **disable ssl:** Add description on disabling ssl verification ([298ec08](https://github.com/IBM/go-sdk-core/commit/298ec08))


### BREAKING CHANGES

* this release introduces a new authentication scheme

# [0.9.0](https://github.com/IBM/go-sdk-core/compare/v0.8.0...v0.9.0) (2019-09-24)


### Features

* **creds:** Search creds in working dir before home ([bf4a377](https://github.com/IBM/go-sdk-core/commit/bf4a377))
* **creds:** Search creds in working dir before home ([fbb900b](https://github.com/IBM/go-sdk-core/commit/fbb900b))
* **disable ssl:** Add description on disabling ssl verification ([298ec08](https://github.com/IBM/go-sdk-core/commit/298ec08))

# [0.8.0](https://github.com/IBM/go-sdk-core/compare/v0.7.0...v0.8.0) (2019-09-23)


### Features

* Defer service URL validation ([6f51c35](https://github.com/IBM/go-sdk-core/commit/6f51c35)), closes [arf/planning-sdk-squad#1011](https://github.com/arf/planning-sdk-squad/issues/1011)
* **creds:** Search creds in working dir before home ([bf4a377](https://github.com/IBM/go-sdk-core/commit/bf4a377))
* **creds:** Search creds in working dir before home ([fbb900b](https://github.com/IBM/go-sdk-core/commit/fbb900b))
* **disable ssl:** Add description on disabling ssl verification ([298ec08](https://github.com/IBM/go-sdk-core/commit/298ec08))
