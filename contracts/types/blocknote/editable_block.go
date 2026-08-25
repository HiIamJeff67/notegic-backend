package blocknote

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type ArborizedEditableBlock struct {
	Id       uuid.UUID                `json:"id" validate:"required"`
	Type     cenums.BlockType         `json:"type" validate:"required,isblocktype"`
	Props    BlockProps               `json:"-"`
	Content  BlockContent             `json:"-"`
	Children []ArborizedEditableBlock `json:"children" validate:"omitempty"`
}

func (block *ArborizedEditableBlock) UnmarshalJSON(data []byte) error {
	type alias ArborizedEditableBlock

	input := &struct {
		Props   json.RawMessage `json:"props"`
		Content json.RawMessage `json:"content"`
		*alias
	}{
		alias: (*alias)(block),
	}
	if err := json.Unmarshal(data, input); err != nil {
		return err
	}

	props, err := ParseProps(string(block.Type), input.Props)
	if err != nil {
		return err
	}
	block.Props = props

	content := bytes.TrimSpace(input.Content)
	if len(content) == 0 || string(content) == "null" {
		return nil
	}

	switch content[0] {
	case '[':
		var inlineContent InlineContentList
		if err := json.Unmarshal(content, &inlineContent); err != nil {
			return err
		}
		block.Content = inlineContent
	case '{':
		var tableContent TableContent
		if err := json.Unmarshal(content, &tableContent); err != nil {
			return err
		}
		block.Content = &tableContent
	case '"':
		var plainContent PlainContent
		if err := json.Unmarshal(content, &plainContent); err != nil {
			return err
		}
		block.Content = plainContent
	default:
		return errors.New("invalid content format: must be array, object, or string")
	}

	return nil
}

func (block ArborizedEditableBlock) MarshalJSON() ([]byte, error) {
	type alias ArborizedEditableBlock

	return json.Marshal(&struct {
		Props   BlockProps   `json:"props"`
		Content BlockContent `json:"content"`
		*alias
	}{
		Props:   block.Props,
		Content: block.Content,
		alias:   (*alias)(&block),
	})
}

type RawFlattenedEditableBlock struct {
	Id            uuid.UUID        `json:"id"`
	ParentBlockId *uuid.UUID       `json:"parentBlockId"`
	PrevBlockId   *uuid.UUID       `json:"prevBlockId"`
	NextBlockId   *uuid.UUID       `json:"nextBlockId"`
	Type          cenums.BlockType `json:"type"`
	Props         json.RawMessage  `json:"props"`
	Content       json.RawMessage  `json:"content"`
}
