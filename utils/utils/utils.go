package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"unsafe"
)

func MD5Str(raw []byte) string {
	return fmt.Sprintf("%x", md5.Sum(raw))
}

func Sha1(raw []byte) string {
	return fmt.Sprintf("%x", sha1.Sum(raw))
}

/*
 *@func:标准base64编码
 */
func Base64(raw []byte) string {
	return base64.StdEncoding.EncodeToString(raw)
}

/*
 *@func: 标准base64 url编码
 */
func Base64URL(raw []byte) string {
	return base64.URLEncoding.EncodeToString(raw)
}

/*
 *@note: 返回的[]byte不可修改
 */
func Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func ToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
