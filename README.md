# Goal

Build a small backend service that manages candidates and their job applications moving through a pipeline.
Expose an HTTP JSON API, persist to a database, and include tests.

## Required API endpoints (MVP)
- POST /candidates — create a candidate (email must be unique).
- POST /applications — create an application for an existing candidate.
- GET /applications — list with filtering (status, role) and pagination (limit + cursor or offset).
- GET /applications/{id} — include candidate info and the application’s event log.
- POST /applications/{id}/transition — body: { "to_status": "interview", "reason": "…" }.

## Requirements
It assumes that you have your environment set up to run [Go](https://go.dev/) code. It also assumes that you installed all packages and have [postgresql](https://www.postgresql.org/download/) on your machine. Do not forget to use a connection assistant like SQLTool to be able to connect to your database

## cURL

```

curl --location 'http://localhost:8080/candidates' \
--header 'Content-Type: application/json' \
--data-raw '{
    "FirstName": "a",
    "LastName": "a",
    "Email": "aaaa@mail.com"
}'

```

```

curl --location 'http://localhost:8080/applications' \
--header 'Content-Type: application/json' \
--data '{
    "CandidateID": "69eb48db-b060-4323-9a89-a7b2395b5439",
    "Role": "b"
}'

```

```

curl --location 'http://localhost:8080/applications'

```

```

curl --location 'http://localhost:8080/applications/b6e67ed4-0fcf-4c64-a4b9-538f663728c8/transition' \
--header 'Idempotency-Key: a' \
--header 'Content-Type: application/json' \
--data '{
    "ToStatus": "hired",
    "Reason": "a"
}'

```

## What remains to do

- Add proper migration
- Add proper testing
- Refactor
- add Docker