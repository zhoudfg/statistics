package main

import "testing"

func TestBuildPathRowsLeafOnly(t *testing.T) {
	opt := options{separator: " > ", status: 0}
	nodes := []categoryNode{
		{CategoryID: "1", ParentID: "0", IsLeaf: 0, NameCn: "\u6839", NameEn: "Root", ShopType: shopTypeHalf, Version: "v2", SellerType: "LOCAL"},
		{CategoryID: "2", ParentID: "1", IsLeaf: 0, NameCn: "\u4e8c\u7ea7", NameEn: "Second", ShopType: shopTypeHalf, Version: "v2", SellerType: "LOCAL"},
		{CategoryID: "3", ParentID: "2", IsLeaf: 1, NameCn: "\u53f6\u5b50", NameEn: "Leaf", ShopType: shopTypeHalf, Version: "v2", SellerType: "LOCAL"},
	}

	rows, stat := buildPathRows(nodes, opt)
	if len(rows) != 1 {
		t.Fatalf("expected 1 path row, got %d", len(rows))
	}
	if stat.leaf != 1 || stat.built != 1 {
		t.Fatalf("unexpected stats: %+v", stat)
	}
	got := rows[0]
	if got.CategoryID != "3" {
		t.Fatalf("unexpected category id: %q", got.CategoryID)
	}
	if got.CategoryPathID != "1 > 2 > 3" {
		t.Fatalf("unexpected id path: %q", got.CategoryPathID)
	}
	if got.CategoryPathCn != "\u6839 > \u4e8c\u7ea7 > \u53f6\u5b50" {
		t.Fatalf("unexpected cn path: %q", got.CategoryPathCn)
	}
	if got.CategoryPathEn != "Root > Second > Leaf" {
		t.Fatalf("unexpected en path: %q", got.CategoryPathEn)
	}
}

func TestBuildPathRowsGroupIsolation(t *testing.T) {
	opt := options{separator: " > ", status: 0}
	nodes := []categoryNode{
		{CategoryID: "1", ParentID: "0", IsLeaf: 0, NameCn: "A", ShopType: shopTypeHalf, Version: "v1", SellerType: "LOCAL"},
		{CategoryID: "2", ParentID: "1", IsLeaf: 1, NameCn: "B", ShopType: shopTypeHalf, Version: "v1", SellerType: "LOCAL"},
		{CategoryID: "1", ParentID: "0", IsLeaf: 0, NameCn: "X", ShopType: shopTypeHalf, Version: "v2", SellerType: "LOCAL"},
		{CategoryID: "2", ParentID: "1", IsLeaf: 1, NameCn: "Y", ShopType: shopTypeHalf, Version: "v2", SellerType: "LOCAL"},
	}

	rows, stat := buildPathRows(nodes, opt)
	if len(rows) != 2 {
		t.Fatalf("expected 2 path rows, got %d", len(rows))
	}
	if stat.built != 2 {
		t.Fatalf("expected built=2, got %+v", stat)
	}
	if rows[0].CategoryPathCn != "A > B" {
		t.Fatalf("unexpected v1 path: %q", rows[0].CategoryPathCn)
	}
	if rows[1].CategoryPathCn != "X > Y" {
		t.Fatalf("unexpected v2 path: %q", rows[1].CategoryPathCn)
	}
}

func TestBuildPathRowsSkipMissingParent(t *testing.T) {
	opt := options{separator: " > ", skipMissingParent: true}
	nodes := []categoryNode{
		{CategoryID: "3", ParentID: "2", IsLeaf: 1, NameCn: "\u53f6\u5b50", ShopType: shopTypeHalf, Version: "v2", SellerType: "LOCAL"},
	}

	rows, stat := buildPathRows(nodes, opt)
	if len(rows) != 0 {
		t.Fatalf("expected no rows, got %d", len(rows))
	}
	if stat.missingParent != 1 {
		t.Fatalf("expected missing_parent=1, got %+v", stat)
	}
}

func TestBuildPathRowsDetectCycle(t *testing.T) {
	opt := options{separator: " > "}
	nodes := []categoryNode{
		{CategoryID: "1", ParentID: "2", IsLeaf: 0, NameCn: "A", ShopType: shopTypeHalf, Version: "v2", SellerType: "LOCAL"},
		{CategoryID: "2", ParentID: "1", IsLeaf: 1, NameCn: "B", ShopType: shopTypeHalf, Version: "v2", SellerType: "LOCAL"},
	}

	rows, stat := buildPathRows(nodes, opt)
	if len(rows) != 0 {
		t.Fatalf("expected no rows for cycle, got %d", len(rows))
	}
	if stat.cycle != 1 {
		t.Fatalf("expected cycle=1, got %+v", stat)
	}
}
