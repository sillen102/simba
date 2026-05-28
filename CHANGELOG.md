## [0.30.0-dev.1](https://github.com/sillen102/simba/compare/v0.29.5...v0.30.0-dev.1) (2026-05-28)

### Features

* abandon websockets with Centrifugal, for now ([1aca8b9](https://github.com/sillen102/simba/commit/1aca8b96d7dc15a972ddf42ed236ef462dacfc56))
* add linked-versions plugin for synchronized releases ([95e03da](https://github.com/sillen102/simba/commit/95e03da757e6182089ce27f4adcaaf500f71d0fc))
* add linked-versions plugin for synchronized releases ([6aae45f](https://github.com/sillen102/simba/commit/6aae45f417b8414331930ee451380d04c3e8ff3f))
* allow pointes in query parameters ([8ed9142](https://github.com/sillen102/simba/commit/8ed9142dca9e91531550223a86a900c7c2cf1a43))
* allow shutdown hooks to be any function ([295cb36](https://github.com/sillen102/simba/commit/295cb3662e342fd6180b159600d7df31b94c51f7))
* change library for websockets from gobwas/ws to coder/websocket ([788cf73](https://github.com/sillen102/simba/commit/788cf73e319bfa9b9367ede93a7ead3edcf5d4d1))
* create dev release action ([8ed44f0](https://github.com/sillen102/simba/commit/8ed44f0d4f114a7ae0e41f492a5975b753276884))
* fix release version pattern ([9345cb4](https://github.com/sillen102/simba/commit/9345cb4299b6d7d042894b05492103b2c34b29f5))
* fix release version pattern ([ee58dfd](https://github.com/sillen102/simba/commit/ee58dfd974a9fac61b976c4b408c31db64e7ab19))
* fix release version pattern ([f934c24](https://github.com/sillen102/simba/commit/f934c2499cd95e96a4a99e072d2acd3d3f2d5b78))
* use error messages from go-playground validator in validation errors instead of maintaining own translations ([b89e7ba](https://github.com/sillen102/simba/commit/b89e7ba1d8c3ebc18553411e4400f098bc980837))
* user json tags in error messages on validation ([ab44aaa](https://github.com/sillen102/simba/commit/ab44aaa31838b9d16c1dd843f0a30deebc56c9dd))

### Bug Fixes

* add modules version sync to release script ([d63b5f5](https://github.com/sillen102/simba/commit/d63b5f53e77368bc7444f9137e8fcf275aefb79e))
* error message when no json tags are present on struct being validated ([5d04a61](https://github.com/sillen102/simba/commit/5d04a61ed1f96c7124bcbe3eb97d14236ae55965))
* release workflow improvements and fixes ([d020448](https://github.com/sillen102/simba/commit/d020448d7e30b47eba0cdc8328cd82fdf96eb3ec))
* remove mutexes since the library writer is already thread safe ([545370d](https://github.com/sillen102/simba/commit/545370d6a5de86ecb2866c873b4ff9f4972d9673))
* set includeComponentInTag to true for linked-versions plugin ([ff23ff3](https://github.com/sillen102/simba/commit/ff23ff31069623f0fcc70ef0aa5843899ee121c7))
* set includeComponentInTag to true for linked-versions plugin ([7a32bac](https://github.com/sillen102/simba/commit/7a32bacb50f2dbe2a867cc5830375765b2538ad3))
* sync module versions ([9849173](https://github.com/sillen102/simba/commit/9849173955c2f5ff5d4b431282f13160524aa2c5))
* update dependencies, use mise instead of Makefile, remove centrifuge (for now) ([1152983](https://github.com/sillen102/simba/commit/115298377c4008ac793fdf6f33c25ddf5a10a529))
* update group PR title pattern to include placeholders ([ffa091b](https://github.com/sillen102/simba/commit/ffa091b822a642579f3eedfde15e9c9fc547cd17))
* use release please ([ff3aa17](https://github.com/sillen102/simba/commit/ff3aa171d099a56c9b82c140ae9cb7406d841b57))

### Reverts

* Revert "feat: fix release version pattern" ([ff3f3c8](https://github.com/sillen102/simba/commit/ff3f3c85d925e853e235fddc65815b4f10adaa41))

## [0.30.0-dev.7](https://github.com/sillen102/simba/compare/v0.30.0-dev.6...v0.30.0-dev.7) (2026-03-31)

### Features

* allow pointes in query parameters ([7d275a0](https://github.com/sillen102/simba/commit/7d275a0daae58aeb688bde61842c083e86e5c3a4))

## [0.30.0-dev.6](https://github.com/sillen102/simba/compare/v0.30.0-dev.5...v0.30.0-dev.6) (2026-03-17)

### Features

* use error messages from go-playground validator in validation errors instead of maintaining own translations ([beee7de](https://github.com/sillen102/simba/commit/beee7de32e25969a09b16bf5ad666ce95cae2487))

## [0.30.0-dev.5](https://github.com/sillen102/simba/compare/v0.30.0-dev.4...v0.30.0-dev.5) (2026-03-17)

### Bug Fixes

* error message when no json tags are present on struct being validated ([efdf1ff](https://github.com/sillen102/simba/commit/efdf1ffae273572b0ccafe10c6723af1ec7a1528))

## [0.30.0-dev.4](https://github.com/sillen102/simba/compare/v0.30.0-dev.3...v0.30.0-dev.4) (2026-03-17)

### Features

* user json tags in error messages on validation ([564fe48](https://github.com/sillen102/simba/commit/564fe4870f2bea19d72b60c8ca066e5f4aff0986))

## [0.30.0-dev.3](https://github.com/sillen102/simba/compare/v0.30.0-dev.2...v0.30.0-dev.3) (2026-03-08)

### Features

* allow shutdown hooks to be any function ([14afff9](https://github.com/sillen102/simba/commit/14afff9f9015245c3494bddc627ac3691d8c5a4a))

## [0.30.0-dev.2](https://github.com/sillen102/simba/compare/v0.30.0-dev.1...v0.30.0-dev.2) (2026-03-05)

### Bug Fixes

* update dependencies, use mise instead of Makefile, remove centrifuge (for now) ([775e935](https://github.com/sillen102/simba/commit/775e935811087573906fe27e8ea7c693214941ad))

## [0.30.0-dev.1](https://github.com/sillen102/simba/compare/v0.29.5...v0.30.0-dev.1) (2026-03-05)

### Features

* abandon websockets with Centrifugal, for now ([187e7d4](https://github.com/sillen102/simba/commit/187e7d40145d351745dbe50256276e3768816e9b))
* add linked-versions plugin for synchronized releases ([d067756](https://github.com/sillen102/simba/commit/d0677561d0f1164209cda7fac8947f45e3b284d7))
* add linked-versions plugin for synchronized releases ([9a71b25](https://github.com/sillen102/simba/commit/9a71b25390b9340b29dc764ca3af0948f758eaa2))
* change library for websockets from gobwas/ws to coder/websocket ([cfe4643](https://github.com/sillen102/simba/commit/cfe4643413b978070981bcb5c0215e5576d35482))
* create dev release action ([98325d6](https://github.com/sillen102/simba/commit/98325d68c611cfa845f064aeff07996c0a59cf10))
* fix release version pattern ([d575ce6](https://github.com/sillen102/simba/commit/d575ce6b3e1e14ac97cb576e2b3e0983ba82dead))
* fix release version pattern ([0e983c1](https://github.com/sillen102/simba/commit/0e983c114637f015054c478e39a90ead73410323))

### Bug Fixes

* add modules version sync to release script ([a3c1cbf](https://github.com/sillen102/simba/commit/a3c1cbf64afdd6ab2e42ae5b2c96b961067b3630))
* release workflow improvements and fixes ([f47915c](https://github.com/sillen102/simba/commit/f47915ce9f15f3b480399d5673bec5bd361755e4))
* remove mutexes since the library writer is already thread safe ([058fae2](https://github.com/sillen102/simba/commit/058fae2cea30d88deff044d338ab51d5d266da6c))
* set includeComponentInTag to true for linked-versions plugin ([1d900df](https://github.com/sillen102/simba/commit/1d900df4bb6e264f1dac2a7aabf8a2bc103807dd))
* set includeComponentInTag to true for linked-versions plugin ([ff8386a](https://github.com/sillen102/simba/commit/ff8386ac7518a53160d8db3679e1902752853ac6))
* sync module versions ([52d4469](https://github.com/sillen102/simba/commit/52d446934d61f6b362e58c97f4193538ea2f688c))
* update group PR title pattern to include placeholders ([a598bc0](https://github.com/sillen102/simba/commit/a598bc04a35ee0fb6cdebd1abf7dc2cd428021cb))
* use release please ([ed8b0e7](https://github.com/sillen102/simba/commit/ed8b0e7ad29c077adaf7f06930ba797b08956efe))

# Changelog

All notable changes to this project will be documented in this file.
