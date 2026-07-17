# dhook roadmap

## Purpose

Make dhook a dependable, small Discord webhook client for Go by proving the
delivery path first, then improving the public API and adoption. This roadmap
tracks shipped work separately from future goals; it is not a claim of current
usage or certification.

## Foundation: current working tree

The following work is implemented and covered by repository tests, pending a
release:

- HTTP contract tests cover `204 No Content`, JSON message responses, and
  adding `wait=true` without dropping an existing query parameter.
- Send operations reject an empty webhook configuration instead of reporting a
  false successful delivery.
- Fractional `Retry-After` values are parsed and waited for.
- `Queue.Stop` drains work accepted before stopping and ignores submissions
  made afterward.
- CI targets Go 1.21, 1.22, and 1.23 and runs formatting, `go vet`, build,
  and race-detector test gates on pushes and pull requests to `master`.

## v0.2: delivery contract

- Define and document which HTTP and network failures retry, and which do not.
- Add context-cancellation, 4xx/5xx, rate-limit, and concurrent-send tests.
- Validate webhook URLs and delivery configuration at public API boundaries.
- Return inspectable errors for invalid webhook URLs, rate limits, and Discord
  API responses.
- Document retry and queue semantics, including ordering, full-queue behavior,
  shutdown, and delivery guarantees.

**Exit evidence:** the behavior is tested with `httptest`, race tests pass in
CI, and the documented contract matches the implementation.

## v0.3: developer experience

- Replace untyped event registration only if a concrete misuse cannot be
  handled by the existing API; keep the smallest compatible surface otherwise.
- Add compile-checked examples for messages, embeds, files, queues, and
  context cancellation.
- Clarify errors, configuration, rate limits, and supported Go versions in the
  README and package documentation.
- Publish user-visible change notes for each version.

**Exit evidence:** a new Go user can follow an example, understand failures,
and run the documented checks without private knowledge of the project.

## v0.4: one useful integration

- Choose one real automation integration based on user demand, with a GitHub
  Action as the first candidate.
- Version it, document secret handling, and provide an end-to-end example.
- Collect feedback from genuine users before adding more integrations.

**Non-goal:** do not create multiple thin integrations merely to inflate the
project surface.

## v1.0: stable public API

Publish v1 only after the delivery API has external use, its compatibility
rules are documented, and breaking changes discovered during pre-v1 use have
been resolved. A version number alone is not proof of production readiness.

## Adoption and maintenance evidence

Track only verifiable, public signals:

- dependent repositories or integration users;
- external issues, merged pull requests, and contributors;
- release history and response time for reported defects;
- package or action usage metrics where the publisher provides them.

Use those facts, plus test and CI evidence, when describing dhook in an
open-source program application. Do not manufacture stars, downloads,
dependents, contributors, testimonials, or security scores. Do not claim
coverage targets, external users, an OpenSSF score, a GitHub Action, or a
release until evidence exists.
