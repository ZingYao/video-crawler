package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// Md5String 返回字符串的 md5 值（32位小写十六进制）
func Md5String(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// Md5Bytes 返回字节切片的 md5 值（32位小写十六进制）
func Md5Bytes(b []byte) string {
	hash := md5.Sum(b)
	return hex.EncodeToString(hash[:])
}

// AddSaltToPassword 对密码进行加盐处理，返回加盐后的密码字符串
func AddSaltToPassword(password, salt string) string {
	return password + ":" + salt
}

// SaltedMd5Password 返回加盐后密码的 md5 值
func SaltedMd5Password(password, salt string) string {
	salted := AddSaltToPassword(password, salt)
	return Md5String(salted)
}
