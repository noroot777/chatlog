package util

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func IsNormalString(b []byte) bool {
	str := string(b)

	// 检查是否为有效的 UTF-8
	if !utf8.ValidString(str) {
		return false
	}

	// 检查是否全部为可打印字符
	for _, r := range str {
		if !unicode.IsPrint(r) {
			return false
		}
	}

	return true
}

func MustAnyToInt(v interface{}) int {
	str := fmt.Sprintf("%v", v)
	if i, err := strconv.Atoi(str); err == nil {
		return i
	}
	return 0
}

func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

func SplitInt64ToTwoInt32(input int64) (int64, int64) {
	return input & 0xFFFFFFFF, input >> 32
}

func Str2List(str string, sep string) []string {
	list := make([]string, 0)

	if str == "" {
		return list
	}

	listMap := make(map[string]bool)
	for _, elem := range strings.Split(str, sep) {
		elem = strings.TrimSpace(elem)
		if len(elem) == 0 {
			continue
		}
		if _, ok := listMap[elem]; ok {
			continue
		}
		listMap[elem] = true
		list = append(list, elem)
	}

	return list
}

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func IntToBase62(n int64) string {
	if n == 0 {
		return "0"
	}
	result := make([]byte, 0)
	negative := n < 0
	if negative {
		n = -n
	}
	for n > 0 {
		remainder := n % 62
		result = append([]byte{base62Chars[remainder]}, result...)
		n = n / 62
	}
	if negative {
		result = append([]byte{'-'}, result...)
	}
	return string(result)
}
