package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
)

type SupportedIcon string

const (
	SupportedIcon_GrinningFace               SupportedIcon = "😀"
	SupportedIcon_SmilingFaceWithSmilingEyes SupportedIcon = "😊"
	SupportedIcon_RedHeart                   SupportedIcon = "❤️"
	SupportedIcon_Fire                       SupportedIcon = "🔥"
	SupportedIcon_Star                       SupportedIcon = "⭐"
	SupportedIcon_Books                      SupportedIcon = "📚"
	SupportedIcon_Notebook                   SupportedIcon = "📓"
	SupportedIcon_PencilPaper                SupportedIcon = "📝"
	SupportedIcon_Lightbulb                  SupportedIcon = "💡"
	SupportedIcon_Rocket                     SupportedIcon = "🚀"
	SupportedIcon_CheckMark                  SupportedIcon = "✅"
	SupportedIcon_Pin                        SupportedIcon = "📌"
	SupportedIcon_FolderOpen                 SupportedIcon = "📂"
	SupportedIcon_Calendar                   SupportedIcon = "📅"
	SupportedIcon_Clock                      SupportedIcon = "⏰"
)

var AllSupportedIcons = []SupportedIcon{
	SupportedIcon_GrinningFace,
	SupportedIcon_SmilingFaceWithSmilingEyes,
	SupportedIcon_RedHeart,
	SupportedIcon_Fire,
	SupportedIcon_Star,
	SupportedIcon_Books,
	SupportedIcon_Notebook,
	SupportedIcon_PencilPaper,
	SupportedIcon_Lightbulb,
	SupportedIcon_Rocket,
	SupportedIcon_CheckMark,
	SupportedIcon_Pin,
	SupportedIcon_FolderOpen,
	SupportedIcon_Calendar,
	SupportedIcon_Clock,
}

var AllSupportedIconStrings = []string{
	string(SupportedIcon_GrinningFace),
	string(SupportedIcon_SmilingFaceWithSmilingEyes),
	string(SupportedIcon_RedHeart),
	string(SupportedIcon_Fire),
	string(SupportedIcon_Star),
	string(SupportedIcon_Books),
	string(SupportedIcon_Notebook),
	string(SupportedIcon_PencilPaper),
	string(SupportedIcon_Lightbulb),
	string(SupportedIcon_Rocket),
	string(SupportedIcon_CheckMark),
	string(SupportedIcon_Pin),
	string(SupportedIcon_FolderOpen),
	string(SupportedIcon_Calendar),
	string(SupportedIcon_Clock),
}

func (value SupportedIcon) Name() string {
	return reflect.TypeOf(value).Name()
}

func (value *SupportedIcon) Scan(raw any) error {
	switch v := raw.(type) {
	case []byte:
		*value = SupportedIcon(string(v))
		return nil
	case string:
		*value = SupportedIcon(v)
		return nil
	}
	return scanError(raw, value)
}

func (value SupportedIcon) Value() (driver.Value, error) {
	return string(value), nil
}

func (value SupportedIcon) String() string {
	return string(value)
}

func (value *SupportedIcon) IsValidEnum() bool {
	return slices.Contains(AllSupportedIcons, *value)
}

func ConvertStringToSupportedIcon(enumString string) (*SupportedIcon, error) {
	for _, enumValue := range AllSupportedIcons {
		if string(enumValue) == enumString {
			return &enumValue, nil
		}
	}
	return nil, fmt.Errorf("invalid SupportedIcon: %s", enumString)
}
