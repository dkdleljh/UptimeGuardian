# Changelog

All notable changes to this project will be documented in this file.

This project follows a simple semantic versioning scheme (`vMAJOR.MINOR.PATCH`).

## [Unreleased]

## [1.1.0] - 2026-03-07

### Added
- Optional API authentication via `API_KEY`.
  - Accepts `X-API-Key: <API_KEY>` or `Authorization: Bearer <API_KEY>`.
- Network timeouts:
  - `HTTP_TIMEOUT_SEC` for endpoint checks
  - `WEBHOOK_TIMEOUT_SEC` for webhook alert delivery
- Stricter request validation:
  - Reject unknown JSON fields
  - 1MB JSON body limit
- Operational safety improvements:
  - Stable ordering for `GET /checks`
  - Safer persistence file permissions (`0600`)
  - `.gitignore` for `.env` and `data/`

### Changed
- Documentation expanded (`README.md`, `USER_GUIDE_KO.md`).

## [1.0.0] - 2026-03-07

### Added
- Initial release: CRUD checks API, scheduler, transition-based webhook alerts, JSON file persistence.
