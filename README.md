# ATS Backend

Backend for an Applicant Tracking System. Accepts CV files (PDF), extracts structured data using Claude AI, and stores the result in a database.

## What it does

- Accepts a PDF CV via HTTP upload
- Extracts text from the PDF
- Sends the text to Claude (Anthropic) which fills in structured fields: name, title, specialization, skills, work experience
- Saves the CV and its skills to PostgreSQL

## Stack

- **Go 1.26** — application
- **Gin** — HTTP server
- **PostgreSQL 17** — storage; skills are normalized in a separate table with a many-to-many join
- **pgx v5** — database driver (connection pool)
- **Claude claude-sonnet-4-5** (Anthropic SDK) — CV parsing via tool use
- **zerolog** — structured JSON logging

## Requirements

- Docker and Docker Compose
- Anthropic API key

## Running

```bash
cd docker
ANTHROPIC_API_KEY=sk-... docker compose up --build
```

The app starts on port `8080`. PostgreSQL migrations are applied automatically on first start.

## API

### Upload a CV

```
POST /api/v1/cv/upload
Content-Type: multipart/form-data

cv: <pdf file>
```

**Responses:**
- `201` — CV uploaded and saved successfully
- `400` — file is missing
- `422` — failed to process the CV
