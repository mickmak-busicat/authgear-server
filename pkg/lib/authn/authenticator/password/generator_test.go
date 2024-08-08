package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func newInt(v int) *int { return &v }

type TestRandSource struct {
}

// Pseudo-random pick
func (s *TestRandSource) RandomBytes(n int) ([]byte, error) {
	return []byte("1234567890"), nil
}

func (s *TestRandSource) Shuffle(list string) (string, error) {
	return list, nil
}

func TestBasicPasswordGeneration(t *testing.T) {
	Convey("Given a password generator with default settings", t, func() {
		generator := &Generator{
			Checker:    &Checker{},
			RandSource: &TestRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength: newInt(8),
			},
		}

		password, err := generator.Generate()

		Convey("should be at least 8 characters long", func() {
			So(len(password), ShouldBeGreaterThanOrEqualTo, 8)
			So(err, ShouldBeNil)
		})
	})
}

func TestUppercaseRequirement(t *testing.T) {
	Convey("Given a password generator requiring at least one uppercase letter", t, func() {
		generator := &Generator{
			Checker:    &Checker{},
			RandSource: &TestRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength:         newInt(8),
				UppercaseRequired: true,
			},
		}

		password, err := generator.Generate()

		Convey("should include at least one uppercase letter", func() {
			So(checkPasswordUppercase(password), ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}

func TestLowercaseRequirement(t *testing.T) {
	Convey("Given a password generator requiring at least one lowercase letter", t, func() {
		generator := &Generator{
			Checker:    &Checker{},
			RandSource: &TestRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength:         newInt(8),
				LowercaseRequired: true,
			},
		}

		password, err := generator.Generate()

		Convey("should include at least one lowercase letter", func() {
			So(checkPasswordLowercase(password), ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}

func TestCombinedRequirements(t *testing.T) {
	Convey("Given a password generator with multiple requirements", t, func() {
		generator := &Generator{
			Checker:    &Checker{},
			RandSource: &TestRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength:         newInt(12),
				UppercaseRequired: true,
				LowercaseRequired: true,
				DigitRequired:     true,
				SymbolRequired:    true,
			},
		}

		password, err := generator.Generate()

		Convey("should meet all requirements", func() {
			So(checkPasswordUppercase(password), ShouldBeTrue)
			So(checkPasswordLowercase(password), ShouldBeTrue)
			So(checkPasswordDigit(password), ShouldBeTrue)
			So(checkPasswordSymbol(password), ShouldBeTrue)
			So(len(password), ShouldEqual, 12)
			So(err, ShouldBeNil)
		})
	})
}

func TestMinLengthRequirement(t *testing.T) {
	Convey("Given a password generator with a minimum length requirement", t, func() {
		generator := &Generator{
			Checker:    &Checker{},
			RandSource: &TestRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength: newInt(40),
			},
		}

		password, err := generator.Generate()

		Convey("should meet the minimum length requirement", func() {
			So(len(password), ShouldBeGreaterThanOrEqualTo, 40)
			So(err, ShouldBeNil)
		})
	})
}

func TestMinGuessableLevelRequirement(t *testing.T) {
	Convey("Given a password generator with a minimum guessable level requirement", t, func() {
		generator := &Generator{
			Checker:    &Checker{},
			RandSource: &CryptoRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength:             newInt(8),
				MinimumGuessableLevel: 4,
			},
		}

		password, err := generator.Generate()

		Convey("should meet the minimum guessable level requirement", func() {
			level, _ := checkPasswordGuessableLevel(password, 4)
			So(len(password), ShouldBeGreaterThanOrEqualTo, 32)
			So(level, ShouldBeGreaterThanOrEqualTo, 4)
			So(err, ShouldBeNil)
		})
	})
}

func TestExcludedKeywordsRequirement(t *testing.T) {
	Convey("Given a password generator with excluded keywords", t, func() {
		excluded := []string{"1", "2", "3"}
		generator := &Generator{
			Checker: &Checker{
				PwExcludedKeywords: excluded,
			},
			RandSource: &CryptoRandSource{},
			Policy: &config.PasswordPolicyConfig{
				MinLength:        newInt(8),
				DigitRequired:    true,
				ExcludedKeywords: excluded,
			},
		}

		password, err := generator.Generate()

		Convey("should not contain any excluded keywords", func() {
			So(checkPasswordExcludedKeywords(password, excluded), ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})
}

func TestPrepareCharacterSet(t *testing.T) {
	Convey("Given a password policy config", t, func() {
		Convey("When no specific requirements are set", func() {
			policy := &config.PasswordPolicyConfig{}
			result, _ := prepareCharacterSet(policy)
			So(result, ShouldContainSubstring, CharListLowercase)
			So(result, ShouldContainSubstring, CharListDigit)
		})

		Convey("When lowercase is required", func() {
			policy := &config.PasswordPolicyConfig{LowercaseRequired: true}
			result, _ := prepareCharacterSet(policy)
			So(result, ShouldContainSubstring, CharListLowercase)
		})

		Convey("When uppercase is required", func() {
			policy := &config.PasswordPolicyConfig{UppercaseRequired: true}
			result, _ := prepareCharacterSet(policy)
			So(result, ShouldContainSubstring, CharListUppercase)
		})

		Convey("When alphabet is required but not lowercase or uppercase", func() {
			policy := &config.PasswordPolicyConfig{AlphabetRequired: true}
			result, _ := prepareCharacterSet(policy)
			So(result, ShouldContainSubstring, CharListAlphabet)
		})

		Convey("When digits are required", func() {
			policy := &config.PasswordPolicyConfig{DigitRequired: true}
			result, _ := prepareCharacterSet(policy)
			So(result, ShouldContainSubstring, CharListDigit)
		})

		Convey("When symbols are required", func() {
			policy := &config.PasswordPolicyConfig{SymbolRequired: true}
			result, _ := prepareCharacterSet(policy)
			So(result, ShouldContainSubstring, CharListSymbol)
		})

		Convey("When all character sets are required", func() {
			policy := &config.PasswordPolicyConfig{
				LowercaseRequired: true,
				UppercaseRequired: true,
				AlphabetRequired:  true,
				DigitRequired:     true,
				SymbolRequired:    true,
			}
			result, _ := prepareCharacterSet(policy)
			So(result, ShouldContainSubstring, CharListLowercase)
			So(result, ShouldContainSubstring, CharListUppercase)
			So(result, ShouldContainSubstring, CharListDigit)
			So(result, ShouldContainSubstring, CharListSymbol)
			So(result, shouldNotContainDuplicates)
		})
	})
}

func shouldNotContainDuplicates(actual interface{}, expected ...interface{}) string {
	set := map[characterSet]struct{}{}
	for _, c := range actual.(string) {
		set[characterSet(c)] = struct{}{}
	}
	if len(set) != len(actual.(string)) {
		return "contains duplicates"
	}
	return ""
}

func TestGetMinLength(t *testing.T) {
	Convey("Test getMinLength function", t, func() {
		Convey("When MinLength is greater than DefaultMinLength and GuessableEnabledMinLength", func() {
			minLength := 15
			policy := &config.PasswordPolicyConfig{
				MinLength:             &minLength,
				MinimumGuessableLevel: 0,
			}
			So(getMinLength(policy), ShouldEqual, minLength)
		})

		Convey("When MinLength is less than DefaultMinLength", func() {
			minLength := 5
			policy := &config.PasswordPolicyConfig{
				MinLength:             &minLength,
				MinimumGuessableLevel: 0,
			}
			So(getMinLength(policy), ShouldEqual, DefaultMinLength)
		})

		Convey("When MinLength is less than GuessableEnabledMinLength and MinimumGuessableLevel is greater than 0", func() {
			minLength := 10
			policy := &config.PasswordPolicyConfig{
				MinLength:             &minLength,
				MinimumGuessableLevel: 1,
			}
			So(getMinLength(policy), ShouldEqual, GuessableEnabledMinLength)
		})
	})
}
