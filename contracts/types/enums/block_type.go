package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type BlockType string

const (
	BlockType_Paragraph BlockType = "paragraph"
	BlockType_Heading   BlockType = "heading"
	BlockType_Quote     BlockType = "quote"

	BlockType_BulletListItem   BlockType = "bulletListItem"
	BlockType_NumberedListItem BlockType = "numberedListItem"
	BlockType_CheckListItem    BlockType = "checkListItem"
	BlockType_ToggleListItem   BlockType = "toggleListItem"

	BlockType_Image BlockType = "image"
	BlockType_Video BlockType = "video"
	BlockType_Audio BlockType = "audio"
	BlockType_File  BlockType = "file"

	BlockType_Table     BlockType = "table"
	BlockType_CodeBlock BlockType = "codeBlock"
	BlockType_MathBlock BlockType = "mathBlock"
	BlockType_Diagram   BlockType = "diagram"
	BlockType_Calendar  BlockType = "calendar"
)

var AllBlockTypes = []BlockType{
	BlockType_Paragraph,
	BlockType_Heading,
	BlockType_Quote,
	BlockType_BulletListItem,
	BlockType_NumberedListItem,
	BlockType_CheckListItem,
	BlockType_ToggleListItem,
	BlockType_Image,
	BlockType_Video,
	BlockType_Audio,
	BlockType_File,
	BlockType_Table,
	BlockType_CodeBlock,
	BlockType_MathBlock,
	BlockType_Diagram,
	BlockType_Calendar,
}

var AllBlockTypeStrings = []string{
	string(BlockType_Paragraph),
	string(BlockType_Heading),
	string(BlockType_Quote),
	string(BlockType_BulletListItem),
	string(BlockType_NumberedListItem),
	string(BlockType_CheckListItem),
	string(BlockType_ToggleListItem),
	string(BlockType_Image),
	string(BlockType_Video),
	string(BlockType_Audio),
	string(BlockType_File),
	string(BlockType_Table),
	string(BlockType_CodeBlock),
	string(BlockType_MathBlock),
	string(BlockType_Diagram),
	string(BlockType_Calendar),
}

func (value BlockType) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *BlockType) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = BlockType(string(v))
		return nil
	case string:
		*value = BlockType(v)
		return nil
	}
	return scanError(raw, value)
}

func (value BlockType) Value() (driver.Value, error) {
	return string(value), nil
}

func (value BlockType) String() string {
	return string(value)
}

func (value *BlockType) IsValidEnum() bool {
	return slices.Contains(AllBlockTypes, *value)
}

func ConvertStringToBlockType(enumString string) (*BlockType, error) {
	for _, enumValue := range AllBlockTypes {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid BlockType: %s", enumString)
}
