package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

// MaterialContentType indicates the MIME content type of material files.
type MaterialContentType enumcontract.MaterialContentType

func (value *MaterialContentType) ToContractable() *enumcontract.MaterialContentType {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.MaterialContentType(*value)
	return &contractValue
}

func MaterialContentTypeToStorable(value *enumcontract.MaterialContentType) *MaterialContentType {
	if value == nil {
		return nil
	}

	storableValue := MaterialContentType(*value)
	return &storableValue
}

const (
	MaterialContentType_None      MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_None)
	MaterialContentType_JSON      MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_JSON)
	MaterialContentType_PDF       MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_PDF)
	MaterialContentType_PlainText MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_PlainText)
	MaterialContentType_HTML      MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_HTML)
	MaterialContentType_Markdown  MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_Markdown)
	MaterialContentType_PNG       MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_PNG)
	MaterialContentType_JPG       MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_JPG)
	MaterialContentType_JPEG      MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_JPEG)
	MaterialContentType_GIF       MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_GIF)
	MaterialContentType_SVG       MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_SVG)
	MaterialContentType_WebP      MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_WebP)
	MaterialContentType_MP4       MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_MP4)
	MaterialContentType_WebM      MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_WebM)
	MaterialContentType_Mpeg      MaterialContentType = MaterialContentType(enumcontract.MaterialContentType_Mpeg)
)

var AllMaterialContentTypes = []MaterialContentType{
	MaterialContentType_None,
	MaterialContentType_JSON,
	MaterialContentType_PDF,
	MaterialContentType_PlainText,
	MaterialContentType_HTML,
	MaterialContentType_Markdown,
	MaterialContentType_PNG,
	MaterialContentType_JPG,
	MaterialContentType_JPEG,
	MaterialContentType_GIF,
	MaterialContentType_SVG,
	MaterialContentType_WebP,
	MaterialContentType_MP4,
	MaterialContentType_WebM,
	MaterialContentType_Mpeg,
}

var AllMaterialContentTypeStrings = []string{
	string(MaterialContentType_None),
	string(MaterialContentType_JSON),
	string(MaterialContentType_PDF),
	string(MaterialContentType_PlainText),
	string(MaterialContentType_HTML),
	string(MaterialContentType_Markdown),
	string(MaterialContentType_PNG),
	string(MaterialContentType_JPG),
	string(MaterialContentType_JPEG),
	string(MaterialContentType_GIF),
	string(MaterialContentType_SVG),
	string(MaterialContentType_WebP),
	string(MaterialContentType_MP4),
	string(MaterialContentType_WebM),
	string(MaterialContentType_Mpeg),
}

func (mct MaterialContentType) Name() string {
	return reflect.TypeOf(mct).Name()
}

func (mct *MaterialContentType) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*mct = MaterialContentType(string(v))
		return nil
	case string:
		*mct = MaterialContentType(v)
		return nil
	}
	return scanError(value, mct)
}

func (mct MaterialContentType) Value() (driver.Value, error) {
	return string(mct), nil
}

func (mct MaterialContentType) String() string {
	return string(mct)
}

func (mct *MaterialContentType) IsValidEnum() bool {
	return slices.Contains(AllMaterialContentTypes, *mct)
}

func ConvertStringToMaterialContentType(enumString string) (*MaterialContentType, error) {
	for _, materialContentType := range AllMaterialContentTypes {
		if string(materialContentType) == enumString {
			return &materialContentType, nil
		}
	}
	return nil, fmt.Errorf("invalid material content type: %s", enumString)
}
