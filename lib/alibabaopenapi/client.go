package alibabaopenapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const publishingRulePath = "param2/1/com.alibaba.product/alibaba.new.product.getSchema"

var errorFieldPattern = regexp.MustCompile(`^[[:alnum:]_.-]{1,64}$`)

type Config struct {
	BaseURL     string
	AppKey      string
	AppSecret   string
	AccessToken string
	Timeout     time.Duration
}

type Client struct {
	endpoint    string
	apiPath     string
	appKey      string
	appSecret   string
	accessToken string
	httpClient  *http.Client
}

type requestError struct {
	message      string
	bizMsg       string
	retryable    bool
	unauthorized bool
}

func (e *requestError) Error() string { return e.message }

// ErrorForStorage includes the platform's redacted business message, when present.
func ErrorForStorage(err error) string {
	if err == nil {
		return ""
	}
	var requestErr *requestError
	if errors.As(err, &requestErr) && requestErr.bizMsg != "" {
		return fmt.Sprintf("%s bizMsg=%q", err, requestErr.bizMsg)
	}
	return err.Error()
}

func IsRetryable(err error) bool {
	var requestErr *requestError
	return errors.As(err, &requestErr) && requestErr.retryable
}

func IsUnauthorized(err error) bool {
	var requestErr *requestError
	return errors.As(err, &requestErr) && requestErr.unauthorized
}

func NewClient(cfg Config) (*Client, error) {
	baseURL, err := url.Parse(strings.TrimSpace(cfg.BaseURL))
	if err != nil || baseURL.Hostname() == "" || baseURL.User != nil || baseURL.ForceQuery || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return nil, errors.New("invalid Alibaba base_url")
	}
	if baseURL.Scheme != "https" && (baseURL.Scheme != "http" || !isLoopback(baseURL.Hostname())) {
		return nil, errors.New("Alibaba base_url requires HTTPS or loopback HTTP")
	}
	if strings.TrimSpace(cfg.AppKey) == "" || strings.TrimSpace(cfg.AppSecret) == "" || strings.TrimSpace(cfg.AccessToken) == "" {
		return nil, errors.New("Alibaba app_key, app_secret and access_token are required")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 60 * time.Second
	}

	return &Client{
		endpoint:    strings.TrimRight(baseURL.String(), "/") + "/" + publishingRulePath + "/" + url.PathEscape(cfg.AppKey),
		apiPath:     publishingRulePath + "/" + cfg.AppKey,
		appKey:      cfg.AppKey,
		appSecret:   cfg.AppSecret,
		accessToken: cfg.AccessToken,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

func isLoopback(host string) bool {
	return host == "127.0.0.1" || host == "::1" || strings.EqualFold(host, "localhost")
}

// NormalizeScene maps the processing alias to the 1688 cbu scene.
func NormalizeScene(scene string) string {
	if scene == "processing" {
		return "cbu"
	}
	return scene
}

// GetPublishingRule returns the JSON contained in the 1688 bizData field.
func (c *Client) GetPublishingRule(ctx context.Context, categoryType int, categoryID int64, scene string) (json.RawMessage, error) {
	if categoryID <= 0 {
		return nil, errors.New("Alibaba category_id must be positive")
	}
	scene = NormalizeScene(scene)
	switch scene {
	case "cbu", "industry", "popular":
	default:
		return nil, fmt.Errorf("unsupported Alibaba scene %q", scene)
	}

	params := map[string]string{"scene": scene, "access_token": c.accessToken}
	categoryIDText := strconv.FormatInt(categoryID, 10)
	switch categoryType {
	case 1, 3:
		params["catId"] = categoryIDText
	case 2:
		params["bizParam"] = `{"industryCategoryId":` + categoryIDText + `}`
	default:
		return nil, fmt.Errorf("unsupported Alibaba category_type %d", categoryType)
	}

	signature := createSign(c.apiPath, params, c.appSecret)
	form := url.Values{"_aop_signature": {signature}}
	for key, value := range params {
		form.Set(key, value)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, errors.New("could not build Alibaba publishing rule request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, transportError(err, "send Alibaba publishing rule request")
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, c.httpError(resp, signature)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, transportError(err, "read Alibaba publishing rule response")
	}
	return c.parsePublishingRule(body, signature)
}

func (c *Client) httpError(resp *http.Response, signature string) error {
	status := resp.StatusCode
	result := &requestError{
		message:      fmt.Sprintf("Alibaba HTTP status %d", status),
		retryable:    status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || status >= http.StatusInternalServerError,
		unauthorized: status == http.StatusUnauthorized,
	}
	// Upstream errors can echo the signed form. Extract only structured diagnostic fields.
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 16<<10))
	var payload map[string]json.RawMessage
	_ = json.Unmarshal(body, &payload)
	var nested map[string]json.RawMessage
	_ = json.Unmarshal(payload["error"], &nested)
	for _, code := range []json.RawMessage{payload["error_code"], payload["errorCode"], nested["code"], payload["code"], payload["error"]} {
		if value := responseCode(code); value != "0" && c.safeErrorField(value, signature) {
			result.message += " code=" + value
			break
		}
	}
	for _, id := range []string{responseCode(payload["request_id"]), responseCode(payload["requestId"]), resp.Header.Get("X-Request-Id")} {
		if c.safeErrorField(id, signature) {
			result.message += " request_id=" + id
			break
		}
	}
	var ruleResult map[string]json.RawMessage
	_ = json.Unmarshal(payload["result"], &ruleResult)
	result.bizMsg = c.safeBizMessage(ruleResult["bizMsg"], signature)
	if result.bizMsg == "" {
		result.bizMsg = c.safeBizMessage(payload["bizMsg"], signature)
	}
	return result
}

func transportError(err error, operation string) error {
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		err = urlErr.Err
	}
	var netErr net.Error
	var opErr *net.OpError
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return &requestError{message: operation + ": timeout", retryable: true}
	}
	return &requestError{
		message:   operation + ": transport error",
		retryable: errors.As(err, &opErr) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF),
	}
}

func createSign(apiPath string, params map[string]string, secret string) string {
	parts := make([]string, 0, len(params))
	for key, value := range params {
		parts = append(parts, key+value)
	}
	sort.Strings(parts)
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(apiPath + strings.Join(parts, "")))
	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
}

func (c *Client) parsePublishingRule(body []byte, signature string) (json.RawMessage, error) {
	var response struct {
		ErrorCode json.RawMessage `json:"error_code"`
		BizMsg    json.RawMessage `json:"bizMsg"`
		Error     *struct {
			Code json.RawMessage `json:"code"`
		} `json:"error"`
		Result *struct {
			Success bool            `json:"success"`
			BizCode json.RawMessage `json:"bizCode"`
			BizMsg  json.RawMessage `json:"bizMsg"`
			BizData json.RawMessage `json:"bizData"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.New("invalid Alibaba response JSON")
	}
	bizMsg := response.BizMsg
	if response.Result != nil && len(response.Result.BizMsg) > 0 {
		bizMsg = response.Result.BizMsg
	}
	if code := responseCode(response.ErrorCode); code != "" && code != "0" {
		return nil, c.platformError(code, bizMsg, signature)
	}
	if response.Error != nil {
		return nil, c.platformError(responseCode(response.Error.Code), bizMsg, signature)
	}
	if response.Result == nil {
		return nil, errors.New("Alibaba response is missing result")
	}
	if !response.Result.Success {
		return nil, c.platformError(responseCode(response.Result.BizCode), bizMsg, signature)
	}

	data := response.Result.BizData
	if len(data) > 0 && data[0] == '"' {
		var jsonText string
		if err := json.Unmarshal(data, &jsonText); err != nil {
			return nil, errors.New("invalid Alibaba bizData string")
		}
		data = json.RawMessage(jsonText)
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 || (data[0] != '{' && data[0] != '[') || !json.Valid(data) {
		return nil, errors.New("Alibaba response is missing valid rule JSON in bizData")
	}
	return data, nil
}

func responseCode(code json.RawMessage) string {
	if len(code) == 0 || string(code) == "null" {
		return ""
	}
	if code[0] == '"' {
		var value string
		if json.Unmarshal(code, &value) == nil {
			return value
		}
	}
	return string(code)
}

func (c *Client) platformError(code string, bizMsg json.RawMessage, signature string) error {
	err := &requestError{message: "Alibaba rule request rejected", bizMsg: c.safeBizMessage(bizMsg, signature)}
	// The platform message may echo signed request parameters, so only a safe code is logged.
	if c.safeErrorField(code, signature) {
		err.message += " (code=" + code + ")"
	}
	normalizedCode := strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(code))
	switch normalizedCode {
	case "401", "invalidtoken", "invalidaccesstoken", "accesstokeninvalid", "expiredaccesstoken", "accesstokenexpired":
		err.unauthorized = true
	case "408", "429", "500", "502", "503", "504", "toomanyrequests", "ratelimitexceeded", "requestlimitexceeded", "systembusy", "throttling":
		err.retryable = true
	}
	return err
}

func (c *Client) safeErrorField(value, signature string) bool {
	return errorFieldPattern.MatchString(value) && !strings.Contains(value, c.accessToken) && !strings.Contains(value, c.appKey) &&
		!strings.Contains(value, c.appSecret) && !strings.Contains(strings.ToUpper(value), signature)
}

func (c *Client) safeBizMessage(raw json.RawMessage, signature string) string {
	var message string
	if json.Unmarshal(raw, &message) != nil || strings.TrimSpace(message) == "" {
		return ""
	}
	var redactions []string
	for _, secret := range []string{c.accessToken, c.appSecret, c.appKey, signature, strings.ToLower(signature)} {
		if secret != "" {
			redactions = append(redactions, secret, "[REDACTED]")
			if encoded := url.QueryEscape(secret); encoded != secret {
				redactions = append(redactions, encoded, "[REDACTED]")
			}
		}
	}
	return strings.NewReplacer(redactions...).Replace(strings.TrimSpace(message))
}
