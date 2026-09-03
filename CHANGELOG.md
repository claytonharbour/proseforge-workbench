# Changelog

All notable changes to the ProseForge Workbench are documented here.

## [Unreleased]

### Changed (BREAKING — narration tools)

Workbench narration surface now speaks **section** instead of **chapter**, matching the renamed proseforge backend API (proseforge#497 Stage 2, commit `09de4cf0`). Workbench MCP customers should update their callers; old `chapter_*` tool names and `chapter_id` arg names will return clean "unknown tool" / "unknown argument" errors.

- **MCP tools renamed:**
  - `narration_chapter_cancel` → `narration_section_cancel`
- **MCP tool arg renames (every narration tool that previously took `chapter_id`):**
  - `chapter_id` → `section_id`
  - `chapters_json` (on `narration_patch`) → `sections_json`
  - JSON entry field `chapter_id` (in `narration_patch` payloads) → `section_id`
  - `chapter_announcements` (on `narration_rebuild`) → `section_announcements`
  - Tool descriptions and drift-envelope field references updated: `contentChangedChapters` → `contentChangedSections`, `nameChangedChapters` → `nameChangedSections`, `deletedChapters` → `deletedSections` (`insertedSections` was already correct).
- **CLI flag renames (every `pfw story narration *` subcommand that previously took a chapter arg):**
  - Positional `<chapter-id>` → `<section-id>`
  - `--chapter-announcements` → `--section-announcements` (on `rebuild`)
  - `--chapter` → `--section` (on `patch`)
- **Service layer:** `RegenerateChapter` → `RegenerateSection`, `RetryChapter` → `RetrySection`, `CancelChapter` → `CancelSection`. `ListSegments` and `RegenerateSegment` keep their names but parameter renamed `chapterID` → `sectionID`.
- **Wire translation note:** `chapter_announcements` MCP input is internally translated to upstream `chapter_announcements` JSON tag until `proseforge#520` ships the backend tag rename — workbench MCP customers see only "section" naming throughout.

What still exposes "chapter" after this release:
- M4B audiobook chapter markers + "Chapter N" audio announcements — correct audiobook-format terminology, stays.
- `/chapters/` prefix in legacy `audio_url` Spaces paths — fixable upstream only (`proseforge#521`).

Cross-repo coordination: deploy with backend prod flip (proseforge first, then workbench within minutes).

## [0.5.0] - 2026-04-05

### Added
- `series_plan` tool — StorySeed handoff from Series Forge to Story Forge
- Version history tools: `story_versions`, `story_version_get`, `story_version_diff`
- `story_assess_version` tool — quality assessment at a specific version SHA
- Story visibility support: `story_publish` accepts optional `visibility` parameter,
  new `story_update_visibility` tool for toggling public/members on published stories
- Graceful `members_only` error handling for 403 responses on members-only stories
- `CHANGELOG.md` for tracking release history

## [0.4.0] - 2026-03-29

### Added
- Image generation and management tools: `image_generate`, `image_get`, `image_list`,
  `image_regenerate`, `image_upload`, `story_image_attach`, `story_image_cover`,
  `story_images`
- Credit estimation and transaction history tools: `credits_balance`,
  `credits_estimate`, `credits_history`
- Search, sort, and content filtering on `story_list` and `story_get`

### Fixed
- Image attach/cover endpoints handling empty 201 response bodies
- Image upload MIME type detection

## [0.3.0] - 2026-03-28

### Added
- Narration tools: `narration_start`, `narration_status`, `narration_audiobook`,
  `narration_regenerate`, `narration_retry`, `narration_rebuild`, `narration_delete`,
  `narration_resume`, `narration_chapter_cancel` (renamed to `narration_section_cancel` in Unreleased)
- Segment-level narration tools: `narration_segments`, `narration_segment_regenerate`
- Narration patch endpoint for batch segment/chapter operations with single rebuild (chapter terminology renamed to section in Unreleased)
- Voice selection and force-regenerate support on chapter and segment tools
- Credit management tools
- Series Forge tools (25 tools): series CRUD, characters, world-building, timeline,
  chat sessions, story planning
- Story Forge tools (10 tools): chat-based story generation pipeline, meta approval,
  generation status
- Vanity URL resolution (`story_resolve`)
- Chain hints in MCP tool descriptions for better AI agent discoverability
- Ghostwriter prompt templates for full story rewrites

## [0.2.2] - 2026-03-24

### Changed
- Standardized structured logging across all service methods

## [0.2.1] - 2026-03-24

### Changed
- Moved genre resolution into story service, removed duplicate implementations

## [0.2.0] - 2026-03-24

### Added
- `review_active` tool — find active review assignments with reviewId
- 17 unit tests across the service layer

### Fixed
- P0 code quality issues for public release
- ReviewId surfaced in review accept/decline/approve/reject responses

## [0.1.0] - 2026-03-24

Initial public release.

### Added
- CLI (`pfw`) with noun-verb command structure for all ProseForge API operations
- MCP server (`pfw-mcp`) with 17 tools for AI-assisted story review
- Story tools: list, get, export, section read/write, create, update, publish
- Review tools: accept, decline, approve, reject, list pending
- Feedback tools: create, list, get, diff, suggestions, item add, section update,
  submit, incorporate
- Reviewer pool tools: list available, request reviewer
- Quality and insights tools: assess, get scores, get insights
- Per-call `url` and `token` overrides on all MCP tools for multi-account workflows
- Retry transport with exponential backoff for transient failures
- Structured error classification (network, auth, validation, rate limit, server)
- Audit logging for all MCP tool calls
- Auto-generated CLI reference documentation

[Unreleased]: https://github.com/claytonharbour/proseforge-workbench/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/claytonharbour/proseforge-workbench/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/claytonharbour/proseforge-workbench/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/claytonharbour/proseforge-workbench/compare/v0.2.2...v0.3.0
[0.2.2]: https://github.com/claytonharbour/proseforge-workbench/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/claytonharbour/proseforge-workbench/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/claytonharbour/proseforge-workbench/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/claytonharbour/proseforge-workbench/releases/tag/v0.1.0
