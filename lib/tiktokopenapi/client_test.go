package tiktokopenapi

import (
	"testing"

	"github.com/statistics/lib/conf"
)

func TestSelectShopConfigMatchesShopTypeVersionAndSeller(t *testing.T) {
	tc := conf.TiktokConfig{
		Shops: map[string]conf.TiktokShopConfig{
			"full_v2": {
				ShopType:        1,
				CategoryVersion: "v2",
				SellerType:      "ALL",
				AppKey:          "full-app",
			},
			"half_cross_border_v1": {
				ShopType:        2,
				CategoryVersion: "v1",
				SellerType:      SellerTypeCrossBorder,
				AppKey:          "half-app",
			},
		},
	}

	full := selectShopConfig(tc, ShopOptions{ShopType: 1, CategoryVersion: "v2", SellerType: "ALL"})
	if full.AppKey != "full-app" {
		t.Fatalf("expected full config, got %q", full.AppKey)
	}

	half := selectShopConfig(tc, ShopOptions{ShopType: 2, CategoryVersion: "v1", SellerType: SellerTypeCrossBorder})
	if half.AppKey != "half-app" {
		t.Fatalf("expected half config, got %q", half.AppKey)
	}
}

func TestSelectShopConfigFallsBackToWildcard(t *testing.T) {
	tc := conf.TiktokConfig{
		Shops: map[string]conf.TiktokShopConfig{
			"default_half": {
				ShopType: 2,
				AppKey:   "half-default",
			},
		},
	}

	shop := selectShopConfig(tc, ShopOptions{ShopType: 2, CategoryVersion: "v2", SellerType: SellerTypeLocal})
	if shop.AppKey != "half-default" {
		t.Fatalf("expected wildcard half config, got %q", shop.AppKey)
	}
}
