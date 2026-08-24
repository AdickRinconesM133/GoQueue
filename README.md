# GoQueue

A lightweight background job processing system built in Go, inspired by Sidekiq and Temporal. Designed as a learning project to explore distributed systems concepts and AWS serverless infrastructure in depth.

## Why this project

Most CRUD APIs don't force you to deal with the hard parts of backend engineering: what happens when a job fails? How do you avoid processing the same job twice? How do you scale workers independently from your API? This project exists to answer those questions with real, working code — not just theory.

All code in this repository is written by hand, without AI code generation. AI tools are used only for non-code assets like this README. The goal is genuine, first-principles learning of Go and distributed systems concepts.

## Design Rationale

The architecture (SQS + Lambda + Redis for idempotency + dead-letter queues) mirrors patterns used in production job processing systems at scale — this isn't an invented setup, it's how AWS itself recommends handling asynchronous, retryable work.

That said, it's worth being upfront: for the actual job volume this project handles as a portfolio piece, this level of infrastructure is more than the problem strictly requires — a simple cron job would suffice at this scale. The architecture is deliberately chosen to practice patterns that become necessary once volume and reliability requirements grow, not because this specific project needs them today. Understanding _when_ to reach for this level of infrastructure (and when not to) is part of the point.

## Features (in progress)

- [x] REST API to enqueue jobs
- [ ] Job processing via worker pool
- [ ] Exponential backoff retries on failure
- [ ] Dead-letter queue for permanently failed jobs
- [ ] Idempotency guarantees (no duplicate processing)
- [ ] AWS SQS integration for durable queuing
- [ ] AWS Lambda-based worker for serverless processing
- [ ] Job status endpoint

## Tech Stack

- **Language:** Go
- **Router:** [chi](https://github.com/go-chi/chi)
- **Queue:** AWS SQS
- **Cache / Idempotency store:** Redis
- **Serverless compute:** AWS Lambda
- **Database (planned):** PostgreSQL, for persistent job history

## Architecture (planned)

```
Client → REST API → SQS → Worker (Go / Lambda) → Redis (idempotency check) → Job executed
                                    ↓ (on repeated failure)
                              Dead-letter queue
```

## Getting Started

```bash
git clone https://github.com/<your-username>/GoQueue.git
cd GoQueue
go mod tidy
go run main.go
```

Server runs on `localhost:8080`.

### Enqueue a job

```bash
curl -X POST http://localhost:8080/jobs \
  -d '{"type":"send_email","payload":{"to":"test@test.com"}}'
```

## Roadmap

This project is being built incrementally in phases:

1. **Phase 1** — Basic API + in-memory job handling (current)
2. **Phase 2** — Retries with exponential backoff + dead-letter queue
3. **Phase 3** — Idempotency via Redis
4. **Phase 4** — AWS SQS integration + Lambda-based worker
5. **Phase 5** — Job status endpoint + PostgreSQL persistence

## License

MIT
