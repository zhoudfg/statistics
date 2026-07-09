package tiktokopenapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/statistics/lib/conf"
)

const (
	DefaultTokenURL       = "https://auth.tiktok-shops.com"
	DefaultUSTokenURL     = "https://auth-us.tiktok-shops.com"
	DefaultNonUSTokenURL  = "https://services.tiktokshop.com"
	DefaultBaseURL        = "https://open-api.tiktokglobalshop.com"
	DefaultVersion        = "202309"
	DefaultFullVersion    = "202405"
	DefaultCategoryAttrV  = "202507"
	DefaultFullCronV      = "202508"
	DefaultFullListV      = "202504"
	DefaultStockOrderV    = "202407"
	DefaultStockOrderConf = "202409"

	SellerTypeCrossBorder = "CROSS_BORDER"
	SellerTypeLocal       = "LOCAL"

	CategoryVersionV1 = "v1"
	CategoryVersionV2 = "v2"

	RegionUS = "US"
	RegionID = "ID"
	RegionSG = "SG"
	RegionMY = "MY"
	RegionVN = "VN"
	RegionTH = "TH"
	RegionPH = "PH"
)

var categoryV2Regions = map[string]struct{}{
	RegionUS: {},
	RegionID: {},
	RegionSG: {},
	RegionMY: {},
	RegionVN: {},
	RegionTH: {},
	RegionPH: {},
}

type Config struct {
	TokenURL                    string
	USTokenURL                  string
	NonUSTokenURL               string
	BaseURL                     string
	Version                     string
	FullVersion                 string
	FullCategoryAttrVersion     string
	FullCronVersion             string
	FullListVersion             string
	FullStockUpOrderVersion     string
	FullStockUpOrderConfVersion string
	Timeout                     time.Duration
	AppKey                      string
	AppSecret                   string
	ServiceID                   string
}

type ShopOptions struct {
	ShopType         int
	ShopSign         string
	PlatformTiktokID string
	AccessToken      string
	RefreshToken     string
	ShopCipher       string
	SellerType       string
	Region           string
	OpenID           string
	CategoryVersion  string
	ShopRegions      []string
	AppKey           string
	AppSecret        string
	ServiceID        string
}

type Client struct {
	httpClient *http.Client

	TokenURL                    string
	USTokenURL                  string
	NonUSTokenURL               string
	BaseURL                     string
	Version                     string
	FullVersion                 string
	FullCategoryAttrVersion     string
	FullCronVersion             string
	FullListVersion             string
	FullStockUpOrderVersion     string
	FullStockUpOrderConfVersion string

	AppKey    string
	AppSecret string
	ServiceID string

	ShopSign         string
	PlatformTiktokID string
	AccessToken      string
	RefreshToken     string
	ShopCipher       string
	SellerType       string
	Region           string
	OpenID           string
	CategoryVersion  string
	ShopRegions      []string
}

type APIError struct {
	StatusCode int
	Code       int64
	Message    string
	RequestID  string
	URL        string
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = "tiktok request failed"
	}
	if e.Code != 0 {
		msg = fmt.Sprintf("%s, code=%d", msg, e.Code)
	}
	if e.RequestID != "" {
		msg = fmt.Sprintf("%s, request_id=%s", msg, e.RequestID)
	}
	if e.StatusCode != 0 {
		msg = fmt.Sprintf("%s, status=%d", msg, e.StatusCode)
	}
	return msg
}

type TokenRefreshData struct {
	AccessToken         string `json:"access_token"`
	RefreshToken        string `json:"refresh_token"`
	AccessTokenExpireIn int64  `json:"access_token_expire_in"`
	RefreshExpiresIn    int64  `json:"refresh_token_expire_in"`
}

type commonResponse struct {
	Code      *int64          `json:"code"`
	Message   string          `json:"message"`
	RequestID string          `json:"request_id"`
	Data      json.RawMessage `json:"data"`
	Error     string          `json:"error"`
}

func NewClient(cfg Config, shop ShopOptions) *Client {
	cfg = withConfigDefaults(cfg)

	if shop.AppKey != "" {
		cfg.AppKey = shop.AppKey
	}
	if shop.AppSecret != "" {
		cfg.AppSecret = shop.AppSecret
	}
	if shop.ServiceID != "" {
		cfg.ServiceID = shop.ServiceID
	}

	categoryVersion := shop.CategoryVersion
	if categoryVersion == "" {
		categoryVersion = CategoryVersionV1
		if isCategoryV2Region(shop.Region) {
			categoryVersion = CategoryVersionV2
		}
	}

	sellerType := shop.SellerType
	if sellerType == "" {
		sellerType = SellerTypeCrossBorder
	}
	region := shop.Region
	if region == "" {
		region = "ALL"
	}

	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},

		TokenURL:                    strings.TrimRight(cfg.TokenURL, "/"),
		USTokenURL:                  strings.TrimRight(cfg.USTokenURL, "/"),
		NonUSTokenURL:               strings.TrimRight(cfg.NonUSTokenURL, "/"),
		BaseURL:                     strings.TrimRight(cfg.BaseURL, "/"),
		Version:                     cfg.Version,
		FullVersion:                 cfg.FullVersion,
		FullCategoryAttrVersion:     cfg.FullCategoryAttrVersion,
		FullCronVersion:             cfg.FullCronVersion,
		FullListVersion:             cfg.FullListVersion,
		FullStockUpOrderVersion:     cfg.FullStockUpOrderVersion,
		FullStockUpOrderConfVersion: cfg.FullStockUpOrderConfVersion,
		AppKey:                      cfg.AppKey,
		AppSecret:                   cfg.AppSecret,
		ServiceID:                   cfg.ServiceID,
		ShopSign:                    shop.ShopSign,
		PlatformTiktokID:            shop.PlatformTiktokID,
		AccessToken:                 shop.AccessToken,
		RefreshToken:                shop.RefreshToken,
		ShopCipher:                  shop.ShopCipher,
		SellerType:                  sellerType,
		Region:                      region,
		OpenID:                      shop.OpenID,
		CategoryVersion:             categoryVersion,
		ShopRegions:                 append([]string(nil), shop.ShopRegions...),
	}
}

func NewClientFromConf(shop ShopOptions) *Client {
	tc := conf.GetTiktok()
	shopConf := selectShopConfig(tc, shop)
	cfg := Config{
		TokenURL:                    firstNonEmpty(shopConf.TokenURL, tc.TokenURL),
		USTokenURL:                  firstNonEmpty(shopConf.USTokenURL, tc.USTokenURL),
		NonUSTokenURL:               firstNonEmpty(shopConf.NonUSTokenURL, tc.NonUSTokenURL),
		BaseURL:                     firstNonEmpty(shopConf.BaseURL, tc.BaseURL),
		Version:                     firstNonEmpty(shopConf.Version, tc.Version),
		FullVersion:                 firstNonEmpty(shopConf.FullVersion, tc.FullVersion),
		FullCategoryAttrVersion:     firstNonEmpty(shopConf.FullCategoryAttrVersion, tc.FullCategoryAttrVersion),
		FullCronVersion:             firstNonEmpty(shopConf.FullCronVersion, tc.FullCronVersion),
		FullListVersion:             firstNonEmpty(shopConf.FullListVersion, tc.FullListVersion),
		FullStockUpOrderVersion:     firstNonEmpty(shopConf.FullStockUpOrderVersion, tc.FullStockUpOrderVersion),
		FullStockUpOrderConfVersion: firstNonEmpty(shopConf.FullStockUpOrderConfVersion, tc.FullStockUpOrderConfVersion),
		AppKey:                      firstNonEmpty(shopConf.AppKey, tc.AppKey),
		AppSecret:                   firstNonEmpty(shopConf.AppSecret, tc.AppSecret),
		ServiceID:                   firstNonEmpty(shopConf.ServiceID, tc.ServiceID),
	}
	if shop.ShopSign == "" {
		shop.ShopSign = firstNonEmpty(shopConf.ShopSign, tc.ShopSign)
	}
	if shop.PlatformTiktokID == "" {
		shop.PlatformTiktokID = firstNonEmpty(shopConf.PlatformTiktokID, tc.PlatformTiktokID)
	}
	if shop.AccessToken == "" {
		shop.AccessToken = firstNonEmpty(shopConf.AccessToken, tc.AccessToken)
	}
	if shop.RefreshToken == "" {
		shop.RefreshToken = firstNonEmpty(shopConf.RefreshToken, tc.RefreshToken)
	}
	if shop.ShopCipher == "" {
		shop.ShopCipher = firstNonEmpty(shopConf.ShopCipher, tc.ShopCipher)
	}
	if shop.SellerType == "" {
		shop.SellerType = firstNonEmpty(shopConf.SellerType, tc.SellerType)
	}
	if shop.Region == "" {
		shop.Region = firstNonEmpty(shopConf.Region, tc.Region)
	}
	if shop.OpenID == "" {
		shop.OpenID = firstNonEmpty(shopConf.OpenID, tc.OpenID)
	}
	if shop.CategoryVersion == "" {
		shop.CategoryVersion = firstNonEmpty(shopConf.CategoryVersion, tc.CategoryVersion)
	}
	if shop.AppKey == "" {
		shop.AppKey = firstNonEmpty(shopConf.AppKey, tc.AppKey)
	}
	if shop.AppSecret == "" {
		shop.AppSecret = firstNonEmpty(shopConf.AppSecret, tc.AppSecret)
	}
	if shop.ServiceID == "" {
		shop.ServiceID = firstNonEmpty(shopConf.ServiceID, tc.ServiceID)
	}
	return NewClient(cfg, shop)
}

func selectShopConfig(tc conf.TiktokConfig, shop ShopOptions) conf.TiktokShopConfig {
	if len(tc.Shops) == 0 {
		return conf.TiktokShopConfig{}
	}

	keys := make([]string, 0, len(tc.Shops))
	for key := range tc.Shops {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	bestScore := -1
	var best conf.TiktokShopConfig
	for _, key := range keys {
		candidate := tc.Shops[key]
		score, ok := shopConfigMatchScore(candidate, shop)
		if !ok {
			continue
		}
		if score > bestScore {
			bestScore = score
			best = candidate
		}
	}
	return best
}

func shopConfigMatchScore(candidate conf.TiktokShopConfig, shop ShopOptions) (int, bool) {
	score := 0
	if candidate.ShopType != 0 {
		if shop.ShopType == 0 || candidate.ShopType != shop.ShopType {
			return 0, false
		}
		score += 4
	}
	if candidate.CategoryVersion != "" {
		if shop.CategoryVersion == "" || !strings.EqualFold(candidate.CategoryVersion, shop.CategoryVersion) {
			return 0, false
		}
		score += 2
	}
	if candidate.SellerType != "" {
		if shop.SellerType == "" || !strings.EqualFold(candidate.SellerType, shop.SellerType) {
			return 0, false
		}
		score++
	}
	if candidate.Region != "" && shop.Region != "" && strings.EqualFold(candidate.Region, shop.Region) {
		score++
	}
	return score, true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (c *Client) NewCategoryService() *CategoryService {
	return NewCategoryService(c)
}

func (c *Client) BaseQueries(locale string) map[string]string {
	return map[string]string{
		"app_key":    c.AppKey,
		"app_secret": c.AppSecret,
		"timestamp":  strconv.FormatInt(time.Now().Unix(), 10),
		"locale":     locale,
	}
}

func (c *Client) CreateSign(path string, queries map[string]string, body []byte) url.Values {
	keys := make([]string, 0, len(queries))
	for key := range queries {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	input := path
	for _, key := range keys {
		input += key + queries[key]
	}
	if len(body) > 0 {
		input = c.AppSecret + input + string(body) + c.AppSecret
	} else {
		input = c.AppSecret + input + c.AppSecret
	}

	hash := hmac.New(sha256.New, []byte(c.AppSecret))
	_, _ = hash.Write([]byte(input))

	values := make(url.Values, len(queries)+1)
	for key, value := range queries {
		values.Set(key, value)
	}
	values.Set("sign", hex.EncodeToString(hash.Sum(nil)))
	return values
}

func (c *Client) Request(ctx context.Context, method, path string, queries map[string]string, body interface{}) (json.RawMessage, error) {
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	headers := map[string]string{
		"Content-Type":       "application/json",
		"x-tts-access-token": c.AccessToken,
	}
	fullURL := c.BaseURL + path
	signedQueries := c.CreateSign(path, queries, bodyBytes)

	return c.requestJSON(ctx, method, fullURL, headers, signedQueries, bodyBytes)
}

func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenRefreshData, error) {
	if refreshToken == "" {
		refreshToken = c.RefreshToken
	}
	values := url.Values{}
	values.Set("app_key", c.AppKey)
	values.Set("app_secret", c.AppSecret)
	values.Set("refresh_token", refreshToken)
	values.Set("grant_type", "refresh_token")

	raw, err := c.requestRaw(ctx, http.MethodGet, c.TokenURL+"/api/v2/token/refresh", nil, values, nil)
	if err != nil {
		return nil, err
	}

	var resp commonResponse
	if err := json.Unmarshal(raw, &resp); err == nil && resp.Code != nil {
		if *resp.Code != 0 {
			return nil, c.makeAPIError(0, resp, c.TokenURL+"/api/v2/token/refresh")
		}
		var data TokenRefreshData
		if len(resp.Data) > 0 {
			if err := json.Unmarshal(resp.Data, &data); err != nil {
				return nil, err
			}
		}
		if data.AccessToken != "" {
			c.AccessToken = data.AccessToken
		}
		if data.RefreshToken != "" {
			c.RefreshToken = data.RefreshToken
		}
		return &data, nil
	}

	var data TokenRefreshData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	if data.AccessToken != "" {
		c.AccessToken = data.AccessToken
	}
	if data.RefreshToken != "" {
		c.RefreshToken = data.RefreshToken
	}
	return &data, nil
}

func (c *Client) requestJSON(ctx context.Context, method, rawURL string, headers map[string]string, queries url.Values, body []byte) (json.RawMessage, error) {
	raw, err := c.requestRaw(ctx, method, rawURL, headers, queries, body)
	if err != nil {
		return nil, err
	}

	var resp commonResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	if resp.Code == nil || *resp.Code != 0 {
		if resp.Code != nil && *resp.Code == 105001 && c.RefreshToken != "" {
			_, _ = c.RefreshAccessToken(ctx, c.RefreshToken)
		}
		return nil, c.makeAPIError(0, resp, rawURL)
	}
	if len(resp.Data) == 0 || string(resp.Data) == "null" {
		return json.RawMessage("{}"), nil
	}
	return resp.Data, nil
}

func (c *Client) requestRaw(ctx context.Context, method, rawURL string, headers map[string]string, queries url.Values, body []byte) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	query := u.Query()
	for key, values := range queries {
		for _, value := range values {
			query.Set(key, value)
		}
	}
	u.RawQuery = query.Encode()

	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		var common commonResponse
		_ = json.Unmarshal(raw, &common)
		return nil, c.makeAPIError(resp.StatusCode, common, u.String())
	}
	return raw, nil
}

func (c *Client) makeAPIError(statusCode int, resp commonResponse, rawURL string) *APIError {
	var code int64
	if resp.Code != nil {
		code = *resp.Code
	}
	message := errorDescription(code)
	if resp.Message != "" {
		if message != "" {
			message += ": " + resp.Message
		} else {
			message = resp.Message
		}
	}
	if message == "" {
		message = resp.Error
	}
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		RequestID:  resp.RequestID,
		URL:        rawURL,
	}
}

func (c *Client) shouldUseCategoryV2() bool {
	if isCategoryV2Region(c.Region) {
		return true
	}
	for _, region := range c.ShopRegions {
		if isCategoryV2Region(region) {
			return true
		}
	}
	return false
}

func isCategoryV2Region(region string) bool {
	_, ok := categoryV2Regions[strings.ToUpper(region)]
	return ok
}

func withConfigDefaults(cfg Config) Config {
	if cfg.TokenURL == "" {
		cfg.TokenURL = DefaultTokenURL
	}
	if cfg.USTokenURL == "" {
		cfg.USTokenURL = DefaultUSTokenURL
	}
	if cfg.NonUSTokenURL == "" {
		cfg.NonUSTokenURL = DefaultNonUSTokenURL
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.Version == "" {
		cfg.Version = DefaultVersion
	}
	if cfg.FullVersion == "" {
		cfg.FullVersion = DefaultFullVersion
	}
	if cfg.FullCategoryAttrVersion == "" {
		cfg.FullCategoryAttrVersion = DefaultCategoryAttrV
	}
	if cfg.FullCronVersion == "" {
		cfg.FullCronVersion = DefaultFullCronV
	}
	if cfg.FullListVersion == "" {
		cfg.FullListVersion = DefaultFullListV
	}
	if cfg.FullStockUpOrderVersion == "" {
		cfg.FullStockUpOrderVersion = DefaultStockOrderV
	}
	if cfg.FullStockUpOrderConfVersion == "" {
		cfg.FullStockUpOrderConfVersion = DefaultStockOrderConf
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 120 * time.Second
	}
	return cfg
}

func errorDescription(code int64) string {
	descriptions := map[int64]string{
		101000:   "invalid parameter",
		101001:   "missing required parameter",
		101002:   "failed to parse parameter",
		101003:   "request length exceeds limit",
		101004:   "unsupported HTTP method",
		102000:   "internal service error",
		102001:   "request timeout",
		103000:   "app request quota exceeded",
		103001:   "request frequency too high",
		103002:   "api quota exceeded",
		104000:   "missing app_key",
		104001:   "invalid app_key",
		104003:   "access token does not match app_key",
		105000:   "missing access token",
		105001:   "invalid access token",
		105002:   "access token expired",
		105004:   "inactive authorization status",
		105005:   "api authorization required",
		106000:   "missing signature",
		106001:   "invalid signature",
		106003:   "missing timestamp",
		106004:   "invalid or expired timestamp",
		106005:   "invalid shop id",
		106007:   "invalid path or method",
		106008:   "traffic routing error",
		106011:   "invalid shop_cipher",
		106012:   "shop_cipher should not be sent to this api",
		106013:   "missing shop_cipher",
		199000:   "unknown error",
		36002010: "system internal error",
		36009002: "rate limit exceeded",
		36009003: "system internal error",
		36009004: "invalid parameter",
		36009007: "system internal timeout",
		36009010: "unsupported HTTP method",
		36009020: "system internal error",
		98001001: "system internal error",
		98001004: "invalid parameter",
		98001021: "system internal error",
		99999999: "system internal error",
	}
	return descriptions[code]
}
