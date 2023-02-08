package gen

import (
	"crypto/rand"
	"errors"
	pv "github.com/wagslane/go-password-validator"
	"image/color"
	"math/big"
	"strings"
	"unicode/utf8"
)

const (
	DefaultLength                uint8 = 16
	DefaultNumberCharSet               = "0123456789"
	DefaultLowercaseCharSet            = "abcdefghijklmnopqrstuvwxyz"
	DefaultIncludeSpecialCharSet       = "~!@#$%^&*-=+"
	DefaultExcludeSpecialCharSet       = "iIl1o0O"
)

var invalidLengthError = errors.New("invalid length error (长度异常)")
var invalidCharsetError = errors.New("invalid charset error (字符集异常)")
var optionsError = errors.New("generation options error (选项异常)")
var buildCharSetError = errors.New("generation build charset error (构建字符集异常)")

func GeneratePassword(conf *PasswdGenConf) (*PasswdGenResult, error) {
	if conf == nil {
		conf = NewDefaultPasswdGenConf()
	}
	if conf.Length <= 0 {
		return nil, invalidLengthError
	}
	if !conf.EnableNumber && !conf.EnableLowercase && !conf.EnableUppercase && len(conf.IncludeSpecialCharSet) == 0 {
		return nil, optionsError
	}
	charSet, err := buildCharSet(conf)
	if err != nil {
		return nil, err
	}
	conf.charSet = charSet
	return internalPasswdGen(conf)
}

type PasswdGenResult struct {
	Password      string
	StrengthInt   float64
	StrengthInfo  string
	StrengthColor color.Color
	CostInfo      string
	CostColor     color.Color
}

type strengthInfo struct {
	strengthInt   float64
	strengthInfo  string
	strengthColor color.Color
	costInfo      string
	costColor     color.Color
}

type PasswdGenConf struct {
	// PUBLIC
	Length                uint8
	EnableNumber          bool
	EnableLowercase       bool
	EnableUppercase       bool
	EnableDuplicate       bool
	IncludeSpecialCharSet string
	ExcludeSpecialCharSet string
	// PRIVATE
	charSet []string
}

func NewDefaultPasswdGenConf() *PasswdGenConf {
	return &PasswdGenConf{
		Length:                DefaultLength,
		EnableNumber:          true,
		EnableLowercase:       true,
		EnableUppercase:       true,
		EnableDuplicate:       true,
		IncludeSpecialCharSet: DefaultIncludeSpecialCharSet,
		ExcludeSpecialCharSet: DefaultExcludeSpecialCharSet,
	}
}

func buildCharSet(conf *PasswdGenConf) ([]string, error) {
	var charSet string
	if conf.EnableNumber {
		charSet += DefaultNumberCharSet
	}
	if conf.EnableLowercase {
		charSet += DefaultLowercaseCharSet
	}
	if conf.EnableUppercase {
		charSet += strings.ToUpper(DefaultLowercaseCharSet)
	}
	lis := len(conf.IncludeSpecialCharSet)
	if lis > 0 {
		if lis != utf8.RuneCountInString(conf.IncludeSpecialCharSet) {
			return nil, buildCharSetError
		}
		charSet += conf.IncludeSpecialCharSet
	}
	les := len(conf.ExcludeSpecialCharSet)
	if les > 0 {
		if les != utf8.RuneCountInString(conf.ExcludeSpecialCharSet) {
			return nil, buildCharSetError
		}
		charSet = removeCharsetFor(charSet, conf.ExcludeSpecialCharSet)
	}
	split := strings.Split(charSet, "")
	// 移除字符集中的重复字符
	split = removeDuplicateChars(split)
	if !conf.EnableDuplicate {
		// 如果不允许重复字符并且要求生成的长度大于现存字符集长度抛出错误
		if int(conf.Length) > len(split) {
			return nil, invalidCharsetError
		}
	}
	return split, nil
}

func removeCharsetFor(input, charset string) string {
	return strings.Map(func(r rune) rune {
		if !strings.ContainsRune(charset, r) {
			return r
		}
		return -1
	}, input)
}

func removeDuplicateChars(sl []string) []string {
	r := make([]string, 0)
	m := make(map[string]struct{})
	for _, val := range sl {
		if _, ok := m[val]; !ok {
			r = append(r, val)
			m[val] = struct{}{}
		}
	}
	return r
}

func internalPasswdGen(conf *PasswdGenConf) (*PasswdGenResult, error) {
	var result string
	charsetToUse := conf.charSet
	max := big.NewInt(int64(len(charsetToUse)))
	if conf.EnableDuplicate {
		for i := uint8(0); i < conf.Length; i++ {
			index, err := rand.Int(rand.Reader, max)
			if err != nil {
				return nil, err
			}
			result += charsetToUse[index.Int64()]
		}
	} else {
		m := make(map[string]struct{}, 0)
		for uint8(len(m)) < conf.Length {
			index, err := rand.Int(rand.Reader, max)
			if err != nil {
				return nil, err
			}
			m[charsetToUse[index.Int64()]] = struct{}{}
		}
		for k := range m {
			result += k
		}
	}
	csi := generateStrengthInfo(result)
	return &PasswdGenResult{
		Password:      result,
		StrengthInt:   csi.strengthInt,
		StrengthInfo:  csi.strengthInfo,
		StrengthColor: csi.strengthColor,
		CostInfo:      csi.costInfo,
		CostColor:     csi.costColor,
	}, nil
}

func generateStrengthInfo(passwd string) *strengthInfo {
	entropy := pv.GetEntropy(passwd)
	// ColorOrange
	cc := color.NRGBA{R: 0xff, G: 0x98, B: 0x00, A: 0xff}
	switch {
	case entropy < 20:
		return &strengthInfo{
			strengthInt:  entropy,
			strengthInfo: "VERY WEAK",
			// ColorRed
			strengthColor: color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0xff},
			costInfo:      "(>>> 0 s)",
			costColor:     cc,
		}
	case entropy < 40:
		return &strengthInfo{
			strengthInt:  entropy,
			strengthInfo: "WEAK",
			// ColorOrange
			strengthColor: color.NRGBA{R: 0xff, G: 0x98, B: 0x00, A: 0xff},
			costInfo:      "(>>> 0 s ~ 3.5 y)",
			costColor:     cc,
		}
	case entropy < 60:
		return &strengthInfo{
			strengthInt:  entropy,
			strengthInfo: "NORMAL",
			// ColorYellow
			strengthColor: color.NRGBA{R: 0xff, G: 0xeb, B: 0x3b, A: 0xff},
			costInfo:      "(>>> 0 s ~ 913 mi)",
			costColor:     cc,
		}
	case entropy < 80:
		return &strengthInfo{
			strengthInt:  entropy,
			strengthInfo: "STRONG",
			// ColorBlue
			strengthColor: color.NRGBA{R: 0x29, G: 0x6f, B: 0xf6, A: 0xff},
			costInfo:      "(3.2 h ~ 958 by)",
			costColor:     cc,
		}
	case entropy >= 80:
		return &strengthInfo{
			strengthInt:  entropy,
			strengthInfo: "VERY STRONG",
			// ColorGreen
			strengthColor: color.NRGBA{R: 0x8b, G: 0xc3, B: 0x4a, A: 0xff},
			costInfo:      "(383 y ~ +Infinity)",
			costColor:     cc,
		}
	default:
		return &strengthInfo{
			strengthInt:   entropy,
			strengthInfo:  "UNKNOWN",
			strengthColor: nil,
			costInfo:      "",
			costColor:     nil,
		}
	}
}
