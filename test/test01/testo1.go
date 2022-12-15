package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"
)

func main() {

	for i := 0; i < 100; i++ {
		fmt.Println()
		i := NewSessionID()
		fmt.Println(i)
		fmt.Println(url.QueryEscape(i))
	}
}

func NewSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		nano := time.Now().UnixNano()
		return strconv.FormatInt(nano, 10)
	}
	return base64.URLEncoding.EncodeToString(b)
}
