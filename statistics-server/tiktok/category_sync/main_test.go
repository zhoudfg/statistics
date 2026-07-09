package main

import (
	"encoding/json"
	"testing"
)

func TestBuildHalfCategoryRowZhCN(t *testing.T) {
	row := buildHalfCategoryRow(halfCategory{
		ID:                 "600002",
		ParentID:           "600001",
		LocalName:          "\u5bb6\u5c45\u7528\u54c1",
		IsLeaf:             false,
		PermissionStatuses: json.RawMessage(`["INVITE_ONLY","NON_MAIN_CATEGORY"]`),
	}, options{
		version:    "v2",
		sellerType: "LOCAL",
		locale:     localeZhCN,
	})

	if row.NameCn != "\u5bb6\u5c45\u7528\u54c1" {
		t.Fatalf("expected Chinese name, got %q", row.NameCn)
	}
	if row.NameEn != "" {
		t.Fatalf("expected empty English name, got %q", row.NameEn)
	}
	if row.IsLeaf != 0 {
		t.Fatalf("expected is_leaf=0, got %d", row.IsLeaf)
	}
	if string(row.PermissionStatuses.([]byte)) != `["INVITE_ONLY","NON_MAIN_CATEGORY"]` {
		t.Fatalf("unexpected permission statuses: %v", row.PermissionStatuses)
	}
}

func TestBuildHalfCategoryRowEnUS(t *testing.T) {
	row := buildHalfCategoryRow(halfCategory{
		ID:        "600002",
		ParentID:  "600001",
		LocalName: "Home Supplies",
		IsLeaf:    true,
	}, options{
		version:    "v1",
		sellerType: "CROSS_BORDER",
		locale:     localeEnUS,
	})

	if row.NameCn != "" {
		t.Fatalf("expected empty Chinese name, got %q", row.NameCn)
	}
	if row.NameEn != "Home Supplies" {
		t.Fatalf("expected English name, got %q", row.NameEn)
	}
	if row.IsLeaf != 1 {
		t.Fatalf("expected is_leaf=1, got %d", row.IsLeaf)
	}
}

func TestBuildFullCategoryRow(t *testing.T) {
	row := buildFullCategoryRow(fullCategory{
		ID:       "7345257688508401429",
		NameCn:   "\u73e0\u5b9d\u4e0e\u884d\u751f\u54c1",
		NameEn:   "Jewellery Accessories & Derivatives",
		ParentID: "0",
		IsLeaf:   false,
		Status:   json.RawMessage(`1`),
		Version:  "v2",
	}, options{})

	if row.ShopType != shopTypeFull {
		t.Fatalf("expected full shop type, got %d", row.ShopType)
	}
	if row.SellerType != "ALL" {
		t.Fatalf("expected ALL seller type, got %q", row.SellerType)
	}
	if row.Status.String != "1" || !row.Status.Valid {
		t.Fatalf("expected status=1, got %+v", row.Status)
	}
	if row.Version != "v2" {
		t.Fatalf("expected version=v2, got %q", row.Version)
	}
}

func TestNormalizeLocale(t *testing.T) {
	if got := normalizeLocale("zh_cn"); got != localeZhCN {
		t.Fatalf("expected zh-CN, got %q", got)
	}
	if got := normalizeLocale("en_us"); got != localeEnUS {
		t.Fatalf("expected en-US, got %q", got)
	}
}
