package repositories

import (
	"fmt"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"strings"
	"time"

	"github.com/google/uuid"

	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type BlockPackYjsRepositoryInterface interface {
	LoadDocumentAndUpdates(blockPackId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.BlockPackYjsDocument, []schemas.BlockPackYjsUpdate, error)
	AppendUpdate(blockPackId uuid.UUID, input inputs.AppendBlockPackYjsUpdateInput, opts ...options.RepositoryOptions) (int64, error)
	GetCompactableYjsDocumentWithUpdates(blockPackId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.BlockPackYjsDocument, []schemas.BlockPackYjsUpdate, error)
	ApplyCompactedYjsDocument(blockPackId uuid.UUID, input inputs.ApplyCompactedBlockPackYjsDocumentInput, opts ...options.RepositoryOptions) (bool, error)
	DeleteCompactedUpdates(input inputs.DeleteCompactedBlockPackYjsUpdatesInput, opts ...options.RepositoryOptions) (int64, error)

	BulkCheckPermissionsAndGetManyByIds(blockPackIds []uuid.UUID, opts ...options.RepositoryOptions) ([]bool, []BlockPackYjsDocumentWithUpdates, error)
	BulkCheckProjectionDocumentsByIds(blockPackIds []uuid.UUID, opts ...options.RepositoryOptions) ([]BlockPackYjsDocumentWithUpdates, error)
	BulkApplyCompactedYjsDocuments(inputs []inputs.BulkApplyCompactedBlockPackYjsDocumentInput, opts ...options.RepositoryOptions) ([]uuid.UUID, error)
}

type BlockPackYjsRepository struct{}

type BlockPackYjsDocumentWithUpdates struct {
	Document schemas.BlockPackYjsDocument
	Updates  []schemas.BlockPackYjsUpdate
}

func NewBlockPackYjsRepository() BlockPackYjsRepositoryInterface {
	return &BlockPackYjsRepository{}
}

func (r *BlockPackYjsRepository) LoadDocumentAndUpdates(
	blockPackId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*schemas.BlockPackYjsDocument, []schemas.BlockPackYjsUpdate, error) {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	db := parsedOptions.DB

	var blockPack schemas.BlockPack
	if err := db.Model(&schemas.BlockPack{}).
		Select("id").
		Where("id = ? AND deleted_at IS NULL", blockPackId).
		First(&blockPack).Error; err != nil {
		return nil, nil, err
	}

	var document schemas.BlockPackYjsDocument
	if err := db.Model(&schemas.BlockPackYjsDocument{}).
		Where("block_pack_id = ? AND deleted_at IS NULL", blockPackId).
		First(&document).Error; err != nil {
		return nil, nil, err
	}

	var updates []schemas.BlockPackYjsUpdate
	if err := db.Model(&schemas.BlockPackYjsUpdate{}).
		Where("block_pack_id = ?", blockPackId).
		Where("update_sequence > ? AND update_sequence <= ?", document.CompactedUntilSequence, document.LastUpdateSequence).
		Order("update_sequence ASC").
		Find(&updates).Error; err != nil {
		return nil, nil, err
	}

	return &document, updates, nil
}

func (r *BlockPackYjsRepository) AppendUpdate(
	blockPackId uuid.UUID,
	input inputs.AppendBlockPackYjsUpdateInput,
	opts ...options.RepositoryOptions,
) (int64, error) {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	tx := parsedOptions.DB

	lockingStrength := "UPDATE"
	var blockPack schemas.BlockPack
	if err := tx.Model(&schemas.BlockPack{}).
		Select("id").
		Scopes(scopes.Locking(&lockingStrength)).
		Where("id = ? AND deleted_at IS NULL", blockPackId).
		First(&blockPack).Error; err != nil {
		if shouldStartTransaction {
			tx.Rollback()
		}

		return 0, err
	}

	var document schemas.BlockPackYjsDocument
	if err := tx.Model(&schemas.BlockPackYjsDocument{}).
		Scopes(scopes.Locking(&lockingStrength)).
		Where("block_pack_id = ? AND deleted_at IS NULL", blockPackId).
		First(&document).Error; err != nil {
		if shouldStartTransaction {
			tx.Rollback()
		}

		return 0, err
	}

	var existingUpdate schemas.BlockPackYjsUpdate
	existingUpdateResult := tx.Model(&schemas.BlockPackYjsUpdate{}).
		Select("update_sequence").
		Where("block_pack_id = ? AND persistence_batch_id = ?", blockPackId, input.PersistenceBatchId).
		Limit(1).
		Find(&existingUpdate)
	if existingUpdateResult.Error != nil {
		if shouldStartTransaction {
			tx.Rollback()
		}

		return 0, existingUpdateResult.Error
	}
	if existingUpdateResult.RowsAffected > 0 {
		if shouldStartTransaction {
			if err := tx.Commit().Error; err != nil {
				return 0, err
			}
		}

		return existingUpdate.UpdateSequence, nil
	}

	nextUpdateSequence := document.LastUpdateSequence + 1
	update := schemas.BlockPackYjsUpdate{
		BlockPackId:        blockPackId,
		UpdateSequence:     nextUpdateSequence,
		PersistenceBatchId: input.PersistenceBatchId,
		Payload:            input.Payload,
		OriginConnectionId: input.OriginConnectionId,
	}
	if err := tx.Create(&update).Error; err != nil {
		if shouldStartTransaction {
			tx.Rollback()
		}

		return 0, err
	}

	if err := tx.Model(&schemas.BlockPackYjsDocument{}).
		Where("id = ?", document.Id).
		Update("last_update_sequence", nextUpdateSequence).Error; err != nil {
		if shouldStartTransaction {
			tx.Rollback()
		}

		return 0, err
	}

	if shouldStartTransaction {
		if err := tx.Commit().Error; err != nil {
			return 0, err
		}
	}

	return nextUpdateSequence, nil
}

func (r *BlockPackYjsRepository) GetCompactableYjsDocumentWithUpdates(
	blockPackId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*schemas.BlockPackYjsDocument, []schemas.BlockPackYjsUpdate, error) {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	tx := parsedOptions.DB

	lockingStrength := "UPDATE"
	var document schemas.BlockPackYjsDocument
	if err := tx.Model(&schemas.BlockPackYjsDocument{}).
		Scopes(scopes.Locking(&lockingStrength)).
		Where("block_pack_id = ? AND deleted_at IS NULL", blockPackId).
		First(&document).Error; err != nil {
		if shouldStartTransaction {
			tx.Rollback()
		}

		return nil, nil, err
	}

	var updates []schemas.BlockPackYjsUpdate
	if err := tx.Model(&schemas.BlockPackYjsUpdate{}).
		Where("block_pack_id = ?", blockPackId).
		Where("update_sequence > ? AND update_sequence <= ?", document.CompactedUntilSequence, document.LastUpdateSequence).
		Order("update_sequence ASC").
		Find(&updates).Error; err != nil {
		if shouldStartTransaction {
			tx.Rollback()
		}

		return nil, nil, err
	}

	if shouldStartTransaction {
		if err := tx.Commit().Error; err != nil {
			return nil, nil, err
		}
	}

	return &document, updates, nil
}

func (r *BlockPackYjsRepository) ApplyCompactedYjsDocument(
	blockPackId uuid.UUID,
	input inputs.ApplyCompactedBlockPackYjsDocumentInput,
	opts ...options.RepositoryOptions,
) (bool, error) {
	if input.CutoffSequence <= input.BaseCompactedUntilSequence {
		return false, fmt.Errorf("yjs compaction cutoff must advance the document checkpoint")
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	tx := parsedOptions.DB

	lockingStrength := "UPDATE"
	var document schemas.BlockPackYjsDocument
	if err := tx.Model(&schemas.BlockPackYjsDocument{}).
		Scopes(scopes.Locking(&lockingStrength)).
		Where("block_pack_id = ? AND deleted_at IS NULL", blockPackId).
		First(&document).Error; err != nil {
		if shouldStartTransaction {
			tx.Rollback()
		}

		return false, err
	}

	if document.CompactedUntilSequence != input.BaseCompactedUntilSequence {
		if shouldStartTransaction {
			if err := tx.Commit().Error; err != nil {
				return false, err
			}
		}

		return false, nil
	}
	if input.CutoffSequence > document.LastUpdateSequence {
		if shouldStartTransaction {
			tx.Rollback()
		}

		return false, fmt.Errorf("yjs compaction cutoff exceeds durable update sequence")
	}

	now := time.Now()
	if err := tx.Model(&schemas.BlockPackYjsDocument{}).
		Where("id = ?", document.Id).
		Updates(map[string]any{
			"snapshot":                 input.Snapshot,
			"state_vector":             input.StateVector,
			"compacted_until_sequence": input.CutoffSequence,
			"last_compacted_at":        now,
			"updated_at":               now,
		}).Error; err != nil {
		if shouldStartTransaction {
			tx.Rollback()
		}

		return false, err
	}

	if err := tx.Model(&schemas.BlockPackYjsUpdate{}).
		Where("block_pack_id = ?", blockPackId).
		Where("update_sequence > ? AND update_sequence <= ?", input.BaseCompactedUntilSequence, input.CutoffSequence).
		Where("compacted_at IS NULL").
		Update("compacted_at", now).Error; err != nil {
		if shouldStartTransaction {
			tx.Rollback()
		}

		return false, err
	}

	if shouldStartTransaction {
		if err := tx.Commit().Error; err != nil {
			return false, err
		}
	}

	return true, nil
}

// BulkCheckPermissionsAndGetManyByIds is system-only. A valid target is an active
// BlockPack with an active Yjs document and a contiguous durable update tail.
func (r *BlockPackYjsRepository) BulkCheckPermissionsAndGetManyByIds(
	blockPackIds []uuid.UUID,
	opts ...options.RepositoryOptions,
) ([]bool, []BlockPackYjsDocumentWithUpdates, error) {
	if len(blockPackIds) == 0 {
		return []bool{}, []BlockPackYjsDocumentWithUpdates{}, nil
	}
	parsedOptions := options.ParseRepositoryOptions(opts...)
	db := parsedOptions.DB

	validBlockPacks := make([]bool, len(blockPackIds))
	blockPackIdSet := make(map[uuid.UUID]bool, len(blockPackIds))
	for _, blockPackId := range blockPackIds {
		if blockPackId != uuid.Nil {
			blockPackIdSet[blockPackId] = true
		}
	}

	validBlockPackIds := make([]uuid.UUID, 0, len(blockPackIdSet))
	for blockPackId := range blockPackIdSet {
		validBlockPackIds = append(validBlockPackIds, blockPackId)
	}
	if len(validBlockPackIds) == 0 {
		return validBlockPacks, []BlockPackYjsDocumentWithUpdates{}, nil
	}

	var documents []schemas.BlockPackYjsDocument
	if err := db.Model(&schemas.BlockPackYjsDocument{}).
		Select(`"BlockPackYjsDocumentTable".*`).
		Joins(`INNER JOIN "BlockPackTable" bp ON bp.id = "BlockPackYjsDocumentTable".block_pack_id AND bp.deleted_at IS NULL`).
		Where(`"BlockPackYjsDocumentTable".block_pack_id IN ?`, validBlockPackIds).
		Where(`"BlockPackYjsDocumentTable".deleted_at IS NULL`).
		Where(`"BlockPackYjsDocumentTable".last_update_sequence > "BlockPackYjsDocumentTable".compacted_until_sequence`).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&documents).Error; err != nil {
		return nil, nil, err
	}
	if len(documents) == 0 {
		return validBlockPacks, []BlockPackYjsDocumentWithUpdates{}, nil
	}

	validDocumentByBlockPackId := make(map[uuid.UUID]schemas.BlockPackYjsDocument, len(documents))
	for _, document := range documents {
		validDocumentByBlockPackId[document.BlockPackId] = document
	}

	var updates []schemas.BlockPackYjsUpdate
	if err := db.Model(&schemas.BlockPackYjsUpdate{}).
		Where("block_pack_id IN ?", validBlockPackIds).
		Where("compacted_at IS NULL").
		Order("block_pack_id ASC").
		Order("update_sequence ASC").
		Find(&updates).Error; err != nil {
		return nil, nil, err
	}

	updatesByBlockPackId := make(map[uuid.UUID][]schemas.BlockPackYjsUpdate, len(documents))
	for _, update := range updates {
		document, exists := validDocumentByBlockPackId[update.BlockPackId]
		if !exists || update.UpdateSequence <= document.CompactedUntilSequence || update.UpdateSequence > document.LastUpdateSequence {
			continue
		}

		updatesByBlockPackId[update.BlockPackId] = append(updatesByBlockPackId[update.BlockPackId], update)
	}

	pairsByBlockPackId := make(map[uuid.UUID]BlockPackYjsDocumentWithUpdates, len(documents))
	for _, document := range documents {
		updates := updatesByBlockPackId[document.BlockPackId]
		expectedSequence := document.CompactedUntilSequence + 1
		for _, update := range updates {
			if update.UpdateSequence != expectedSequence {
				updates = nil
				break
			}

			expectedSequence++
		}
		if len(updates) == 0 || expectedSequence-1 != document.LastUpdateSequence {
			continue
		}

		pairsByBlockPackId[document.BlockPackId] = BlockPackYjsDocumentWithUpdates{
			Document: document,
			Updates:  updates,
		}
	}

	pairs := make([]BlockPackYjsDocumentWithUpdates, 0, len(pairsByBlockPackId))
	for index, blockPackId := range blockPackIds {
		pair, exists := pairsByBlockPackId[blockPackId]
		if !exists {
			continue
		}

		validBlockPacks[index] = true
		pairs = append(pairs, pair)
		delete(pairsByBlockPackId, blockPackId)
	}

	return validBlockPacks, pairs, nil
}

func (r *BlockPackYjsRepository) BulkCheckProjectionDocumentsByIds(
	blockPackIds []uuid.UUID,
	opts ...options.RepositoryOptions,
) ([]BlockPackYjsDocumentWithUpdates, error) {
	if len(blockPackIds) == 0 {
		return []BlockPackYjsDocumentWithUpdates{}, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	db := parsedOptions.DB

	var documents []schemas.BlockPackYjsDocument
	if err := db.Model(&schemas.BlockPackYjsDocument{}).
		Select(`"BlockPackYjsDocumentTable".*`).
		Joins(`INNER JOIN "BlockPackTable" bp ON bp.id = "BlockPackYjsDocumentTable".block_pack_id AND bp.deleted_at IS NULL`).
		Where(`"BlockPackYjsDocumentTable".block_pack_id IN ?`, blockPackIds).
		Where(`"BlockPackYjsDocumentTable".deleted_at IS NULL`).
		Where(`"BlockPackYjsDocumentTable".last_update_sequence > "BlockPackYjsDocumentTable".projected_until_sequence`).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&documents).Error; err != nil {
		return nil, err
	}
	if len(documents) == 0 {
		return []BlockPackYjsDocumentWithUpdates{}, nil
	}

	documentByBlockPackId := make(map[uuid.UUID]schemas.BlockPackYjsDocument, len(documents))
	for _, document := range documents {
		documentByBlockPackId[document.BlockPackId] = document
	}

	var updates []schemas.BlockPackYjsUpdate
	if err := db.Model(&schemas.BlockPackYjsUpdate{}).
		Joins(`INNER JOIN "BlockPackYjsDocumentTable" document ON document.block_pack_id = "BlockPackYjsUpdateTable".block_pack_id AND document.deleted_at IS NULL`).
		Where(`"BlockPackYjsUpdateTable".block_pack_id IN ?`, blockPackIds).
		Where(`"BlockPackYjsUpdateTable".update_sequence > document.compacted_until_sequence`).
		Where(`"BlockPackYjsUpdateTable".update_sequence <= document.last_update_sequence`).
		Order(`"BlockPackYjsUpdateTable".block_pack_id ASC`).
		Order(`"BlockPackYjsUpdateTable".update_sequence ASC`).
		Find(&updates).Error; err != nil {
		return nil, err
	}

	updatesByBlockPackId := make(map[uuid.UUID][]schemas.BlockPackYjsUpdate, len(documents))
	for _, update := range updates {
		if _, exists := documentByBlockPackId[update.BlockPackId]; exists {
			updatesByBlockPackId[update.BlockPackId] = append(updatesByBlockPackId[update.BlockPackId], update)
		}
	}

	pairs := make([]BlockPackYjsDocumentWithUpdates, 0, len(documents))
	for _, blockPackId := range blockPackIds {
		document, exists := documentByBlockPackId[blockPackId]
		if !exists {
			continue
		}

		updates := updatesByBlockPackId[blockPackId]
		expectedSequence := document.CompactedUntilSequence + 1
		for _, update := range updates {
			if update.UpdateSequence != expectedSequence {
				updates = nil
				break
			}
			expectedSequence++
		}
		if expectedSequence-1 != document.LastUpdateSequence {
			continue
		}

		pairs = append(pairs, BlockPackYjsDocumentWithUpdates{
			Document: document,
			Updates:  updates,
		})
	}

	return pairs, nil
}

func (r *BlockPackYjsRepository) BulkApplyCompactedYjsDocuments(
	bulkInputs []inputs.BulkApplyCompactedBlockPackYjsDocumentInput,
	opts ...options.RepositoryOptions,
) ([]uuid.UUID, error) {
	if len(bulkInputs) == 0 {
		return []uuid.UUID{}, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	db := parsedOptions.DB

	valueRows := make([]string, 0, len(bulkInputs))
	args := make([]any, 0, len(bulkInputs)*5)
	for _, input := range bulkInputs {
		if input.BlockPackId == uuid.Nil || input.CutoffSequence <= input.BaseCompactedUntilSequence {
			return nil, fmt.Errorf("invalid compacted yjs document input")
		}

		valueRows = append(valueRows, "(?::uuid, ?::bigint, ?::bigint, ?::bytea, ?::bytea)")
		args = append(
			args,
			input.BlockPackId,
			input.BaseCompactedUntilSequence,
			input.CutoffSequence,
			input.Snapshot,
			input.StateVector,
		)
	}

	now := time.Now()
	args = append(args, now)
	query := `
		WITH target(block_pack_id, base_compacted_until_sequence, cutoff_sequence, snapshot, state_vector) AS (
			VALUES ` + strings.Join(valueRows, ",") + `
		)
		UPDATE "BlockPackYjsDocumentTable" AS document
		SET
			snapshot = target.snapshot,
			state_vector = target.state_vector,
			compacted_until_sequence = target.cutoff_sequence,
			last_compacted_at = ?,
			updated_at = ?
		FROM target
		WHERE document.block_pack_id = target.block_pack_id
			AND document.deleted_at IS NULL
			AND document.compacted_until_sequence = target.base_compacted_until_sequence
			AND document.last_update_sequence >= target.cutoff_sequence
		RETURNING document.block_pack_id`
	args = append(args, now)

	type appliedDocument struct {
		BlockPackId uuid.UUID `gorm:"column:block_pack_id"`
	}

	var appliedDocuments []appliedDocument
	result := db.Raw(query, args...).Scan(&appliedDocuments)
	if err := result.Error; err != nil {
		return nil, err
	}

	appliedBlockPackIds := make([]uuid.UUID, len(appliedDocuments))
	for index, document := range appliedDocuments {
		appliedBlockPackIds[index] = document.BlockPackId
	}
	if len(appliedBlockPackIds) == 0 {
		return appliedBlockPackIds, nil
	}

	appliedBlockPackIdSet := make(map[uuid.UUID]bool, len(appliedBlockPackIds))
	for _, blockPackId := range appliedBlockPackIds {
		appliedBlockPackIdSet[blockPackId] = true
	}

	valueRows = valueRows[:0]
	args = make([]any, 0, len(appliedBlockPackIds)*3+1)
	for _, input := range bulkInputs {
		if !appliedBlockPackIdSet[input.BlockPackId] {
			continue
		}

		valueRows = append(valueRows, "(?::uuid, ?::bigint, ?::bigint)")
		args = append(args, input.BlockPackId, input.BaseCompactedUntilSequence, input.CutoffSequence)
	}
	args = append(args, now)
	query = `
		WITH target(block_pack_id, base_compacted_until_sequence, cutoff_sequence) AS (
			VALUES ` + strings.Join(valueRows, ",") + `
		)
		UPDATE "BlockPackYjsUpdateTable" AS yjs_update
		SET compacted_at = ?
		FROM target
		WHERE yjs_update.block_pack_id = target.block_pack_id
			AND yjs_update.update_sequence > target.base_compacted_until_sequence
			AND yjs_update.update_sequence <= target.cutoff_sequence
			AND yjs_update.compacted_at IS NULL`
	result = db.Exec(query, args...)
	if err := result.Error; err != nil {
		return nil, err
	}

	return appliedBlockPackIds, nil
}

func (r *BlockPackYjsRepository) DeleteCompactedUpdates(
	input inputs.DeleteCompactedBlockPackYjsUpdatesInput,
	opts ...options.RepositoryOptions,
) (int64, error) {
	if input.Limit <= 0 {
		return 0, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	db := parsedOptions.DB

	subQuery := db.
		Model(&schemas.BlockPackYjsUpdate{}).
		Select("id").
		Where("compacted_at IS NOT NULL AND compacted_at <= ?", input.Before).
		Order("compacted_at ASC").
		Limit(input.Limit)

	result := db.
		Where("id IN (?)", subQuery).
		Delete(&schemas.BlockPackYjsUpdate{})

	return result.RowsAffected, result.Error
}
