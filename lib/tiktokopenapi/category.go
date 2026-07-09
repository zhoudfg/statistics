package tiktokopenapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type CategoryService struct {
	client *Client
}

func NewCategoryService(client *Client) *CategoryService {
	return &CategoryService{client: client}
}

func NewCategoryServiceFromConf(shop ShopOptions) *CategoryService {
	return NewClientFromConf(shop).NewCategoryService()
}

func (s *CategoryService) GetCategoryList(ctx context.Context) (json.RawMessage, error) {
	return s.GetCategoryListWithPageToken(ctx, "")
}

func (s *CategoryService) GetCategoryListWithPageToken(ctx context.Context, pageToken string) (json.RawMessage, error) {
	queries := s.client.BaseQueries("en-US")
	path, err := s.productCategoryPath(queries)
	if err != nil {
		return nil, err
	}
	s.addLegacyCategoryVersion(queries)
	if pageToken != "" {
		queries["page_token"] = pageToken
	}

	return s.client.Request(ctx, http.MethodGet, path, queries, nil)
}

func (s *CategoryService) GetCategoryAttributeList(ctx context.Context, categoryID string) (json.RawMessage, error) {
	queries := s.client.BaseQueries("zh-CN")
	path, err := s.productAttributePath(categoryID, queries)
	if err != nil {
		return nil, err
	}
	s.addLegacyCategoryVersion(queries)

	return s.client.Request(ctx, http.MethodGet, path, queries, nil)
}

func (s *CategoryService) GetFullCategoryAttributeList(ctx context.Context, categoryID string) (json.RawMessage, error) {
	queries := s.client.BaseQueries("zh-CN")
	path := fmt.Sprintf("/gs_full_service_commodity/%s/beta/categories/%s/attributes", s.client.FullCategoryAttrVersion, categoryID)

	return s.client.Request(ctx, http.MethodGet, path, queries, nil)
}

func (s *CategoryService) GetFullCategoryList(ctx context.Context, params interface{}) (json.RawMessage, error) {
	queries := s.client.BaseQueries("zh-CN")
	path := fmt.Sprintf("/gs_full_service_commodity/%s/beta/categories/search", s.client.FullVersion)

	return s.client.Request(ctx, http.MethodPost, path, queries, params)
}

func (s *CategoryService) GetImageUploadRequirements(ctx context.Context, categoryID string) (json.RawMessage, error) {
	queries := s.client.BaseQueries("zh-CN")
	path := fmt.Sprintf("/gs_full_service_commodity/%s/beta/categories/%s/image_upload_requirements", s.client.FullVersion, categoryID)

	return s.client.Request(ctx, http.MethodGet, path, queries, nil)
}

func (s *CategoryService) GetCertificationsList(ctx context.Context, categoryID string) (json.RawMessage, error) {
	queries := s.client.BaseQueries("zh-CN")
	path, err := s.productRulesPath(categoryID, queries)
	if err != nil {
		return nil, err
	}
	s.addLegacyCategoryVersion(queries)

	data, err := s.client.Request(ctx, http.MethodGet, path, queries, nil)
	if err != nil {
		return nil, err
	}
	return rawFieldOrEmptyArray(data, "product_certifications")
}

func (s *CategoryService) GetFullCertificationsList(ctx context.Context, categoryID string) (json.RawMessage, error) {
	queries := s.client.BaseQueries("zh-CN")
	path := fmt.Sprintf("/gs_full_service_commodity/%s/beta/categories/%s/certifications", s.client.FullCategoryAttrVersion, categoryID)

	data, err := s.client.Request(ctx, http.MethodGet, path, queries, nil)
	if err != nil {
		return nil, err
	}
	return rawFieldOrEmptyArray(data, "certifications")
}

func (s *CategoryService) GetCategoryRecommend(ctx context.Context, params map[string]interface{}) (json.RawMessage, error) {
	queries := s.client.BaseQueries("zh-CN")
	path, err := s.productRecommendPath(queries)
	if err != nil {
		return nil, err
	}
	if s.client.Region == RegionUS {
		params["category_version"] = CategoryVersionV2
	} else if s.client.shouldUseCategoryV2() {
		queries["category_version"] = CategoryVersionV2
	}

	return s.client.Request(ctx, http.MethodPost, path, queries, params)
}

func (s *CategoryService) GetHalfCategoryList(ctx context.Context) (json.RawMessage, error) {
	return s.GetHalfCategoryListWithLocale(ctx, "zh-CN")
}

func (s *CategoryService) GetHalfCategoryListWithLocale(ctx context.Context, locale string) (json.RawMessage, error) {
	if locale == "" {
		locale = "zh-CN"
	}
	queries := s.client.BaseQueries(locale)
	path, err := s.productCategoryPath(queries)
	if err != nil {
		return nil, err
	}
	queries["category_version"] = s.client.CategoryVersion

	return s.client.Request(ctx, http.MethodGet, path, queries, nil)
}

func (s *CategoryService) GetHalfCategoryAttributeList(ctx context.Context, categoryID string) (json.RawMessage, error) {
	queries := s.client.BaseQueries("zh-CN")
	path, err := s.productAttributePath(categoryID, queries)
	if err != nil {
		return nil, err
	}
	queries["category_version"] = s.client.CategoryVersion

	return s.client.Request(ctx, http.MethodGet, path, queries, nil)
}

func (s *CategoryService) GetHalfCertificationsList(ctx context.Context, categoryID string) (json.RawMessage, error) {
	queries := s.client.BaseQueries("zh-CN")
	path := fmt.Sprintf("/product/%s/categories/%s/rules", s.client.Version, categoryID)
	queries["shop_cipher"] = s.client.ShopCipher
	queries["category_version"] = s.client.CategoryVersion

	return s.client.Request(ctx, http.MethodGet, path, queries, nil)
}

func (s *CategoryService) GetHalfCategoryRecommend(ctx context.Context, params interface{}) (json.RawMessage, error) {
	queries := s.client.BaseQueries("zh-CN")
	path := fmt.Sprintf("/product/%s/categories/recommend", s.client.Version)
	queries["shop_cipher"] = s.client.ShopCipher
	queries["category_version"] = s.client.CategoryVersion

	return s.client.Request(ctx, http.MethodPost, path, queries, params)
}

func (s *CategoryService) productCategoryPath(queries map[string]string) (string, error) {
	switch s.client.SellerType {
	case SellerTypeCrossBorder:
		return fmt.Sprintf("/product/%s/global_categories", s.client.Version), nil
	case SellerTypeLocal:
		queries["shop_cipher"] = s.client.ShopCipher
		return fmt.Sprintf("/product/%s/categories", s.client.Version), nil
	default:
		return "", fmt.Errorf("unsupported tiktok seller type: %s", s.client.SellerType)
	}
}

func (s *CategoryService) productAttributePath(categoryID string, queries map[string]string) (string, error) {
	switch s.client.SellerType {
	case SellerTypeCrossBorder:
		return fmt.Sprintf("/product/%s/categories/%s/global_attributes", s.client.Version, categoryID), nil
	case SellerTypeLocal:
		queries["shop_cipher"] = s.client.ShopCipher
		return fmt.Sprintf("/product/%s/categories/%s/attributes", s.client.Version, categoryID), nil
	default:
		return "", fmt.Errorf("unsupported tiktok seller type: %s", s.client.SellerType)
	}
}

func (s *CategoryService) productRulesPath(categoryID string, queries map[string]string) (string, error) {
	switch s.client.SellerType {
	case SellerTypeCrossBorder:
		return fmt.Sprintf("/product/%s/categories/%s/global_rules", s.client.Version, categoryID), nil
	case SellerTypeLocal:
		queries["shop_cipher"] = s.client.ShopCipher
		return fmt.Sprintf("/product/%s/categories/%s/rules", s.client.Version, categoryID), nil
	default:
		return "", fmt.Errorf("unsupported tiktok seller type: %s", s.client.SellerType)
	}
}

func (s *CategoryService) productRecommendPath(queries map[string]string) (string, error) {
	switch s.client.SellerType {
	case SellerTypeCrossBorder:
		return fmt.Sprintf("/product/%s/global_categories/recommend", s.client.Version), nil
	case SellerTypeLocal:
		queries["shop_cipher"] = s.client.ShopCipher
		return fmt.Sprintf("/product/%s/categories/recommend", s.client.Version), nil
	default:
		return "", fmt.Errorf("unsupported tiktok seller type: %s", s.client.SellerType)
	}
}

func (s *CategoryService) addLegacyCategoryVersion(queries map[string]string) {
	if s.client.shouldUseCategoryV2() {
		queries["category_version"] = CategoryVersionV2
	}
}

func rawFieldOrEmptyArray(data json.RawMessage, field string) (json.RawMessage, error) {
	if len(data) == 0 {
		return json.RawMessage("[]"), nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	value, ok := obj[field]
	if !ok || len(value) == 0 || string(value) == "null" {
		return json.RawMessage("[]"), nil
	}
	return value, nil
}
