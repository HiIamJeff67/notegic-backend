package apicontract

import cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"

type InitializeBlockPackYjsDocumentReqDto struct {
	Blocks []cblocknote.ArborizedEditableBlock `json:"blocks"`
}

type InitializeBlockPackYjsDocumentResDto struct {
	Snapshot    []byte `json:"snapshot"`
	StateVector []byte `json:"stateVector"`
}
