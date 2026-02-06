# Changelog

## 1.0.0 (2026-02-06)


### Features

* add --dry-run flag to preview generations ([eeadec2](https://github.com/kevinmichaelchen/replicate-images/commit/eeadec2600303768133a36cb12a2451fb8fbac90))
* add --json flag for machine-readable output ([a02cc63](https://github.com/kevinmichaelchen/replicate-images/commit/a02cc63ce5be09c598b54a80deb97b7a5d231a15))
* add -q/--quiet flag to suppress output ([c121e03](https://github.com/kevinmichaelchen/replicate-images/commit/c121e03cfb23b7273dabe206b3cb57b7bc719e1a))
* add batch command for YAML-based prompt processing ([a56e1e2](https://github.com/kevinmichaelchen/replicate-images/commit/a56e1e25b550f06eeef0cbe40c3976bc1fd7b3f0))
* add install script for curl-based installation ([39c4862](https://github.com/kevinmichaelchen/replicate-images/commit/39c48621ee76c134b2d1d65e76a72d52ab54661d))
* add manual e2e test script ([334fde5](https://github.com/kevinmichaelchen/replicate-images/commit/334fde594e87daabf6cba022ffaaf7d81a9a6533))
* add model registry with 4 supported models ([77d18bd](https://github.com/kevinmichaelchen/replicate-images/commit/77d18bde1acc97efcd200b39f1b738a32f20d6c1))
* add structured exit codes for agent consumption ([5ffcd8a](https://github.com/kevinmichaelchen/replicate-images/commit/5ffcd8a5ef97381562cfa98b6f6d9988af38104a))
* add validate subcommand for YAML file checking ([81bd15a](https://github.com/kevinmichaelchen/replicate-images/commit/81bd15a15aa22d0c69885ea71f1a9b15dd27377e))
* **batch:** support named output files and duplicate name validation ([856752f](https://github.com/kevinmichaelchen/replicate-images/commit/856752f60a9ab98dfc8b32b7578b16bc1ecbdd04))
* initial CLI for generating images via Replicate API ([4debef2](https://github.com/kevinmichaelchen/replicate-images/commit/4debef2ea860432add26902f6656b92bfbba8213))


### Bug Fixes

* **ci:** migrate to golangci-lint v2 ([6b86ad5](https://github.com/kevinmichaelchen/replicate-images/commit/6b86ad571f0c0ce4390161975f8b2ca23a1c3be0))
* **ci:** upgrade to golangci-lint-action v7 ([696fe1d](https://github.com/kevinmichaelchen/replicate-images/commit/696fe1de11a13d12d4200aea8a7970f86f034cb4))
* **ci:** use golangci-lint v2.8.0 for Go 1.25 support ([e48131e](https://github.com/kevinmichaelchen/replicate-images/commit/e48131e63e1c176aed8773cc4d608f53493fe429))
* prevent duplicate cache entries on regeneration ([c674622](https://github.com/kevinmichaelchen/replicate-images/commit/c674622130afa907911f55ebe9684b41b10238cb))
* resolve golangci-lint warnings ([b2925d6](https://github.com/kevinmichaelchen/replicate-images/commit/b2925d6af406e5719f7eb353190a302e844ee8c0))
