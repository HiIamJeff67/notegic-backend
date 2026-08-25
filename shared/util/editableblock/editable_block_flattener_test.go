package editableblock

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
)

func TestFlattenEditableBlocksPreservesTreeRelationships(t *testing.T) {
	firstRootId := uuid.New()
	secondRootId := uuid.New()
	firstChildId := uuid.New()
	secondChildId := uuid.New()

	payload := []byte(`[
		{
			"id": "` + firstRootId.String() + `",
			"type": "paragraph",
			"props": {"backgroundColor":"default","textColor":"default","textAlignment":"left"},
			"content": [],
			"children": [
				{"id": "` + firstChildId.String() + `", "type": "paragraph", "props": {"backgroundColor":"default","textColor":"default","textAlignment":"left"}, "content": [], "children": []},
				{"id": "` + secondChildId.String() + `", "type": "paragraph", "props": {"backgroundColor":"default","textColor":"default","textAlignment":"left"}, "content": [], "children": []}
			]
		},
		{"id": "` + secondRootId.String() + `", "type": "paragraph", "props": {"backgroundColor":"default","textColor":"default","textAlignment":"left"}, "content": [], "children": []}
	]`)

	var roots []cblocknote.ArborizedEditableBlock
	if err := json.Unmarshal(payload, &roots); err != nil {
		t.Fatalf("unmarshal arborized blocks: %v", err)
	}

	blocks, _, err := FlattenEditableBlocks(roots)
	if err != nil {
		t.Fatalf("flatten arborized blocks: %v", err)
	}
	if len(blocks) != 4 {
		t.Fatalf("expected 4 flattened blocks, got %d", len(blocks))
	}

	blocksById := make(map[uuid.UUID]cblocknote.RawFlattenedEditableBlock, len(blocks))
	for _, block := range blocks {
		blocksById[block.Id] = block
	}

	firstRoot := blocksById[firstRootId]
	if firstRoot.ParentBlockId != nil || firstRoot.PrevBlockId != nil ||
		firstRoot.NextBlockId == nil || *firstRoot.NextBlockId != secondRootId {
		t.Fatalf("unexpected first root pointers: %#v", firstRoot)
	}

	firstChild := blocksById[firstChildId]
	if firstChild.ParentBlockId == nil || *firstChild.ParentBlockId != firstRootId ||
		firstChild.PrevBlockId != nil || firstChild.NextBlockId == nil ||
		*firstChild.NextBlockId != secondChildId {
		t.Fatalf("unexpected first child pointers: %#v", firstChild)
	}
}
