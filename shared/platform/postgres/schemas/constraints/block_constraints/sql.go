package blockconstraints

import (
	_ "embed"
)

var (
	//go:embed block_sibling_pointer_constraints.sql
	BlockSiblingPointerConstraintsSQL string
)
