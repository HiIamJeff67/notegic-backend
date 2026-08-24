package enums

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type SupportedIcon enumcontract.SupportedIcon

func (value *SupportedIcon) ToContractable() *enumcontract.SupportedIcon {
	if value == nil {
		return nil
	}

	contractValue := enumcontract.SupportedIcon(*value)
	return &contractValue
}

func SupportedIconToStorable(value *enumcontract.SupportedIcon) *SupportedIcon {
	if value == nil {
		return nil
	}

	storableValue := SupportedIcon(*value)
	return &storableValue
}

const (
	SupportedIcon_GrinningFace               SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_GrinningFace)
	SupportedIcon_SmilingFaceWithSmilingEyes SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_SmilingFaceWithSmilingEyes)
	SupportedIcon_RedHeart                   SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_RedHeart)
	SupportedIcon_Fire                       SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_Fire)
	SupportedIcon_Star                       SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_Star)
	SupportedIcon_Books                      SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_Books)
	SupportedIcon_Notebook                   SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_Notebook)
	SupportedIcon_PencilPaper                SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_PencilPaper)
	SupportedIcon_Lightbulb                  SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_Lightbulb)
	SupportedIcon_Rocket                     SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_Rocket)
	SupportedIcon_CheckMark                  SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_CheckMark)
	SupportedIcon_Pin                        SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_Pin)
	SupportedIcon_FolderOpen                 SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_FolderOpen)
	SupportedIcon_Calendar                   SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_Calendar)
	SupportedIcon_Clock                      SupportedIcon = SupportedIcon(enumcontract.SupportedIcon_Clock)
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

func (si SupportedIcon) Name() string {
	return reflect.TypeOf(si).Name()
}

func (si *SupportedIcon) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*si = SupportedIcon(string(v))
		return nil
	case string:
		*si = SupportedIcon(v)
		return nil
	}
	return scanError(value, si)
}

func (si SupportedIcon) Value() (driver.Value, error) {
	return string(si), nil
}

func (si SupportedIcon) String() string {
	return string(si)
}

func (si *SupportedIcon) IsValidEnum() bool {
	return slices.Contains(AllSupportedIcons, *si)
}

func ConvertStringToSupportedIcon(enumString string) (*SupportedIcon, error) {
	for _, supportedIcon := range AllSupportedIcons {
		if string(supportedIcon) == enumString {
			return &supportedIcon, nil
		}
	}
	return nil, fmt.Errorf("invalid supported icon: %s", enumString)
}
