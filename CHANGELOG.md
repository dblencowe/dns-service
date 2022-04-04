# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

### [1.1.1](https://github.com/dblencowe/dns-service/compare/v1.1.0...v1.1.1) (2022-04-04)


### Bug Fixes

* add README and ci files ([afd69d8](https://github.com/dblencowe/dns-service/commit/afd69d858e5bb32004b902519a77f0c00b46a947))

## 1.1.0 (2022-04-04)


### Features

* adds docker-compose file for client / server setup ([48ae8cb](https://github.com/dblencowe/dns-service/commit/48ae8cb0355a438b66c3d2d3874567eaf1559f6a))
* adds Dockerfile ([99c8075](https://github.com/dblencowe/dns-service/commit/99c80753806e5dd5f1e1180a4ebae13bb261d215))
* adds output formatting & verbosity controlled using OUTPUT_LEVEL ([dff98bc](https://github.com/dblencowe/dns-service/commit/dff98bcc24278c06a4e1a5fe44bad97aa4102fb4))
* fetches A records from cloudflare and storing in memory cache ([9d26940](https://github.com/dblencowe/dns-service/commit/9d269400fa7f02c1d542b2d6fa521b352ec6cc57))
* fix: set response header ([9f9a204](https://github.com/dblencowe/dns-service/commit/9f9a204255c0a8e904a01c7b03618085f5049b9f))
* handle error responses from dns forwarder ([f0df5f5](https://github.com/dblencowe/dns-service/commit/f0df5f5b96d752d39bb96285bae152644daa73fe))
* implements remaining protocol types ([05df8d9](https://github.com/dblencowe/dns-service/commit/05df8d9ac01c851112c17aac1da666801747c8b1))
* implements support for most record types ([5186d7a](https://github.com/dblencowe/dns-service/commit/5186d7a3d77575e6465337ee03b83c56522c075d))
* moves question code to go routine ([09ad02f](https://github.com/dblencowe/dns-service/commit/09ad02f82c05ac35a1ab3314f51432ebc4559fb5))
* refactor code into service ([4e8d9ad](https://github.com/dblencowe/dns-service/commit/4e8d9ad6f1544c57bea8d7a73917a19a73a71f59))
