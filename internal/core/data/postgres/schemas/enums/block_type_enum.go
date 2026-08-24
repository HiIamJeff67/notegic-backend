package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type BlockType enumcontract.BlockType

func (value *BlockType) ToContractable() *enumcontract.BlockType {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.BlockType(*value)
	return &contractValue
}

func BlockTypeToStorable(value *enumcontract.BlockType) *BlockType {
	if value == nil {
		return nil
	}

	storableValue := BlockType(*value)
	return &storableValue
}

const (
	BlockType_Paragraph BlockType = BlockType(enumcontract.BlockType_Paragraph)
	BlockType_Heading   BlockType = BlockType(enumcontract.BlockType_Heading)
	BlockType_Quote     BlockType = BlockType(enumcontract.BlockType_Quote)

	BlockType_BulletListItem   BlockType = BlockType(enumcontract.BlockType_BulletListItem)
	BlockType_NumberedListItem BlockType = BlockType(enumcontract.BlockType_NumberedListItem)
	BlockType_CheckListItem    BlockType = BlockType(enumcontract.BlockType_CheckListItem)
	BlockType_ToggleListItem   BlockType = BlockType(enumcontract.BlockType_ToggleListItem)

	BlockType_Image BlockType = BlockType(enumcontract.BlockType_Image)
	BlockType_Video BlockType = BlockType(enumcontract.BlockType_Video)
	BlockType_Audio BlockType = BlockType(enumcontract.BlockType_Audio)
	BlockType_File  BlockType = BlockType(enumcontract.BlockType_File)

	BlockType_Table     BlockType = BlockType(enumcontract.BlockType_Table)
	BlockType_CodeBlock BlockType = BlockType(enumcontract.BlockType_CodeBlock)
	BlockType_MathBlock BlockType = BlockType(enumcontract.BlockType_MathBlock)
	BlockType_Diagram   BlockType = BlockType(enumcontract.BlockType_Diagram)
	BlockType_Calendar  BlockType = BlockType(enumcontract.BlockType_Calendar)
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
	string(BlockType_Quote),
	string(BlockType_Heading),
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

func (bt BlockType) Name() string {
	return reflect.TypeOf(bt).Name()
}

func (bt *BlockType) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*bt = BlockType(string(v))
		return nil
	case string:
		*bt = BlockType(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into BlockType", value)
}

func (bt BlockType) Value() (driver.Value, error) {
	return string(bt), nil
}

func (bt BlockType) String() string {
	return string(bt)
}

func (bt *BlockType) IsValidEnum() bool {
	return slices.Contains(AllBlockTypes, *bt)
}

func ConvertStringToBlockType(enumString string) (*BlockType, error) {
	for _, blockType := range AllBlockTypes {
		if string(blockType) == enumString {
			return &blockType, nil
		}
	}
	return nil, fmt.Errorf("invalid block type: %s", enumString)
}
