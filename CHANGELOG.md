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
