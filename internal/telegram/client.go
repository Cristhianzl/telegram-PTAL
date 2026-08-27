// Package telegram sends the messages and discovers the destination chat.
//
// All communication is outbound or long polling, so the daemon works behind
// any home NAT: there is no webhook, no open port, and no need for a public
// IP address.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MaxMessageLen is Telegram's hard per-message limit.
const MaxMessageLen = 4096

// Client talks to the Bot API.
type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

// New creates the client. The timeout is generous because getUpdates uses
// long polling.
func New(token string) *Client {
	return &Client{
		token:   token,
		baseURL: "https://api.telegram.org",
		http:    &http.Client{Timeout: 70 * time.Second},
	}
}

// SetBaseURL points the client at another server, used by tests.
func (c *Client) SetBaseURL(url string) { c.baseURL = strings.TrimSuffix(url, "/") }

// Bot describes the bot's identity.
type Bot struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"first_name"`
}

// Chat is the destination conversation.
type Chat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Username string `json:"username"`
	Title    string `json:"title"`
	First    string `json:"first_name"`
}

// Label returns the most readable identification of the chat.
func (c Chat) Label() string {
	switch {
	case c.Username != "":
		return "@" + c.Username
	case c.Title != "":
		return c.Title
	case c.First != "":
		return c.First
	}
	return fmt.Sprintf("chat %d", c.ID)
}

// Update is one update received through getUpdates.
type Update struct {
	UpdateID int64 `json:"update_id"`
	Message  *struct {
		MessageID int    `json:"message_id"`
		Text      string `json:"text"`
		Chat      Chat   `json:"chat"`
		From      struct {
			Username string `json:"username"`
		} `json:"from"`
	} `json:"message"`
}

// Message is the message returned by sendMessage.
type Message struct {
	MessageID int  `json:"message_id"`
	Chat      Chat `json:"chat"`
}

// APIError carries the code and retry_after returned by Telegram.
type APIError struct {
	Code        int
	Description string
	RetryAfter  time.Duration
}

func (e *APIError) Error() string {
	return fmt.Sprintf("telegram %d: %s", e.Code, e.Description)
}

// IsAuth reports whether the error is an invalid token, in which case
// retrying is pointless.
func (e *APIError) IsAuth() bool { return e.Code == 401 || e.Code == 404 }

// GetMe validates the token and returns the bot's identity.
func (c *Client) GetMe(ctx context.Context) (*Bot, error) {
	var bot Bot
	if err := c.call(ctx, "getMe", nil, &bot); err != nil {
		return nil, err
	}
	return &bot, nil
}

// GetUpdates fetches incoming messages. With timeout > 0 Telegram holds the
// connection until something arrives, which makes long polling cheap.
func (c *Client) GetUpdates(ctx context.Context, offset int64, timeoutSec int) ([]Update, error) {
	payload := map[string]any{
		"timeout":          timeoutSec,
		"allowed_updates":  []string{"message"},
	}
	if offset != 0 {
		payload["offset"] = offset
	}
	var updates []Update
	if err := c.call(ctx, "getUpdates", payload, &updates); err != nil {
		return nil, err
	}
	return updates, nil
}

// SendOptions tunes how the message arrives.
type SendOptions struct {
	// Silent delivers without a sound, used during quiet hours.
	Silent bool
	// Buttons are links shown below the message.
	Buttons [][]Button
}

// Button is an inline button that opens a URL.
type Button struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// Send posts an HTML message with the link preview card disabled, which would
// otherwise turn every alert into a giant GitHub block.
func (c *Client) Send(ctx context.Context, chatID, text string, opts SendOptions) (*Message, error) {
	payload := map[string]any{
		"chat_id":              chatID,
		"text":                 Truncate(text),
		"parse_mode":           "HTML",
		"link_preview_options": map[string]any{"is_disabled": true},
	}
	if opts.Silent {
		payload["disable_notification"] = true
	}
	if len(opts.Buttons) > 0 {
		payload["reply_markup"] = map[string]any{"inline_keyboard": opts.Buttons}
	}

	var msg Message
	if err := c.call(ctx, "sendMessage", payload, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// Edit updates an already-sent message, used by the pinned panel.
func (c *Client) Edit(ctx context.Context, chatID string, messageID int, text string) error {
	payload := map[string]any{
		"chat_id":              chatID,
		"message_id":           messageID,
		"text":                 Truncate(text),
		"parse_mode":           "HTML",
		"link_preview_options": map[string]any{"is_disabled": true},
	}
	err := c.call(ctx, "editMessageText", payload, nil)
	// Editing to identical text returns an error; that is not a problem.
	var apiErr *APIError
	if err != nil && asAPIError(err, &apiErr) &&
		strings.Contains(apiErr.Description, "message is not modified") {
		return nil
	}
	return err
}

// Pin pins the message to the top of the conversation.
func (c *Client) Pin(ctx context.Context, chatID string, messageID int) error {
	return c.call(ctx, "pinChatMessage", map[string]any{
		"chat_id":              chatID,
		"message_id":           messageID,
		"disable_notification": true,
	}, nil)
}

// Truncate keeps the message inside Telegram's limit, cutting at the last
// line break so an HTML tag is never split in half.
func Truncate(text string) string {
	if len(text) <= MaxMessageLen {
		return text
	}
	const suffix = "\n\n<i>… message truncated</i>"
	limit := MaxMessageLen - len(suffix)
	cut := text[:limit]
	if idx := strings.LastIndex(cut, "\n"); idx > limit/2 {
		cut = cut[:idx]
	}
	return cut + suffix
}

// call invokes a Bot API method, retrying in a way that honors retry_after
// exactly when Telegram applies rate limiting.
func (c *Client) call(ctx context.Context, method string, payload any, out any) error {
	const maxAttempts = 3
	var lastErr error

	for attempt := range maxAttempts {
		if attempt > 0 {
			wait := time.Duration(1<<attempt) * time.Second
			var apiErr *APIError
			if asAPIError(lastErr, &apiErr) && apiErr.RetryAfter > 0 {
				wait = apiErr.RetryAfter
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		err := c.callOnce(ctx, method, payload, out)
		if err == nil {
			return nil
		}
		var apiErr *APIError
		if asAPIError(err, &apiErr) && apiErr.IsAuth() {
			return err
		}
		lastErr = err
	}
	return lastErr
}

func (c *Client) callOnce(ctx context.Context, method string, payload any, out any) error {
	url := fmt.Sprintf("%s/bot%s/%s", c.baseURL, c.token, method)

	var body io.Reader = http.NoBody
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}

	var envelope struct {
		OK          bool            `json:"ok"`
		Result      json.RawMessage `json:"result"`
		Description string          `json:"description"`
		ErrorCode   int             `json:"error_code"`
		Parameters  struct {
			RetryAfter int `json:"retry_after"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("unreadable Telegram response (%d)", resp.StatusCode)
	}
	if !envelope.OK {
		code := envelope.ErrorCode
		if code == 0 {
			code = resp.StatusCode
		}
		return &APIError{
			Code:        code,
			Description: envelope.Description,
			RetryAfter:  time.Duration(envelope.Parameters.RetryAfter) * time.Second,
		}
	}
	if out != nil {
		return json.Unmarshal(envelope.Result, out)
	}
	return nil
}

func asAPIError(err error, target **APIError) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*APIError); ok {
		*target = e
		return true
	}
	return false
}
