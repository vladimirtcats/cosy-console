# cosy-console

A companion demo for the article "A production REPL: Yaegi over the composition
root": a minimal Go backend with two domains (orders, catalog), PostgreSQL, Redis
and an interactive console over the composition root — like `rails c`, only in Go.

## Running

```bash
docker-compose up --build
```

Five services come up: postgres, redis, migrations (one-shot), seed (one-shot)
and api with hot reload via [Air](https://github.com/air-verse/air) — edits to
`.go` files apply without rebuilding the container.

Checking the API:

```bash
curl -s localhost:8080/health

curl -s "localhost:8080/api/v1/items?page=1&page_size=2"

curl -s -X POST localhost:8080/api/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"lines":[{"item_id":"00000000-0000-0000-0000-0000000000a1","qty":2}]}'
```

## Console

A separate mode of the same project tree: it brings up the composition root
(the same databases, redis, repositories, usecases) and hands it to a REPL on
[Yaegi](https://github.com/traefik/yaegi). HTTP handlers are not part of the console.

### Running it

```bash
# interactive session, read-only by default
docker-compose run --rm api go run ./cmd/console

# with writes allowed
docker-compose run --rm api go run ./cmd/console -write

# a one-shot expression, no REPL
docker-compose run --rm api go run ./cmd/console -e '1 + 1'
```

Exit with `Ctrl+D`. `Ctrl+C` cancels the current statement and returns to the
prompt; the session stays alive.

Scripts from `scripts/` can be pasted into a session whole or piped in
(statement echo is off when piping, so only explicit `fmt.Println` output remains):

```bash
docker-compose run --rm api go run ./cmd/console < scripts/check_orders.txt
```

### The production stack

A separate `docker-compose.prod.yml` with its own project name
(`cosy-console-prod`): the image is built from the multi-stage `Dockerfile` (two
static binaries — `/app/api` and `/app/console`), with no source mounts and no
hot reload, and only the API port published. The project name isolates the
production stack from the dev one: `docker-compose up` and `make up-prod` don't
overwrite each other's images and containers. Bring it up with:

```bash
make up-prod
```

The console here is what production looks like: `exec` into the **running** api
container. Not `run --rm` as in dev — in production the api container *is* the
deployment, and the console runs inside it (the image entrypoint is `/app/api`,
so `run` would start the server instead of the console).

```bash
# read-only session
make console-prod
# the same thing by hand
docker-compose -f docker-compose.prod.yml exec api /app/console

# with writes allowed
make console-prod-write
```

Scripts from `scripts/` work here too:

```bash
docker-compose -f docker-compose.prod.yml exec -T api /app/console \
    < scripts/check_orders.txt
```

### What's available

| Symbol | What it is |
|---|---|
| `App` | `*app.Container` — the whole composition root: `App.Orders`, `App.Catalog`, `App.Shared` |
| `Ctx` | `context.Context` for calls that need a context |

Packages are available **without import statements**: `fmt`, `time`, `errors`,
`context` (stdlib) and the domain ones — `orders`, `orders_models`, `catalog`,
`pagination`, `uuid`.

Repositories and usecases are called through live values — their packages don't
need to be in the symbol table. Same for infrastructure: `App.Shared.DB`,
`App.Shared.Tx`, `App.Shared.Redis`.

### Examples

Each block is self-contained — paste it whole into a fresh session. The REPL
prints the result of the last statement with a `: ` prefix:

```text
// an order by id (uuid arguments are constructed inline)
> id, err := uuid.Parse("00000000-0000-0000-0000-0000000000b1")
> order, err := App.Orders.Repos.OrderRepo.Get(Ctx, id)
> order.Status
: pending

// a usecase with a filter and pagination
> items, page, err := App.Catalog.Usecases.Item.List(Ctx, catalog.ItemFilter{SKU: "TEE"}, pagination.WithPageSize(2))
> items[0].Title
: Go gopher tee

// domain params, same as in the wiring code (fails in a read-only session, see below)
> note := "gift wrap"
> updated, err := App.Orders.Repos.OrderRepo.Update(Ctx, id, orders.UpdateAttrs{Note: &note})

// output models are read field by field, no package registration needed
> order.Lines[0].PriceCents
: 990

// redis, as a live value
> err = App.Shared.Redis.Set(Ctx, "probe", "hello", time.Minute).Err()
: <nil>
```

### Transactions

The REPL can't parse a multi-line closure passed as a call argument (a
limitation of stock Yaegi). The pattern that works: define the closure in an
assignment — its body buffers to any size — and keep the call on one line. The
`> ` prompt is printed only before the first line; the rest of a pasted block is
echoed by the terminal without a prompt:

```text
> id, _ := uuid.Parse("00000000-0000-0000-0000-0000000000a1")
> getItem := func(ctx context.Context) error {
item, err := App.Catalog.Repos.ItemRepo.Get(ctx, id)
if err != nil {
return err
}
fmt.Println("inside tx:", item.Title)
return nil
}
> err := App.Shared.Tx.InTransaction(Ctx, getItem)
inside tx: Ceramic mug
: <nil>
> err
: <nil>
```

In one-shot mode (`-e`) a multi-line call with a closure works as is — the whole
thing runs in a single eval:

```bash
docker-compose run --rm api go run ./cmd/console -e 'id, _ := uuid.Parse("00000000-0000-0000-0000-0000000000a1")
err := App.Shared.Tx.InTransaction(Ctx, func(ctx context.Context) error {
	item, err := App.Catalog.Repos.ItemRepo.Get(ctx, id)
	return err
})
err'
```

### Read-only

By default the console connects with `default_transaction_read_only=on` in its
startup parameters: the server rejects every write at the protocol level.

```text
> note := "check"
> _, err := App.Orders.Repos.OrderRepo.Update(Ctx, uuid.MustParse("00000000-0000-0000-0000-0000000000b1"), orders.UpdateAttrs{Note: &note})
> err
: ERROR: cannot execute UPDATE in a read-only transaction (SQLSTATE 25006)
```

Writes take a deliberate `-write` flag. Delivery of the parameter is verified at
startup by asking the server (`SHOW default_transaction_read_only`): a pooler
that strips startup parameters crashes the console at startup instead of leaving
it unprotected.

### Changing data

To write, restart the console with the flag:

```bash
docker-compose run --rm api go run ./cmd/console -write   # or make console-write
```

Changes go through the same repositories and usecases as the application, with
no raw SQL:

```text
// an update through the domain params types
> note := "gift wrap"
> order, err := App.Orders.Repos.OrderRepo.Update(Ctx, uuid.MustParse("00000000-0000-0000-0000-0000000000b1"), orders.UpdateAttrs{Note: &note})
> order.Note
: gift wrap

// a cancel with the business status-transition check (pending → cancelled)
> cancelled, err := App.Orders.Usecases.Order.Cancel(Ctx, uuid.MustParse("00000000-0000-0000-0000-0000000000b1"))

// several operations, atomically, through the transaction manager
> update := func(ctx context.Context) error {
_, err := App.Orders.Repos.OrderRepo.Update(ctx, uuid.MustParse("00000000-0000-0000-0000-0000000000b1"), orders.UpdateAttrs{Note: &note})
return err
}
> err = App.Shared.Tx.InTransaction(Ctx, update)
```

### Gotchas

- **A closure as a call argument** — only as a single statement in the REPL
  (see transactions above).
- **Multi-value assignment**: write `res, err := ...` — with `err := f()` the
  first return value lands in `err`, not the error.
- **The prompt doesn't change** in the middle of multi-line input: silence after
  Enter means the buffer is still filling; a result or an error means the input ran.

## Layout

```
cmd/api/                  — HTTP API
cmd/console/              — console (read-only by default, -write, -e)
config/                   — configuration from ENV
internal/
  app/                    — composition root, router, cross-domain adapters
  domain/
    catalog/              — items domain: models, repo, usecase (+ redis cache)
    orders/               — orders domain: models, repo, usecase (+ pricer through a port)
  infrastructure/         — postgres (UTC pin, read-only), redis
  console/                — console core and symbol table
  transport/http/         — handlers and DTOs
pkg/
  pagination/             — pagination options and result
  transaction/            — transaction manager (ExtractDB from the context)
db/
  migrations/             — one migration file per table
  seeds/                  — idempotent seed with fixed uuids
```

The domains don't know about each other: orders get catalog prices through the
`ItemPricer` port, implemented by an adapter in `internal/app/adapters`; the item
cache goes through the `ActiveItemsCache` port over redis. The console follows
the same principle: it has no dependencies of its own — only `App` and `Ctx`.

## Keeping symbols in sync

The symbol table is a snapshot of a domain's params/models. Change `UpdateAttrs`
or add a model and the snapshot is stale — constructing the type in the console
gives `undefined`. To sync:

```bash
make console-symbols DOMAIN=orders    # a single domain
make console-symbols DOMAIN=orders,catalog   # several, comma-separated
make console-symbols                  # every domain at once
```

What the target does — all of it automatic, nothing to do by hand:

1. runs `yaegi extract` for the domain and its models right inside
   `internal/console/symbols/` (extract writes to the current directory);
2. renames the raw files to `<domain>.go` and `<domain>_models.go`
   (`mv -f`): a missing file is created, an existing one overwritten.

No keys to edit: the session names (`orders`, `orders_models`) are derived by
`symbols.Exports()` in code. After regenerating, check with the tests and one
command:

```bash
docker-compose run --rm api go test ./internal/console/...
docker-compose run --rm api go run ./cmd/console -e 'orders.UpdateAttrs{}.Note == nil'
```

A forgotten regeneration isn't silent: `undefined` shows up on exactly the type
that changed.

## Tests

```bash
docker-compose run --rm api go test ./...
```

The console tests don't need a database: the interpreter starts in milliseconds
and input is scripted through `Options{Stdin: ...}`.

## Migrations

Each table is its own pair of files in `db/migrations`
(`000001_create_users.up.sql` / `.down.sql`, and so on), applied by
[migrate](https://github.com/golang-migrate/migrate) when compose starts.
