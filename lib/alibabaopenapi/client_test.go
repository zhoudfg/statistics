package alibabaopenapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestGetPublishingRuleForms(t *testing.T) {
	tests := []struct {
		scene        string
		wantScene    string
		categoryType int
		categoryID   int64
		form         url.Values
	}{
		{"cbu", "cbu", 1, 123, url.Values{
			"catId":          {"123"},
			"_aop_signature": {"061D654097E3E7B371B6A552E4F2E74B0926CEA8"},
		}},
		{"processing", "cbu", 1, 123, url.Values{
			"catId":          {"123"},
			"_aop_signature": {"061D654097E3E7B371B6A552E4F2E74B0926CEA8"},
		}},
		{"industry", "industry", 2, 9223372036854775807, url.Values{
			"bizParam":       {`{"industryCategoryId":9223372036854775807}`},
			"_aop_signature": {"C1278FDFEC4500131CE3100CE2B06489BA9239F1"},
		}},
		{"popular", "popular", 3, 456, url.Values{
			"catId":          {"456"},
			"_aop_signature": {"16B5FB7FA8F9905B219BD48E81A02F958CE0A0CB"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.scene, func(t *testing.T) {
			tt.form.Set("scene", tt.wantScene)
			tt.form.Set("access_token", "test-token")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				wantPath := "/openapi/param2/1/com.alibaba.product/alibaba.new.product.getSchema/test-app"
				if req.Method != http.MethodPost || req.URL.Path != wantPath || req.URL.RawQuery != "" {
					t.Errorf("unexpected request target: %s %s", req.Method, req.URL)
				}
				if got := req.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
					t.Errorf("unexpected content type: %q", got)
				}
				if err := req.ParseForm(); err != nil {
					t.Errorf("ParseForm: %v", err)
				}
				if !reflect.DeepEqual(req.PostForm, tt.form) {
					t.Errorf("form = %v, want %v", req.PostForm, tt.form)
				}
				io.WriteString(w, `{"result":{"success":true,"bizData":"{\"rules\":[1]}"}}`)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL+"/openapi/")
			rule, err := client.GetPublishingRule(context.Background(), tt.categoryType, tt.categoryID, tt.scene)
			if err != nil || string(rule) != `{"rules":[1]}` {
				t.Fatalf("rule = %q, err = %v", rule, err)
			}
		})
	}
}

func TestGetPublishingRuleFailures(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		body         string
		retryable    bool
		unauthorized bool
	}{
		{"timeout", 408, "", true, false},
		{"rate limit", 429, "", true, false},
		{"server failure", 503, "", true, false},
		{"authorization failure", 401, "", false, true},
		{"permission failure", 403, "", false, false},
		{"platform rate limit", 200, `{"error_code":429}`, true, false},
		{"expired access token", 200, `{"error":{"code":"access_token_expired","message":"test-token"}}`, false, true},
		{"business failure", 200, `{"result":{"success":false,"bizCode":"CATEGORY_INVALID","bizMsg":"test-secret test-token"}}`, false, false},
		{"missing rules", 200, `{"result":{"success":true,"bizData":null}}`, false, false},
		{"invalid rules", 200, `{"result":{"success":true,"bizData":"broken"}}`, false, false},
		{"malformed response", 200, `{"result"`, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(tt.status)
				io.WriteString(w, tt.body)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL)
			_, err := client.GetPublishingRule(context.Background(), 1, 123, "cbu")
			if err == nil || IsRetryable(err) != tt.retryable || IsUnauthorized(err) != tt.unauthorized {
				t.Fatalf("error = %v, retryable = %t, unauthorized = %t", err, IsRetryable(err), IsUnauthorized(err))
			}
			for _, secret := range []string{"test-secret", "test-token", "test-app"} {
				if strings.Contains(err.Error(), secret) {
					t.Errorf("error exposes request credentials: %v", err)
				}
			}
		})
	}
}

func TestGetPublishingRuleDoesNotExposeSignature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, `{"result":{"success":false,"bizCode":%q}}`, req.FormValue("_aop_signature"))
	}))
	defer server.Close()
	client := newTestClient(t, server.URL)
	_, err := client.GetPublishingRule(context.Background(), 1, 123, "cbu")
	if err == nil || err.Error() != "Alibaba rule request rejected" {
		t.Fatalf("error = %v", err)
	}
}

func TestGetPublishingRuleReportsSafeHTTPErrorDetails(t *testing.T) {
	for _, tt := range []struct {
		name, body, requestID, want string
	}{
		{"platform code", `{"errorCode":"INVALID_SIGNATURE","message":"test-token","requestId":"req-1688"}`, "ignored", "Alibaba HTTP status 401 code=INVALID_SIGNATURE request_id=req-1688"},
		{"nested code", `{"error":{"code":"invalid_token"}}`, "trace-123", "Alibaba HTTP status 401 code=invalid_token request_id=trace-123"},
		{"redacted fields", `{"error_code":"test-secret","request_id":"test-token"}`, "trace-123", "Alibaba HTTP status 401 request_id=trace-123"},
		{"non-JSON body", "<html>test-token</html>", "", "Alibaba HTTP status 401"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("X-Request-Id", tt.requestID)
				w.WriteHeader(http.StatusUnauthorized)
				io.WriteString(w, tt.body)
			}))
			defer server.Close()
			client := newTestClient(t, server.URL)
			_, err := client.GetPublishingRule(context.Background(), 2, 123, "industry")
			if !IsUnauthorized(err) || IsRetryable(err) || err.Error() != tt.want {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestGetPublishingRulePreservesRedactedBizMessageForStorage(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"business failure", http.StatusOK, `{"result":{"success":false,"bizCode":"CATEGORY_INVALID","bizMsg":%q}}`},
		{"HTTP failure with nested bizMsg", http.StatusBadRequest, `{"result":{"success":false,"bizMsg":%q}}`},
		{"HTTP failure with top-level bizMsg", http.StatusBadRequest, `{"bizMsg":%q}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				message := "类目无效\n凭据 test-app test-secret test-token " + req.FormValue("_aop_signature")
				w.WriteHeader(tt.status)
				fmt.Fprintf(w, tt.body, message)
			}))
			defer server.Close()
			client := newTestClient(t, server.URL)
			_, err := client.GetPublishingRule(context.Background(), 1, 123, "cbu")
			if err == nil {
				t.Fatal("expected an error")
			}
			if strings.Contains(err.Error(), "类目无效") || strings.Contains(err.Error(), "bizMsg") {
				t.Fatalf("ordinary error exposes bizMsg: %q", err.Error())
			}
			want := err.Error() + ` bizMsg="类目无效\n凭据 [REDACTED] [REDACTED] [REDACTED] [REDACTED]"`
			if got := ErrorForStorage(err); got != want {
				t.Fatalf("ErrorForStorage = %q, want %q", got, want)
			}
		})
	}
}

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	client, err := NewClient(Config{BaseURL: baseURL, AppKey: "test-app", AppSecret: "test-secret", AccessToken: "test-token"})
	if err != nil {
		t.Fatal(err)
	}
	return client
}
