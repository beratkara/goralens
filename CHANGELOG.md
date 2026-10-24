# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- Watchers for requests, exceptions, logs, queries, models, events, jobs, mail, cache, gates,
  outgoing HTTP requests, Artisan commands, scheduled tasks, dumps and Redis (go-redis hook).
- Batches that link everything a request, command, job or scheduled task did.
- Embedded dashboard with tag search, related entries, auto-refresh, pause/clear,
  tag monitoring and light/dark themes.
- Database storage (goralens_* tables) and in-memory storage.
- `goralens:prune` and `goralens:clear` commands.
- Authorization, filter, tag and user resolver callbacks; masking of sensitive values.
- `package:install` support.
