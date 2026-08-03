# Changelog

## [0.4.0](https://github.com/themaiby/stet/compare/v0.3.1...v0.4.0) (2026-08-03)


### Features

* add a skill for writing, not only for checking ([#17](https://github.com/themaiby/stet/issues/17)) ([683a999](https://github.com/themaiby/stet/commit/683a99950392a15afb12261628df8c1c92c0b65d))


### Fixes

* refuse a target that is not there, and find the rules from anywhere ([#19](https://github.com/themaiby/stet/issues/19)) ([8b91ac3](https://github.com/themaiby/stet/commit/8b91ac3180d7c9295246d2ab78b7a4e1b7611492))

## [0.3.1](https://github.com/themaiby/stet/compare/v0.3.0...v0.3.1) (2026-08-03)


### Fixes

* stop naming the standard hooks file in the manifest ([#15](https://github.com/themaiby/stet/issues/15)) ([c374fe4](https://github.com/themaiby/stet/commit/c374fe496a882ef5f9988464773aacecf335bcfa))

## [0.3.0](https://github.com/themaiby/stet/compare/v0.2.0...v0.3.0) (2026-08-03)


### Features

* check English grammar beside the pattern rules ([#11](https://github.com/themaiby/stet/issues/11)) ([110297f](https://github.com/themaiby/stet/commit/110297f676cba0b34fbfbdb69bb5f249cc66cf01))
* keep named paths out of the linter and the formatter ([#8](https://github.com/themaiby/stet/issues/8)) ([b361beb](https://github.com/themaiby/stet/commit/b361beb072a116e27f8a0e97e2a014375e556c19))
* let the skill keep a project's own vocabulary ([#6](https://github.com/themaiby/stet/issues/6)) ([169b41f](https://github.com/themaiby/stet/commit/169b41f609459711696d839eabb285016f0b5f9d))
* lint this repository with its own measured preset ([#14](https://github.com/themaiby/stet/issues/14)) ([2e4dc67](https://github.com/themaiby/stet/commit/2e4dc673c4ab1837f529f3e263d9c4cfc3d6dc9a))


### Fixes

* point the vocabulary at the project rather than the plugin ([#7](https://github.com/themaiby/stet/issues/7)) ([d12558a](https://github.com/themaiby/stet/commit/d12558a6b5a8f8dd5723f639f9a802a937be017f))
* stop the barbarism message naming the wrong word ([#10](https://github.com/themaiby/stet/issues/10)) ([1450d1c](https://github.com/themaiby/stet/commit/1450d1c4c76f3496c7d43ad2a59d34b60b652119))
* take the language from the config when --lang is absent ([#13](https://github.com/themaiby/stet/issues/13)) ([396ac13](https://github.com/themaiby/stet/commit/396ac13a222b50bdd700da61bfebe103807bcacd))


### Documentation

* state the order the tools run in ([#12](https://github.com/themaiby/stet/issues/12)) ([a2aed46](https://github.com/themaiby/stet/commit/a2aed464a536d3c4c553fae5171f36371c3dd455))


### Refactor

* split the skill by the task the user is asking for ([#9](https://github.com/themaiby/stet/issues/9)) ([2c2f28f](https://github.com/themaiby/stet/commit/2c2f28fd1790d1a09a8c0d6df9983550a41e6e46))


### Chores

* leave the generated changelog out of the formatter ([#5](https://github.com/themaiby/stet/issues/5)) ([8d451e1](https://github.com/themaiby/stet/commit/8d451e188b50ebe4bd61c27aacb7997acf9de412))

## [0.2.0](https://github.com/themaiby/stet/compare/v0.1.0...v0.2.0) (2026-08-03)


### Features

* prose linting for Ukrainian and English through Vale ([996ed8b](https://github.com/themaiby/stet/commit/996ed8bc324f9d58b0693b147aa94915b445986d))
* replace the shell and Python layer with a single Go binary ([2a77d13](https://github.com/themaiby/stet/commit/2a77d13562e8d73797a22b1a435f49d189f98246))


### Chores

* gate pull requests and publish release binaries ([#2](https://github.com/themaiby/stet/issues/2)) ([c2beeb6](https://github.com/themaiby/stet/commit/c2beeb6c6203d004dab80c8bc15b04e9359193b3))
* keep working notes and agent files out of the tree ([537d9f4](https://github.com/themaiby/stet/commit/537d9f4144cc322a606220014b8b1963a86b0a48))
