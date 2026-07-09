package main

import (
	"database/sql"
	"testing"
)

func TestBuildHalfExportRows(t *testing.T) {
	row := categoryAttrRow{
		AttrID:     1,
		CategoryID: "601",
		RawJSON: []byte(`{
			"attributes": [
				{
					"id": "100107",
					"name": "\u4fdd\u4fee\u7c7b\u578b",
					"type": "PRODUCT_PROPERTY",
					"values": [
						{"id": "1000057", "name": "\u65e0\u4fdd\u4fee"},
						{"id": "1000054", "name": "\u4f9b\u5e94\u5546\u4fdd\u4fee"}
					],
					"is_requried": false,
					"is_multiple_selection": true
				}
			]
		}`),
		ShopType:   sql.NullInt64{Int64: shopTypeHalf, Valid: true},
		Version:    sql.NullString{String: "v2", Valid: true},
		SellerType: sql.NullString{String: "LOCAL", Valid: true},
		PathCn:     "\u624b\u673a",
		PathID:     "601",
		PathEn:     "Phones",
	}

	rows, err := buildExportRows(row)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 export row, got %d", len(rows))
	}
	got := rows[0]
	if got.Attr.String != "\u4fdd\u4fee\u7c7b\u578b" {
		t.Fatalf("unexpected attr: %q", got.Attr.String)
	}
	if got.AttrID.String != "100107" {
		t.Fatalf("unexpected attr_id: %q", got.AttrID.String)
	}
	if got.AttrType.String != "PRODUCT_PROPERTY" {
		t.Fatalf("unexpected attr_type: %q", got.AttrType.String)
	}
	wantValues := "\u3010\u201c\u65e0\u4fdd\u4fee\u201d\uff0c\u201c\u4f9b\u5e94\u5546\u4fdd\u4fee\u201d\u3011"
	if got.AttrValues.String != wantValues {
		t.Fatalf("unexpected attr_values: %q", got.AttrValues.String)
	}
	if got.IsRequired.Int64 != 0 || !got.IsRequired.Valid {
		t.Fatalf("expected required=false, got %+v", got.IsRequired)
	}
	if got.IsMultipleSelection.Int64 != 1 || !got.IsMultipleSelection.Valid {
		t.Fatalf("expected multiple=true, got %+v", got.IsMultipleSelection)
	}
}

func TestBuildFullExportRows(t *testing.T) {
	row := categoryAttrRow{
		AttrID:     2,
		CategoryID: "700",
		RawJSON: []byte(`{
			"attributes": [
				{
					"id": "7509392418581219079",
					"type": "BASE",
					"values": [
						{"id": "7338649511073875730", "name_en": "No", "name_zh": "\u5426"},
						{"id": "7452605711303132929", "name_en": "Yes", "name_zh": "\u662f"}
					],
					"name_en": "Product Contain Wooden Materials",
					"name_zh": "\u5546\u54c1\u542b\u6728\u5236\u6750\u6599",
					"value_type": "ENUM",
					"is_required": true,
					"is_mutiple_selection": false
				}
			]
		}`),
		ShopType:   sql.NullInt64{Int64: shopTypeFull, Valid: true},
		Version:    sql.NullString{String: "v1", Valid: true},
		SellerType: sql.NullString{String: "ALL", Valid: true},
		PathCn:     "\u5bb6\u5c45",
		PathID:     "700",
		PathEn:     "Home",
	}

	rows, err := buildExportRows(row)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 export row, got %d", len(rows))
	}
	got := rows[0]
	if got.Attr.String != "\u5546\u54c1\u542b\u6728\u5236\u6750\u6599" {
		t.Fatalf("unexpected attr: %q", got.Attr.String)
	}
	if got.AttrType.String != "ENUM" {
		t.Fatalf("unexpected attr_type: %q", got.AttrType.String)
	}
	wantValues := "\u3010\u201c\u5426\u201d\uff0c\u201c\u662f\u201d\u3011"
	if got.AttrValues.String != wantValues {
		t.Fatalf("unexpected attr_values: %q", got.AttrValues.String)
	}
	if got.IsRequired.Int64 != 1 || !got.IsRequired.Valid {
		t.Fatalf("expected required=true, got %+v", got.IsRequired)
	}
	if got.IsMultipleSelection.Int64 != 0 || !got.IsMultipleSelection.Valid {
		t.Fatalf("expected multiple=false, got %+v", got.IsMultipleSelection)
	}
}
