package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

var ErrInvalidInitData = errors.New("invalid telegram init data")

type TelegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
}

// ValidateInitData checks the Telegram Web App initData HMAC and returns the user.
// See https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func ValidateInitData(initData, botToken string) (*TelegramUser, error) {
	vals, err := url.ParseQuery(initData)
	if err != nil {
		return nil, ErrInvalidInitData
	}

	hash := vals.Get("hash")
	if hash == "" {
		return nil, ErrInvalidInitData
	}

	// Build the data-check string: sorted key=value pairs (excluding "hash"), joined by \n.
	var pairs []string
	for k, vs := range vals {
		if k == "hash" || len(vs) == 0 {
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, vs[0]))
	}
	sort.Strings(pairs)
	checkString := strings.Join(pairs, "\n")

	// secret_key = HMAC-SHA256(botToken, key="WebAppData")
	mac := hmac.New(sha256.New, []byte("WebAppData"))
	mac.Write([]byte(botToken))
	secretKey := mac.Sum(nil)

	// expected hash = HMAC-SHA256(checkString, key=secretKey)
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(checkString))
	expected := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(hash), []byte(expected)) {
		return nil, ErrInvalidInitData
	}

	userJSON := vals.Get("user")
	if userJSON == "" {
		return nil, ErrInvalidInitData
	}

	var user TelegramUser
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, ErrInvalidInitData
	}

	return &user, nil
}

// usernameOf returns the display name for a TelegramUser.
func usernameOf(u *TelegramUser) string {
	if u.Username != "" {
		return u.Username
	}
	return u.FirstName
}
