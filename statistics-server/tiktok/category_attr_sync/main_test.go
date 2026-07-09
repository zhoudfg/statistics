package main

import (
	"database/sql"
	"testing"

	"github.com/statistics/lib/model/erp_model"
	"github.com/statistics/lib/tiktokopenapi"
)

func TestValidateOptionsRequiresHalfSelectors(t *testing.T) {
	err := validateOptions(options{
		shopType:       shopTypeHalf,
		version:        "v1",
		workers:        1,
		batchSize:      1,
		requestTimeout: 1,
	})
	if err == nil {
		t.Fatal("expected missing seller-type error")
	}
}

func TestShopOptionsForFullAlwaysUseFullAllSelector(t *testing.T) {
	path := &erp_model.Temp1TiktokCategoryPath{
		Id:         1,
		CategoryId: "123",
		ShopType:   sql.NullInt64{Int64: shopTypeFull, Valid: true},
		Version:    sql.NullString{String: "v2", Valid: true},
		SellerType: sql.NullString{String: tiktokopenapi.SellerTypeLocal, Valid: true},
	}

	shopOptions, err := shopOptionsForPath(options{}, path, shopTypeFull)
	if err != nil {
		t.Fatal(err)
	}
	if shopOptions.ShopType != shopTypeFull {
		t.Fatalf("expected full shop type, got %d", shopOptions.ShopType)
	}
	if shopOptions.SellerType != "ALL" {
		t.Fatalf("expected ALL seller selector, got %q", shopOptions.SellerType)
	}
	if shopOptions.CategoryVersion != "" {
		t.Fatalf("expected no category version for full config selection, got %q", shopOptions.CategoryVersion)
	}
}

func TestShopOptionsForHalfUseVersionAndSeller(t *testing.T) {
	path := &erp_model.Temp1TiktokCategoryPath{
		Id:         1,
		CategoryId: "123",
		ShopType:   sql.NullInt64{Int64: shopTypeHalf, Valid: true},
		Version:    sql.NullString{String: "v2", Valid: true},
		SellerType: sql.NullString{String: "local", Valid: true},
	}

	shopOptions, err := shopOptionsForPath(options{}, path, shopTypeHalf)
	if err != nil {
		t.Fatal(err)
	}
	if shopOptions.ShopType != shopTypeHalf {
		t.Fatalf("expected half shop type, got %d", shopOptions.ShopType)
	}
	if shopOptions.CategoryVersion != "v2" {
		t.Fatalf("expected v2 category version, got %q", shopOptions.CategoryVersion)
	}
	if shopOptions.SellerType != tiktokopenapi.SellerTypeLocal {
		t.Fatalf("expected LOCAL seller type, got %q", shopOptions.SellerType)
	}
}
