# Design notes

Why this code looks the way it does, what each choice costs, and what I would do next. 

## Scope

I built the four endpoints requested by the exercise and have left out two from implementation:
* search by media: not implemented but the data model is designed to support this feature
* get media: as I have used local filesystem for media storage, ideally we should have an endpoint to retrieve the media with the URL. 
For simplicity I have left this out of the scope as well

**Production-ready definition for this exercise**

* code easy to read, maintainable, easy to evolve
* good error handling with failures having defined status codes 
* graceful shutdown
* structured logs 
* domain is well tested with unit tests

**What I have left out**
* Authentication
* Observability
* CI/CD


## Architecture

Hexagonal with three well definied layers with its boundaries. The core (`domain` + `application`) should know nothing about HTTP
or PostgreSQL; the adapters implement ports the core owns. `cmd/api` is the only place that
sees all three sides.

## Technology choices

* chi for routing HTTP requests. Not as complex as gin, but it serves me well for a tiny API
* pgx instead of using an ORM
* go-envconfig simple configuration management
* google/uuid UUIDv7 generation
* testify for tests assertions

## Data model

```sql
tags        (id UUID PK, name TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL)
            UNIQUE INDEX tags_name_lower_key ON (lower(name))

media       (id UUID PK, name TEXT NOT NULL, storage_key TEXT NOT NULL UNIQUE,
             type TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL)

media_tags  (media_id UUID → media ON DELETE CASCADE,
             tag_id   UUID → tags  ON DELETE RESTRICT,
             PRIMARY KEY (media_id, tag_id))
            INDEX idx_media_tags_tag_id ON (tag_id, media_id)
```

A join table fits better our requirements instead of embedding an array into the media data model. 
Faster lookups and easier to fine tune it with indexes. The counterpart is that the ports implementations and the use cases
are bit more complex.

IDs are generated in the application.

### The index for the search that isn't built

The brief's future search is "all media tagged X". That query is:

```sql
SELECT m.id, m.name
FROM media_tags mt
JOIN media m ON m.id = mt.media_id
JOIN tags  t ON t.id = mt.tag_id
WHERE lower(t.name) = lower($1);
```

The index `idx_media_tags_tag_id (tag_id, media_id)` makes this query cheap. The composite primary key
is not enough as we would need to read all media_id before finding the tag_ids. 

## API decisions

**`tags` may be empty on upload → `201`.** The asset is the file; tagging is optional. 

Get-or-create-by-name at upload time is *not* implemented for simplicity.

**Duplicate tag name → `409`, case-insensitively.** "Messi" and "messi" should be the same tag to avoid confusions and incorrect tagging

### The `FileStore` port

This one is not completed on purpose for the sake of time. It should also provide a method to get the files.

```go
type FileStore interface {
	Put(ctx context.Context, key string, r io.Reader) (string, error)
	URL(key string) string
}
```

### Transactionality and failure modes

`CreateWithTags` is one transaction: resolve the tag ids, insert the media row, insert the
links, commit. An unknown tag id aborts it, so a `400` leaves zero rows.

The ordering is: write the object, then commit the row. That leaves one failure window — object
written, transaction failed — and the result is an **orphaned object**. I have decided to log this at `WARN` level with
the file key:

```
WARN media file orphaned: stored but not persisted in db storage_key=01a0…f189
```

I picked that direction on purpose for simplicity but this should be taken into account when rolling out to production.
Maybe just track the metrics and if this happens too often then apply a better solution

### Error model

One envelope, `{"error":{"code","message","details?"}}`, and one mapping function in the HTTP
adapter. Domain errors carry no status codes; the transport decides those.

Unmapped errors return a generic `500` with the real error logged server-side.


## Testing

- `domain`: table-driven tests on `NewTag` / `NewMedia` / `ParseMediaType` — trimming, the length
  limits, empty values, an incorrect `MediaType`.
- `application`: with hand-written spies for the three ports.
- `infra/in/http`: `httptest` against fakes of local interfaces.
- `infra/out/postgres`: the real repositories against a PostgreSQL container, migrations
  applied from `migrations/`. The atomicity proof lives here — unknown tag ⇒ error *and* zero
  rows in `media`.
- `infra/out/storage`: using `t.TempDir()`

## What would I do with more time

1. **CI pipelines in GitHub**
1. **Request ids and context-propagated log correlation.** 
2. **The search endpoint** 
3. **S3 adapter** with presigned URLs
4. **`FileStore.Delete` plus a reconciliation mechanism** over `storage_key`, and an `Idempotency-Key` on upload to avoid orphan files
5. **Load** test with k6 and fine tune database model
