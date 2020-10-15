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
