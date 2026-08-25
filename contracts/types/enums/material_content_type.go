package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type MaterialContentType string

const (
	MaterialContentType_None      MaterialContentType = "none"
	MaterialContentType_JSON      MaterialContentType = "application/json"
	MaterialContentType_PDF       MaterialContentType = "application/pdf"
	MaterialContentType_PlainText MaterialContentType = "text/plain"
	MaterialContentType_HTML      MaterialContentType = "text/html"
	MaterialContentType_Markdown  MaterialContentType = "text/markdown"
	MaterialContentType_PNG       MaterialContentType = "image/png"
	MaterialContentType_JPG       MaterialContentType = "image/jpg"
	MaterialContentType_JPEG      MaterialContentType = "image/jpeg"
	MaterialContentType_GIF       MaterialContentType = "image/gif"
	MaterialContentType_SVG       MaterialContentType = "image/svg+xml"
	MaterialContentType_WebP      MaterialContentType = "image/webp"
	MaterialContentType_MP4       MaterialContentType = "video/mp4"
	MaterialContentType_WebM      MaterialContentType = "video/webm"
	MaterialContentType_Mpeg      MaterialContentType = "audio/mpeg"
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

func (value MaterialContentType) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *MaterialContentType) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = MaterialContentType(string(v))
		return nil
	case string:
		*value = MaterialContentType(v)
		return nil
	}
	return scanError(raw, value)
}

func (value MaterialContentType) Value() (driver.Value, error) {
	return string(value), nil
}

func (value MaterialContentType) String() string {
	return string(value)
}

func (value *MaterialContentType) IsValidEnum() bool {
	return slices.Contains(AllMaterialContentTypes, *value)
}

func ConvertStringToMaterialContentType(enumString string) (*MaterialContentType, error) {
	for _, enumValue := range AllMaterialContentTypes {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid MaterialContentType: %s", enumString)
}
