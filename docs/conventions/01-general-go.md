# General Go Style

## Formatting and Files

- Every Go file must be formatable by `gofmt`; the workspace runs the Go formatter on save.
- Use `snake_case.go` filenames named after their responsibility, such as `root_shelf_service.go` and `root_shelf_controller_test.go`.
- Keep package names lowercase and singular unless an established plural convention exists; do not create a package for a single use. Each directory under `contracts/core/v1/types/<domain>/` is an independent Go package, but its package clause consistently uses `coretypes`; when a file imports multiple domains, use domain-prefixed aliases to avoid collisions.
- Group imports with one blank line between responsibility groups in this order: standard library, third-party Go modules, `contracts`, `shared`, then runtime-owned imports. Same-layer imports may use their package's direct name. Cross-layer imports must make the direction visible: runtime → `contracts` uses a `c...` alias, runtime → `shared` uses an `s...` alias, and `shared` → `contracts` uses a `c...` alias. Use the package's real singular/plural form after the prefix, such as `cenums`, `cschemas`, `spostgres`, `srepositories`, `sschemas`, and `sinputs`; never use redundant aliases such as `platformpostgres`, `platformrepositories`, or `cenumcontract`. `contracts` and `shared` must not import runtime-owned packages.
- Expand every non-empty struct literal across multiple lines, with one field per line and a trailing comma. Do not put DTO, response, or even a one-field struct literal on one line. This also applies to struct literals in `return` statements and nested struct literals.

  ```go
  return &UpdateMyStationByIdResponse{
  	UpdatedAt: station.UpdatedAt,
  }, nil
  ```

## Method Call Arguments

- Keep method/function call arguments on one line only when there are few and they are simple. When a call has more than two arguments, nested expressions, distinct semantic groups, or enough horizontal content to obscure each argument's role, expand to one argument per line with a trailing comma.
- Repository calls are especially strict: normal domain arguments and `options.With...` arguments are separate semantic groups; two or more options always require multiple lines. Even one option should be multiline when the call already has multiple domain arguments.
- Do not compress multiple `options.With...` calls onto one line or omit clearly named options to preserve a one-line call. Put each option on its own line and keep the established semantic order: DB/transaction, permission, soft-delete, locking, then batch.

```go
station, permission, exception := s.stationRepository.CheckPermissionAndGetOneById(
	request.Body.StationId,
	request.ContextFields.UserId,
	nil,
	allowedPermissions,
	options.WithTransactionDB(tx),
	options.WithAllowedPermissions(allowedPermissions),
	options.WithOnlyDeleted(types.Ternary_Negative),
	options.WithLockingStrength(options.LockingStrengthUpdate),
)
```

## Naming and Types

- Use `PascalCase` for exported types/functions and `camelCase` for unexported identifiers. Follow the project's existing Go abbreviation style, such as `Id`, `Url`, and `Db`; do not mix spellings within one domain.
- Versioned transport DTOs use the complete `XxxRequestDto` and `XxxResponseDto` names; database write models use `XxxInput`; database tables use `schemas.Xxx`. `gatewaycontract.Request[Dto]` and `gatewaycontract.Response[Dto]` are the transport envelopes between Gateway and Core; the DTO is the operation data unit.
- Align new files, transport controllers, adapters, services, repositories, and scopes by domain, for example `station_*`.
- Create an interface only when the caller needs a replaceable implementation, a boundary already exists, or tests genuinely require it. New interfaces follow the existing `XxxInterface` convention; do not pre-abstract a single struct.

## Service DTO and Repository Input Boundaries

- `XxxInput` under `shared/platform/postgres/repositories/inputs/` is a repository persistence contract: it describes data for create, update, partial-update, or bulk SQL and may only be used as repository input. Services import the shared package as `inputs`; controllers and Gateways must not use it as a transport request/response contract.
- When a service method has many parameters, represents one complete intent, or receives a Gateway request that must produce an external response, use `*XxxRequestDto` as the single request parameter and return `*XxxResponseDto`. Name variables `request` and `response`; do not introduce unclear `req` or `res` abbreviations.
- Service-only or Gateway-only workflows may also use concrete `XxxRequestDto`/`XxxResponseDto` types to group related data, context, and output instead of long parameter lists or anonymous structs. The name must describe the operation; do not use generic names such as `Data`, `Params`, or `Payload`.
- Keep direct parameters when there are few and their meaning is clear; do not force a DTO merely to apply a pattern. Create one only when parameter count, shared lifecycle, or a call boundary materially improves readability.
- A service is the conversion boundary between the two contracts: map the validated `request` to `inputs.CreateXxxInput`, `inputs.PartialUpdateXxxInput`, or a bulk input, then call the repository. Repositories import the same runtime's `inputs` package; they must not import or depend on transport request/response types.

```go
func (s *StationService) UpdateMyStationById(
	ctx context.Context,
	request *UpdateMyStationByIdRequestDto,
) (*UpdateMyStationByIdResponseDto, *exceptions.Exception) {
	input := inputs.PartialUpdateStationInput{
		Values: inputs.UpdateStationInput{
			Name: request.Body.Values.Name,
		},
		SetNull: request.Body.SetNull,
	}

	station, exception := s.stationRepository.UpdateOneById(
		request.Body.StationId,
		request.ContextFields.UserId,
		input,
	)
	if exception != nil {
		return nil, exception
	}

	return &UpdateMyStationByIdResponseDto{
		UpdatedAt: station.UpdatedAt,
	}, nil
}
```

## Implementation Principles

- Follow the pattern of adjacent files in the same domain before writing a helper; reuse an existing repository, scope, option, validator, or exception instead of recreating it.
- Each function should work at one layer. Keep public HTTP parsing/validation and HTTP responses in the Gateway controller, workflows and transactions in services, and queries/permission SQL in repositories/scopes.
- Pass `context.Context` downward: HTTP services obtain it from `ctx.Request.Context()`, and DB sessions use `db.WithContext(ctx)`.
- Never swallow errors. Map expected business/infrastructure failures to `*exceptions.Exception` and preserve the cause with `WithOrigin(err)`.
- Prefer direct, readable code; do not use generics, reflection, global state, or a new dependency to solve a single feature.

## Helper and Constructor Rules

- A helper must have a responsibility that is meaningful at its owning layer. Do
  not create a constructor-like helper for a DTO, response DTO, database model,
  or another type that is not owned by the enclosing component. Map that value
  inline at the call site unless the mapping is a real, repeated domain
  operation owned by a dedicated component.
- Do not keep a helper that only wraps a return value, immediately returns a
  delegated call, performs a shallow pointer conversion, or adds a simple
  `if err != nil` before returning. Keep that code inline so the owning public
  workflow and its status/error behaviour remain visible.
- A helper whose first statement is `return` is presumed to be an unnecessary
  wrapper. It may remain only when it is a meaningful global method of a
  reusable package or component and has multiple external callers.
- A helper that does more work may still be unnecessary when it only adapts an
  externally supplied object into another method on the same component's own
  properties. Use the owned property directly. Keep a helper only when it
  contains a substantial algorithm, hides complexity from the primary
  workflow, or is reused by at least two methods for the same named concept.
- Helpers must be declared immediately after the constructor and before the
  main methods. If a file has no constructor, declare them before all main
  methods. Do not place a private helper between public methods or after the
  method that calls it.
- The allowed exception is a meaningful global method on a reusable package or
  component that is composed by multiple external callers or embedded as a
  property by another component. Context packages are an example: their
  methods provide a complete, clean package boundary rather than saving a few
  lines. Large shared visualization logic may likewise use a global
  `visualize...()` method when multiple visualization service methods delegate
  to it.
- Do not add a helper merely to make a singular method call a one-item private
  implementation of its plural counterpart. When the workflows are genuinely
  equivalent, the singular public method may call the plural public method with
  one item; otherwise keep each workflow explicit.

## Batch Database Operations (Mandatory)

**Per-row database operations are strictly forbidden.** A `for` loop may only normalize input, build sets/maps, assemble batch inputs/placeholders, or construct a response; it must not perform any database operation, including:

- raw SQL：`Exec`、`Raw`、`Scan`。
- GORM queries/mutations: `Model`, `Create`, `Updates`, `Update`, `Delete`, `Find`, `First`, `Count`, `Pluck`, and so on.
- Repository methods, other service methods, or any helper that indirectly executes SQL.

This rule is not relaxed for small data sets. If a requirement appears to require per-row work, first change it to a batch interface, collection query, or set-based SQL operation; do not leave an N+1 implementation in the function. When database parameter or statement-size limits apply, split the work into fixed-size batches, but each batch must still be one set operation rather than one query per row.

Use the following order of preference:

1. For one read, use `IN ?`, joins, preloads, CTEs, or an existing bulk repository method, then match the data in memory with a map/set.
2. For batch inserts, use `CreateInBatches` or an existing `CreateMany`/bulk repository method.
3. For batch updates, upserts, or relation changes, assemble `valuePlaceholders` and `valueArgs` in a loop, then execute one `VALUES`/CTE statement; every value must be a bound parameter, never a user-data string interpolated into SQL.
4. When SQL must preserve input order or return per-row results, include an index in `VALUES`, then map one `RETURNING`/`Scan` result in memory.

```go
valuePlaceholders := make([]string, 0, len(bulkInputs))
valueArgs := make([]any, 0, len(bulkInputs)*2)
for _, input := range bulkInputs {
	valuePlaceholders = append(valuePlaceholders, "(?::uuid, ?::text)")
	valueArgs = append(valueArgs, input.Id, input.Name)
}

sql := fmt.Sprintf(`
	UPDATE "StationTable" AS station
	SET name = value.name
	FROM (VALUES %s) AS value(id, name)
	WHERE station.id = value.id::uuid
`, strings.Join(valuePlaceholders, ","))
result := tx.Exec(sql, valueArgs...)
```

Handle empty input first to avoid invalid `VALUES` statements. Handle the `Error`, `RowsAffected`, and transaction completion of a single bulk query according to the existing exception and transaction conventions.

## Shared Libraries

- Before adding a helper, check [shared/lib](../../shared/lib/) for a package that already provides the responsibility. Reuse an existing implementation instead of copying it into a service/repository.
- Choose the existing package for the problem: use `shared/lib/array` for deduplication/sets, `shared/lib/searchcursor` for cursor pagination, `shared/lib/concurrency` for concurrent work, and `shared/lib/queue`/`shared/lib/stack` for queues/stacks. Flatten EditableBlock trees with `shared/util/editableblock.FlattenEditableBlock(s)`; use `shared/util/responsewriter` and `shared/util/exceptionwriter` for cross-runtime HTTP response formatting and public exception rendering. Rate limiting remains owned by each Gateway runtime.
- `shared/lib` contains only cross-domain, reusable logic independent of the application layer. It must not import Notegic project code; required third-party libraries are allowed. Business rules used by one domain stay in that domain instead of moving to shared for possible future reuse.
- When an existing library is close but not an exact fit, add the smallest generic capability to that library. If the need belongs to one domain, use a small domain-local helper and avoid creating a one-off shared package.

## Change Scope

- A feature change modifies only the necessary layers and their tests; do not reformat unrelated files or perform broad renames as cleanup.
- Existing uncommitted user changes are outside the current scope unless explicitly requested or directly conflicting with the files being changed.
- New settings, environment variables, API fields, database columns, and event formats are contracts; update the corresponding documentation and consumers in the same change.
