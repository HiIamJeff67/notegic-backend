# Service and Persistence Conventions

## Service

- A service is the business workflow and transaction boundary. Validate every public workflow with `validation.Validator.Struct(request)` first, then obtain a request-scoped DB with `s.db.WithContext(ctx)`.
- A service returns a response DTO or necessary domain data with `*exceptions.Exception`; it must not return `gin.Context`, an HTTP status, or `gin.H`.
- Cache, email, token, storage, and realtime operations must have an explicit order in the service workflow. Failure compensation and retry semantics must remain consistent with database commit semantics.

## Service Method Blocks and Spacing

A service method should read as adjacent statements within one semantic block, with one blank line between separate stages, rather than being divided by line count or a fixed template. A common order is: request validation → preconditions/preparation → DB session or transaction → read/write workflow → commit → response. Do not add empty stages or comments for phases that do not occur.

- A single operation and its immediate error handling are one semantic block: place `if exception != nil` or `if err != nil` immediately after the repository/DB call without a blank line.
- Put the validation guard at the top of the method; after the guard, leave one blank line before the next independent stage. The blank line may be omitted only when validation is immediately followed by a precondition in the same semantic stage.
- `db := s.db.WithContext(ctx)` or `tx := s.db.WithContext(ctx).Begin()` is the workflow execution environment and should form its own block, separated by one blank line from the previous stage and the first query/workflow after it.
- Derived values, input assembly, query-result-to-DTO mapping, and loop processing each form a continuous block; separate a change of work with one blank line. Do not insert blank lines for visual effect within one block.
- Ordinary methods use blank lines to express stages. Use `sep30` only when one file contains two or more clearly separate and complex method families (for example helper, HTTP service, chart service, or GraphQL service). When a file contains only main service methods, do not add a `Services for Something` separator above them:

  ```go
  /* ============================== Service Methods for GraphQL Station ============================== */
  ```

  `sep30` is for navigating multiple file-level families. Do not use it to separate steps within one method, individual `if` statements, or struct fields, and do not use it to title the only main method family; a service file with one method family does not need it.

## Service Validation and DB Query Format

Every service method that receives a request validates it first and maps the validator error to a domain exception. Only then does it create the request-scoped DB; these are separate stages, so keep the blank line between them.

```go
func (s *StationService) CreateStation(
	ctx context.Context,
	request *CreateStationRequest,
) (*CreateStationResponse, *exceptions.Exception) {
	if err := validation.Validator.Struct(request); err != nil {
		return nil, apiexceptions.Station.InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	newStationId, exception := s.stationRepository.CreateOne(
		request.ContextFields.UserId,
		input,
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &CreateStationResponse{
		Id: *newStationId, 
	}, nil
}
```

Break a GORM chain after more than one operation; keep the receiver's `.` at the end of the line so each `Model`, `Where`, `Order`, `Find`, or `Updates` is a clear step. Query construction and execution are one block; check the result/error immediately afterward without an intervening blank line.

```go
var blocks []schemas.Block
if err := db.Model(&schemas.Block{}).
	Where("block_pack_id = ?", request.Param.BlockPackId).
	Order("created_at ASC").
	Order("id ASC").
	Find(&blocks).Error; err != nil {
	return nil, apiexceptions.Block.NotFound().WithOrigin(err)
}

result := tx.
	Model(&schemas.Station{}).
	Where("id = ?", station.Id).
	Update("deleted_at", time.Now())
if result.Error != nil {
	tx.Rollback()
	return nil, apiexceptions.Station.FailedToUpdate().WithOrigin(result.Error)
}
```

Short, single DB operations may stay on one line; do not split a simple `tx.Model(&schema).Update(...)` solely to apply chain formatting.

## Service and Repository Transactions

When multiple reads and writes must succeed or fail together, the service opens the transaction and is solely responsible for `Commit`/`Rollback`. Every repository call uses the same `tx` and `options.WithTransactionDB(tx)`; that option carries the DB and marks the transaction as started so the repository does not open a nested transaction because `IsTransactionStarted` is false.

Create `tx` as an independent block. If `Begin()` returns an error, there is no transaction to roll back, so return immediately. After the transaction starts, every failure's rollback and return form one adjacent error-closing block with no blank line or other code between them.

```go
tx := s.db.WithContext(ctx).Begin()
if err := tx.Error; err != nil {
	return nil, apiexceptions.Shelf.FailedToCommitTransaction("failed to begin transaction").WithOrigin(err)
}

rootShelf, exception := s.rootShelfRepository.CheckPermissionAndGetOneById(
	request.Body.RootShelfId,
	request.ContextFields.UserId,
	nil,
	allowedPermissions,
	options.WithTransactionDB(tx),
	options.WithAllowedPermissions(allowedPermissions),
	options.WithOnlyDeleted(types.Ternary_Negative),
	options.WithLockingStrength(options.LockingStrengthUpdate),
)
if exception != nil {
	tx.Rollback()
	return nil, exception
}

// All remaining reads and writes use tx or options.WithTransactionDB(tx).

if err := tx.Commit().Error; err != nil {
	tx.Rollback()
	return nil, apiexceptions.Shelf.FailedToCommitTransaction().WithOrigin(err)
}
```

When a single repository method already owns a complete atomic workflow, keep its existing behavior of opening and committing its own transaction; a service must not mix another `s.db` or transaction inside an outer transaction. New workflows spanning repositories are always managed by an outer service transaction.

## Repository and Scope

- Repositories centralize GORM/raw SQL access and use action-oriented public names such as `GetOneById`, `CreateMany`, and `UpdateOneById`.
- Every repository ORM query must begin with `parsedOptions.DB.Model(&schemas.Xxx{})`, using the schema that owns the operation. Do not use `db.Table(...)` or `parsedOptions.DB.Table(...)`: it bypasses schema metadata and can produce incorrect model/table behavior. Raw SQL may name its required quoted tables explicitly, but its GORM session must still start from the owning `Model(&schemas.Xxx{})`.
- Name parsed repository options `parsedOptions`, never a shortened `parsed`. Treat options parsing as preparation: leave one blank line before the query/result block. Keep a query result and its immediate error handling adjacent; use one blank line only when moving to a separate preparation, query, mapping, or return stage.
- Use `schemas.Xxx` models and existing scopes to encapsulate permission, preload, soft-delete, and locking; do not repeat access-control `Where` clauses in services/controllers.
- Pass repository options through existing options such as `options.WithDB`, `WithAllowedPermissions`, `WithOnlyDeleted`, and `WithLockingStrength`. When `WithAllowedPermissions` is present, apply the validated Gateway route policy; when it is absent, the repository operates directly and does not use a skip-permission flag. `HasPermission`, `HavePermissions`, and `CheckPermission...` methods must still receive `allowedPermissions` explicitly when required, and the service call must also pass `options.WithAllowedPermissions(allowedPermissions)`.
- Use create/update/partial-update types from `shared/platform/postgres/repositories/inputs`. Do not pass a request directly to GORM. All repository inputs are shared persistence contracts; runtime-specific workflow DTOs remain in the runtime service layer.
- Map GORM result errors to the appropriate domain exception while preserving the origin. After `First` or `Find`, also handle the domain meaning of an empty result rather than checking only `result.Error`.

## Repository Partial Updates

When an operation supports partial updates or explicitly setting `NULL`, use the existing partial-update flow; do not construct a map manually or pass a request directly to `Updates`.

1. Define `UpdateXxxInput` under `shared/platform/postgres/repositories/inputs/<domain>_input.go`. Use pointers to distinguish values supplied by this request and keep the correct `json` and `gorm:"column:..."` tags.
2. Connect the input to `partial_update_input.go` under the same data owner: `type PartialUpdateXxxInput = PartialUpdateInput[UpdateXxxInput]`. `Values` contains values to overwrite, and `SetNull` identifies fields that must become `NULL`.
3. Within the same transaction, the repository first gets the existing schema after permission checks, validates related resources/ownership, then calls `util.PartialUpdatePreprocess(input.Values, input.SetNull, *existing)`.
4. Write the merged result with `Select("*").Updates(&updates)`. Because the processor preserves existing values for fields not supplied, `Select("*")` is required to write explicit zero values or `NULL`.

```go
type UpdateStationInput struct {
	Name        *string `json:"name" gorm:"column:name;"`
	Description *string `json:"description" gorm:"column:description;"`
}

type PartialUpdateStationInput = PartialUpdateInput[UpdateStationInput]
```

```go
updates, err := util.PartialUpdatePreprocess(input.Values, input.SetNull, *existingStation)
if err != nil {
	parsedOptions.DB.Rollback()
	return nil, exceptions.New(
		"FailedToPreprocessPartialUpdate",
		"Repository",
		"Update",
		"Failed to preprocess partial update",
		http.StatusInternalServerError,
		true,
	).WithOrigin(err)
}

result := parsedOptions.DB.Model(&schemas.Station{}).
	Where("id = ? AND deleted_at IS NULL", id).
	Select("*").
	Updates(&updates)
```

- `SetNull` keys use the corresponding Go field name. The processor handles case and underscore differences, but new API/DTO contracts should still use the established camelCase form.
- Validate related resources before a partial update only when that field has a new value and is not marked in `SetNull`; for example, check destination permission before moving a parent resource.
- Handle processor errors, DB errors, and `RowsAffected == 0` with domain exceptions and `exceptions.Cover()` according to the [exception conventions](06-exceptions.md); in a repository that owns the transaction, rollback and return remain adjacent.

## Schema, Migration, and SQL

- Runtime-owned exception builders belong under each runtime's `internal/<runtime>/exceptions/`. PostgreSQL repository exceptions are split by domain under `shared/platform/postgres/repositories/exceptions/`, injected into each repository through its `exceptions` field, and used as `r.exceptions...`. All PostgreSQL GORM table schemas, relations, table names, shared DDL packages, repository implementations, repository inputs, and repository exceptions belong under `shared/platform/postgres/`. Runtimes import these persistence contracts directly and must not duplicate them. Every repository constructor receives the owning runtime's DB pool; shared PostgreSQL code must not keep a package-global default DB or import a runtime. Runtime-specific business methods stay in services/workers. Runtime-owned enums remain under `internal/<service>/data/postgres/schemas/enums/` because their PostgreSQL scan/value behavior is runtime persistence code. Contract enum values remain under `contracts/types/enums/`.
- Every PostgreSQL runtime opens its own pool in the application composition root with `spostgres.Connect` and closes it during shutdown. Migration belongs to an explicit deployment/admin command, not application startup or `Connect`.
- The default pool is runtime-owned if a runtime needs one: `shared/platform/postgres/repositories` must not keep a package-global DB or `SetDefaultDB`; each runtime composition root constructs shared repository instances with its own pool, while `WithDB` and `WithTransactionDB` override it for request and transaction scope.
- All reusable PostgreSQL scopes belong under `shared/platform/postgres/scopes/`; scopes are not runtime-owned. Migration DDL belongs beside the operation it migrates under `shared/platform/postgres/schemas/views/`, `schemas/triggers/`, or `schemas/constraints/`. Runtime-only raw query helpers belong under that runtime's `data/postgres/sqls/`; they are not shared platform assets.
- Each database-owning runtime declares one `migration_manifest.go` under `internal/<service>/data/postgres/`. Its `MigrationManifest.Runtime` uses `contracts/types.Runtime` and contains the inline enum map and shared schema/DDL registrations handled by that runtime. The runtime's `migrateDB` deployment/admin command uses `postgres.Migrate()` through the owner role, then applies `DatabasePermissionManifest` with `postgres.ApplyPermissions()` through the admin connection. Application startup must not bootstrap roles, migrate, or reconcile permissions.
- For database invariants such as soft-delete, ownership, projection, and accounting, prefer the existing trigger/constraint/scope patterns; do not rely only on controller checks.
- When changing a trigger or raw SQL, verify that table names, columns, and migration registration are updated together. Database-facing semantic changes must also update the corresponding `docs/codebase-design/`, `docs/api-route-design/`, or `docs/system-design/` document.
