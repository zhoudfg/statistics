package conf

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestAlibabaResolveShop(t *testing.T) {
	const configYAML = `
alibaba:
  base_url: https://example.test/openapi
  app_key: test-app
  app_secret: test-secret
  default_shop: main
  shops:
    main:
      access_token: main-token
    backup:
      base_url: https://backup.example.test/openapi
      app_key: backup-app
      app_secret: backup-secret
      access_token: backup-token
`
	var config Conf
	if err := yaml.UnmarshalStrict([]byte(configYAML), &config); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		want AlibabaShopConfig
	}{
		{"", AlibabaShopConfig{"https://example.test/openapi", "test-app", "test-secret", "main-token"}},
		{"backup", AlibabaShopConfig{"https://backup.example.test/openapi", "backup-app", "backup-secret", "backup-token"}},
	} {
		shop, err := config.Alibaba.ResolveShop(test.name)
		if err != nil || shop != test.want {
			t.Errorf("ResolveShop(%q) = %+v, %v; want %+v", test.name, shop, err, test.want)
		}
	}

	config.Alibaba.DefaultShop = ""
	if _, err := config.Alibaba.ResolveShop(""); err == nil {
		t.Error("expected multiple shops without a selection to fail")
	}
	if _, err := config.Alibaba.ResolveShop("unknown"); err == nil {
		t.Error("expected an unknown shop to fail")
	}
	delete(config.Alibaba.Shops, "backup")
	shop, err := config.Alibaba.ResolveShop("")
	if err != nil || shop.AccessToken != "main-token" {
		t.Errorf("expected single shop selection, got %v", err)
	}
}

func TestAlibabaResolveShopMissingCredentials(t *testing.T) {
	config := AlibabaConfig{BaseURL: "https://example.test/openapi", AppKey: "test-app", AppSecret: "test-secret"}
	if _, err := config.ResolveShop(""); err == nil || !strings.Contains(err.Error(), "access_token") {
		t.Fatalf("expected missing access_token, got %v", err)
	}
}
