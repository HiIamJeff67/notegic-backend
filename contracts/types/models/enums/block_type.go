package enums

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
