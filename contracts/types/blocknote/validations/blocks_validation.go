package validations

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10" // make sure we use the version 10

	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

func RegisterShelfBlockValidation(validate *validator.Validate) {
	validate.RegisterValidation("isshelfname", func(fl validator.FieldLevel) bool {
		shelfNameStr := fl.Field().String()
		if len(shelfNameStr) > cblocknote.MaxShelfNameLength {
			return false
		}
		return !regexp.MustCompile(`[\/\\:\*\?"<>\|]`).MatchString(shelfNameStr)
	})
	validate.RegisterValidation("isitemname", func(fl validator.FieldLevel) bool {
		itemNameStr := fl.Field().String()
		if len(itemNameStr) > cblocknote.MaxItemNameLength {
			return false
		}
		return !regexp.MustCompile(`[\/\\:\*\?"<>\|]`).MatchString(itemNameStr)
	})
	validate.RegisterValidation("isfileblockname", func(fl validator.FieldLevel) bool {
		fileBlockNameStr := fl.Field().String()
		return len(fileBlockNameStr) <= cblocknote.MaxFileBlockNameLength
	})
	validate.RegisterValidation("isfileblockcaption", func(fl validator.FieldLevel) bool {
		fileBlockCaptionStr := fl.Field().String()
		return len(fileBlockCaptionStr) <= cblocknote.MaxFileBlockCaptionLength
	})
	validate.RegisterValidation("istextalignment", func(fl validator.FieldLevel) bool {
		textAlignmentStr := fl.Field().String()
		return textAlignmentStr == "left" || textAlignmentStr == "right" || textAlignmentStr == "center" || textAlignmentStr == "justify"
	})
	validate.RegisterValidation("isheadinglevel", func(fl validator.FieldLevel) bool {
		headingLevel := fl.Field().Int()
		return headingLevel > 0 && headingLevel <= int64(cblocknote.MaxHeadingLevel)
	})
	validate.RegisterValidation("isprogramminglanguage", func(fl validator.FieldLevel) bool {
		programmingLanguageStr := fl.Field().String()
		if len(programmingLanguageStr) > cblocknote.MaxProgrammingLanguageLength {
			return false
		}
		for _, l := range cenums.AllSupportedProgrammingLanguageStrings {
			if programmingLanguageStr == l {
				return true
			}
		}
		return false
	})
	validate.RegisterValidation("iscolororhexcode", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()

		if val == "default" {
			return true
		}
		validColors := []string{
			"gray", "brown", "orange", "yellow", "green", "blue", "purple", "pink", "red",
		}
		for _, color := range validColors {
			if val == color {
				return true
			}
		}

		if !strings.HasPrefix(val, "#") {
			return false
		}
		cleanHexCode := val[1:]
		length := len(cleanHexCode)
		if length != 3 && length != 4 && length != 6 && length != 8 {
			return false
		}

		return regexp.MustCompile(`^[0-9a-fA-F]+$`).MatchString(cleanHexCode)
	})
}
