# TODO

## Synonyms Governance
- Define `POST /bible/v2/synonyms/suggest` for user submissions (JSON).
- Payload should include:
  - `term` (string)
  - `synonyms` (array of strings)
  - `reasons` (array of strings)
  - `examples` (array of strings, Bible verses or references)
  - `submitted_by` (string, user id or email)
- Store submissions in Redis (short term) or SQLite (long term).
- Add admin review workflow:
  - `GET /admin/synonyms/submissions` (list pending)
  - `POST /admin/synonyms/submissions/{id}/approve`
  - `POST /admin/synonyms/submissions/{id}/deny`
- On approve:
  - Write to OpenSearch synonyms set
  - Reload analyzers for `kjv_v2`
- On deny:
  - Record reason and notify requester (if email is provided)

## Auth + Roles
- Add admin role claim (e.g., `role=admin`) to app JWT.
- Protect admin review endpoints with `role=admin`.
- Add audit logging for approvals/denials.

## UI
- Build a submission form in `/v2`:
  - term, synonyms, reasons, examples
  - submit status + history
- Build an admin review panel:
  - approve/deny with notes
  - filters by book/testament in examples

## Data Model
- Decide storage for submissions:
  - Option A: SQLite table `synonym_submissions`
  - Option B: OpenSearch index `synonym_submissions`
- Include fields: id, status, created_at, reviewed_at, reviewer, payload, decision_note

## Validation
- Enforce schema and limits:
  - max synonyms per request
  - max reasons/examples
  - deny special characters or empty entries

## Notifications
- Email (optional) on approval/denial
- Webhook (optional) for external systems
