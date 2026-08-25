package validations

import (
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10" // make sure we use the version 10

	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	stringutil "github.com/HiIamJeff67/notegic-backend/shared/lib/strings"
)

var urlWhiteList = []string{"http", "https", "mailto", "tel", "ws"}

var urlBlackList = []string{"javascript", "vbscript", "file", "data"}

func RegisterStringsValidation(validate *validator.Validate) {
	validate.RegisterValidation("isaccount", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()

		// try email validation
		if stringutil.IsEmailString(val) {
			return true
		}

		// try alphaandnum validation
		hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(val)
		hasDigit := regexp.MustCompile(`\d`).MatchString(val)

		return hasLetter && hasDigit
	})
	validate.RegisterValidation("alphaandnum", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return stringutil.IsAlphaAndNumberString(val)
	})
	validate.RegisterValidation("isstrongpassword", func(fl validator.FieldLevel) bool {
		password := strings.TrimSpace(fl.Field().String())
		if len(password) < constants.MinPasswordLength || len(password) > constants.MaxPasswordLength {
			return false
		}
		hasUpperCaseLetter := regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLowerCaseLetter := regexp.MustCompile(`[a-z]`).MatchString(password)
		hasDigit := regexp.MustCompile(`\d`).MatchString(password)
		hasSpecialCharacter := regexp.MustCompile(`[^\w\s]`).MatchString(password)
		return hasUpperCaseLetter && hasLowerCaseLetter && hasDigit && hasSpecialCharacter
	})
	validate.RegisterValidation("isuseragent", func(fl validator.FieldLevel) bool {
		userAgentStr := strings.TrimSpace(fl.Field().String())
		if len(userAgentStr) < 3 || len(userAgentStr) > constants.MaxUserAgentLength {
			return false
		}

		// check if the userAgent contain some malicious content
		if strings.Contains(userAgentStr, "<script>") ||
			strings.Contains(userAgentStr, "javascript:") ||
			strings.Contains(userAgentStr, "data:") {
			return false
		}

		return true
	})
	validate.RegisterValidation("isnumberstring", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return stringutil.IsNumberString(val)
	})
	validate.RegisterValidation("isgoogleauthenticationcode", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return value != "" &&
			len(value) <= 4096 &&
			value == strings.TrimSpace(value) &&
			!strings.ContainsFunc(value, unicode.IsSpace) &&
			!strings.ContainsFunc(value, unicode.IsControl)
	})
	validate.RegisterValidation("isurl", func(fl validator.FieldLevel) bool {
		urlStr := strings.TrimSpace(fl.Field().String())
		if len(urlStr) > constants.MaxURLLength {
			return false
		}
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return false
		}

		scheme := strings.ToLower(parsedURL.Scheme)
		if slices.Contains(urlWhiteList, scheme) {
			return true
		}
		if slices.Contains(urlBlackList, scheme) {
			return false
		}

		return parsedURL.Scheme != "" && parsedURL.Host != ""
	})
	validate.RegisterValidation("isimageurl", func(fl validator.FieldLevel) bool {
		urlStr := strings.TrimSpace(fl.Field().String())
		if len(urlStr) > constants.MaxURLLength {
			return false
		}
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return false
		}
		scheme := strings.ToLower(parsedURL.Scheme)
		if scheme == "http" || scheme == "https" {
			return true
		}

		return scheme != "" && parsedURL.Host != ""
	})
	validate.RegisterValidation("ishexcodecolor", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^#[0-9a-fA-F]{6}$`).MatchString(fl.Field().String())
	})
	validate.RegisterValidation("istimezone", func(fl validator.FieldLevel) bool {
		tzStr := fl.Field().String()
		_, err := time.LoadLocation(tzStr)
		return err == nil
	})
}
