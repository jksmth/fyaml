# Changelog

## [1.12.0](https://github.com/jksmth/fyaml/compare/v1.11.0...v1.12.0) (2026-01-10)


### Features

* add public API for programmatic usage ([#67](https://github.com/jksmth/fyaml/issues/67)) ([0b2d079](https://github.com/jksmth/fyaml/commit/0b2d079558e787b34fd367e64a6deb7c9dbebec6))

## [1.11.0](https://github.com/jksmth/fyaml/compare/v1.10.0...v1.11.0) (2026-01-09)


### Features

* add stdin support for --check command ([#64](https://github.com/jksmth/fyaml/issues/64)) ([c918a88](https://github.com/jksmth/fyaml/commit/c918a88f8fb8e2e7b0e7fb584b330e0d5d4a4970))

## [1.10.0](https://github.com/jksmth/fyaml/compare/v1.9.1...v1.10.0) (2026-01-09)


### Features

* add shallow and deep merge strategies ([#62](https://github.com/jksmth/fyaml/issues/62)) ([ce94ac6](https://github.com/jksmth/fyaml/commit/ce94ac6fdb43a2f4384dea77490a6c027f718e96))

## [1.9.1](https://github.com/jksmth/fyaml/compare/v1.9.0...v1.9.1) (2026-01-09)


### Bug Fixes

* normalize non-string map keys for canonical mode and JSON output ([#60](https://github.com/jksmth/fyaml/issues/60)) ([75557ad](https://github.com/jksmth/fyaml/commit/75557ad6f4dd493dce4236aa71ce30dc112c19e6))

## [1.9.0](https://github.com/jksmth/fyaml/compare/v1.8.0...v1.9.0) (2026-01-09)


### Features

* **mode:** add preserve mode for comment and key order preservation ([#56](https://github.com/jksmth/fyaml/issues/56)) ([b6ea4af](https://github.com/jksmth/fyaml/commit/b6ea4afa2a43a10d63e76744d6d29565ec422aae))

## [1.8.0](https://github.com/jksmth/fyaml/compare/v1.7.0...v1.8.0) (2026-01-04)


### Features

* move pack to rootCmd and update docs for default directory ([#54](https://github.com/jksmth/fyaml/issues/54)) ([57e1eaa](https://github.com/jksmth/fyaml/commit/57e1eaa0aed20e9df5e20f18ed54d5b065b61def))

## [1.7.0](https://github.com/jksmth/fyaml/compare/v1.6.0...v1.7.0) (2026-01-03)


### Features

* add --indent flag for YAML/JSON output ([4766c2a](https://github.com/jksmth/fyaml/commit/4766c2a9812a35bed1f23c4eb26362258f497315))
* Improve documentation, add indent flag, and fix error handling ([#44](https://github.com/jksmth/fyaml/issues/44)) ([4766c2a](https://github.com/jksmth/fyaml/commit/4766c2a9812a35bed1f23c4eb26362258f497315))


### Bug Fixes

* replace panic with error handling in mergeTree ([4766c2a](https://github.com/jksmth/fyaml/commit/4766c2a9812a35bed1f23c4eb26362258f497315))

## [1.6.0](https://github.com/jksmth/fyaml/compare/v1.5.0...v1.6.0) (2026-01-03)


### Features

* **extensions:** add @ directory support ([#42](https://github.com/jksmth/fyaml/issues/42)) ([b25cc8a](https://github.com/jksmth/fyaml/commit/b25cc8a242019d17ad6837e5bb12f54676bc459e))

## [1.5.0](https://github.com/jksmth/fyaml/compare/v1.4.0...v1.5.0) (2026-01-03)


### Features

* **includes:** add unified tag-based include system ([#39](https://github.com/jksmth/fyaml/issues/39)) ([9e927c2](https://github.com/jksmth/fyaml/commit/9e927c29de15834dfbd1fc2d6f9b9a308e2ea7fa))

## [1.4.0](https://github.com/jksmth/fyaml/compare/v1.3.0...v1.4.0) (2026-01-02)


### Features

* add --convert-booleans flag and improve CLI organization ([#37](https://github.com/jksmth/fyaml/issues/37)) ([c0c14b5](https://github.com/jksmth/fyaml/commit/c0c14b58a508d505e1b31856df5788770c408f30))

## [1.3.0](https://github.com/jksmth/fyaml/compare/v1.2.0...v1.3.0) (2026-01-02)


### Features

* add verbose logging with shared logger package ([#34](https://github.com/jksmth/fyaml/issues/34)) ([8428688](https://github.com/jksmth/fyaml/commit/842868805ebdfef82b7fddc14f70c9f272555d21))

## [1.2.0](https://github.com/jksmth/fyaml/compare/v1.1.0...v1.2.0) (2026-01-01)


### Features

* Include path confinement ([#31](https://github.com/jksmth/fyaml/issues/31)) ([7980b7d](https://github.com/jksmth/fyaml/commit/7980b7d68776d3d0f884b1938679437b27976a83))

## [1.1.0](https://github.com/jksmth/fyaml/compare/v1.0.7...v1.1.0) (2026-01-01)


### Features

* add support for file includes ([#29](https://github.com/jksmth/fyaml/issues/29)) ([674aca0](https://github.com/jksmth/fyaml/commit/674aca08bb61c949333868d2beaab427b4adb310))

## [1.0.7](https://github.com/jksmth/fyaml/compare/v1.0.6...v1.0.7) (2025-12-30)


### Bug Fixes

* update license date ([#24](https://github.com/jksmth/fyaml/issues/24)) ([9a668b4](https://github.com/jksmth/fyaml/commit/9a668b481307b4e9088fa60bd77f939bfc42bff0))

## [1.0.6](https://github.com/jksmth/fyaml/compare/v1.0.5...v1.0.6) (2025-12-30)


### Bug Fixes

* update release workflow to use goreleaser for the GH release ([#21](https://github.com/jksmth/fyaml/issues/21)) ([674e12b](https://github.com/jksmth/fyaml/commit/674e12b173c0eb20c73494ce9489052032db6539))

## [1.0.5](https://github.com/jksmth/fyaml/compare/v1.0.4...v1.0.5) (2025-12-30)


### Bug Fixes

* use config file for release-please ([#18](https://github.com/jksmth/fyaml/issues/18)) ([b35d2fe](https://github.com/jksmth/fyaml/commit/b35d2fe9d053d456f4fdc256668a4661ee16d2aa))
* use default options for release-please ([#19](https://github.com/jksmth/fyaml/issues/19)) ([d1663ee](https://github.com/jksmth/fyaml/commit/d1663ee9d58647ee1612df63c206993d80c41226))

## [1.0.4](https://github.com/jksmth/fyaml/compare/v1.0.3...v1.0.4) (2025-12-30)


### Bug Fixes

* use keep-existing mode for goreleaser release ([#16](https://github.com/jksmth/fyaml/issues/16)) ([ef67007](https://github.com/jksmth/fyaml/commit/ef67007a3f548d8a96980ba3f84ec9de28211bd6))

## [1.0.3](https://github.com/jksmth/fyaml/compare/v1.0.2...v1.0.3) (2025-12-29)


### Bug Fixes

* set ids for sign in goreleaser ([#11](https://github.com/jksmth/fyaml/issues/11)) ([27dfec2](https://github.com/jksmth/fyaml/commit/27dfec2c1b0a992a0e20ba3c6acbb415c481aaab))

## [1.0.2](https://github.com/jksmth/fyaml/compare/v1.0.1...v1.0.2) (2025-12-29)


### Bug Fixes

* use PAT for release-please ([#9](https://github.com/jksmth/fyaml/issues/9)) ([3c0dda8](https://github.com/jksmth/fyaml/commit/3c0dda8e0cd370213bf502a5cb14c30aa9264eae))

## [1.0.1](https://github.com/jksmth/fyaml/compare/v1.0.0...v1.0.1) (2025-12-29)


### Bug Fixes

* add release-please and release automation ([#5](https://github.com/jksmth/fyaml/issues/5)) ([0f9a0a1](https://github.com/jksmth/fyaml/commit/0f9a0a10338325e3aff9435a2d9c2c6c35d74f76))

## [Unreleased]
