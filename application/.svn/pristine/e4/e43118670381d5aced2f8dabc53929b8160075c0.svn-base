package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var (
	ErrAuthDateMissing  = errors.New("auth_date is missing")
	ErrSignMissing      = errors.New("sign is missing")
	ErrSignInvalid      = errors.New("sign is invalid")
	ErrUnexpectedFormat = errors.New("init data has unexpected format")
	ErrExpired          = errors.New("init data is expired")
)

// InitData contains init data.
// https://docs.telegram-mini-apps.com/launch-parameters/init-data#parameters-list
type InitData struct {
	// The date the initialization data was created. Is a number representing a
	// Unix timestamp.
	AuthDateRaw int `json:"auth_date"`

	// Optional. The number of seconds after which a message can be sent via
	// the method answerWebAppQuery.
	// https://core.telegram.org/bots/api#answerwebappquery
	CanSendAfterRaw int `json:"can_send_after"`

	// Optional. An object containing information about the chat with the bot in
	// which the Mini Apps was launched. It is returned only for Mini Apps
	// opened through the attachment menu.
	Chat Chat `json:"chat"`

	// Optional. The type of chat from which the Mini Apps was opened.
	// Returned only for applications opened by direct link.
	ChatType ChatType `json:"chat_type"`

	// Optional. A global identifier indicating the chat from which the Mini
	// Apps was opened. Returned only for applications opened by direct link.
	ChatInstance int64 `json:"chat_instance"`

	// Initialization data signature.
	// https://core.telegram.org/bots/webapps#validating-data-received-via-the-web-app
	Hash string `json:"hash"`

	// Optional. The unique session ID of the Mini App. Used in the process of
	// sending a message via the method answerWebAppQuery.
	// https://core.telegram.org/bots/api#answerwebappquery
	QueryID string `json:"query_id"`

	// Optional. An object containing data about the chat partner of the current
	// user in the chat where the bot was launched via the attachment menu.
	// Returned only for private chats and only for Mini Apps launched via the
	// attachment menu.
	Receiver User `json:"receiver"`

	// Optional. The value of the startattach or startapp query parameter
	// specified in the link. It is returned only for Mini Apps opened through
	// the attachment menu.
	StartParam string `json:"start_param"`

	// Optional. An object containing information about the current user.
	User User `json:"user"`
}

// AuthDate returns AuthDateRaw as time.Time.
func (d *InitData) AuthDate() time.Time {
	return time.Unix(int64(d.AuthDateRaw), 0)
}

// CanSendAfter returns computed time which depends on CanSendAfterRaw and
// AuthDate. Originally, CanSendAfterRaw means time in seconds, after which
// `answerWebAppQuery` method can be called and that's why this value could
// be computed as time.
func (d *InitData) CanSendAfter() time.Time {
	return d.AuthDate().Add(time.Duration(d.CanSendAfterRaw) * time.Second)
}

// User describes user information:
// https://docs.telegram-mini-apps.com/launch-parameters/init-data#user
type User struct {
	// Optional. True, if this user added the bot to the attachment menu.
	AddedToAttachmentMenu bool `json:"added_to_attachment_menu"`

	// Optional. True, if this user allowed the bot to message them.
	AllowsWriteToPm bool `json:"allows_write_to_pm"`

	// First name of the user or bot.
	FirstName string `json:"first_name"`

	// A unique identifier for the user or bot.
	ID int64 `json:"id"`

	// Optional. True, if this user is a bot. Returned in the `receiver` field
	// only.
	IsBot bool `json:"is_bot"`

	// Optional. True, if this user is a Telegram Premium user.
	IsPremium bool `json:"is_premium"`

	// Optional. Last name of the user or bot.
	LastName string `json:"last_name"`

	// Optional. Username of the user or bot.
	Username string `json:"username"`

	// Optional. IETF language tag of the user's language. Returns in user
	// field only.
	// https://en.wikipedia.org/wiki/IETF_language_tag
	LanguageCode string `json:"language_code"`

	// Optional. URL of the user’s profile photo. The photo can be in .jpeg or
	// .svg formats. Only returned for Web Apps launched from the
	// attachment menu.
	PhotoURL string `json:"photo_url"`
}

const (
	ChatTypeSender     ChatType = "sender"
	ChatTypePrivate    ChatType = "private"
	ChatTypeGroup      ChatType = "group"
	ChatTypeSupergroup ChatType = "supergroup"
	ChatTypeChannel    ChatType = "channel"
)

// ChatType describes type of chat.
type ChatType string

// Known returns true if current chat type is known.
func (c ChatType) Known() bool {
	switch c {
	case ChatTypeSender,
		ChatTypePrivate,
		ChatTypeGroup,
		ChatTypeSupergroup,
		ChatTypeChannel:
		return true
	default:
		return false
	}
}

// Chat describes chat information:
// https://docs.telegram-mini-apps.com/launch-parameters/init-data#chat
type Chat struct {
	// Unique identifier for this chat.
	ID int64 `json:"id"`

	// Type of chat.
	Type ChatType `json:"type"`

	// Title of the chat.
	Title string `json:"title"`

	// Optional. URL of the chat’s photo. The photo can be in .jpeg or .svg
	// formats. Only returned for Web Apps launched from the attachment menu.
	PhotoURL string `json:"photo_url"`

	// Optional. Username of the chat.
	Username string `json:"username"`
}

var (
	// List of properties which should always be interpreted as strings.
	_stringProps = map[string]bool{
		"start_param": true,
	}
)

// Parse converts passed init data presented as query string to InitData
// object.
func Parse(initData string) (InitData, error) {
	decoded, err := url.QueryUnescape(initData)
	if err != nil {
		fmt.Println("Error decoding query:", err)
		return InitData{}, ErrUnexpectedFormat
	}

	// Parse passed init data as query string.
	q, err := url.ParseQuery(decoded)
	if err != nil {
		return InitData{}, ErrUnexpectedFormat
	}

	// According to documentation, we could only meet such types as int64,
	// string, or another object. So, we create
	pairs := make([]string, 0, len(q))
	for k, v := range q {
		// Derive real value. We know that there can not be any arrays and value
		// can be the only one.
		val := v[0]
		valFormat := "%q:%q"

		// If passed value is valid in the context of JSON, it means, we could
		// insert this value without formatting.
		if isString := _stringProps[k]; !isString && json.Valid([]byte(val)) {
			valFormat = "%q:%s"
		}

		pairs = append(pairs, fmt.Sprintf(valFormat, k, val))
	}

	// Unmarshal JSON to our custom structure.
	var d InitData
	jStr := fmt.Sprintf("{%s}", strings.Join(pairs, ","))
	if err := json.Unmarshal([]byte(jStr), &d); err != nil {
		return InitData{}, ErrUnexpectedFormat
	}
	return d, nil
}
