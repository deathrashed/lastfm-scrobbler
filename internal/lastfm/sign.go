package lastfm

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"
)

// sign generates a Last.fm API signature for the given parameters.
// The Last.fm signing algorithm is:
//  1. Sort parameters alphabetically by key
//  2. Concatenate each key+value pair (in sorted order)
//  3. Append the apiSecret
//  4. MD5 hash the result
//  5. Return the hex-encoded digest
func sign(params map[string]string, apiSecret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(params[k])
	}
	sb.WriteString(apiSecret)

	h := md5.Sum([]byte(sb.String()))
	return hex.EncodeToString(h[:])
}
