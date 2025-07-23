package entities

type GetSyncProductsRequest struct {
	Search string 
	Limit int 
	Offset int 
}

type SyncProductsResponse struct {
	Code   int                `json:"code"`
	Paging SyncProductsPaging `json:"paging,omitzero"`
	Result []SyncProduct      `json:"result"`
}

type SyncProductsPaging struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type SyncProduct struct {
	ID           int    `json:"id"`
	ExternalID   string `json:"external_id"`
	Name         string `json:"name"`
	Variants     int    `json:"variants"`
	Synced       int    `json:"synced"`
	ThumbnailURL string `json:"thumbnail_url"`
	IsIgnored    bool   `json:"is_ignored"`
}

// GetSyncProductResponse represents the response from Printful get sync product API
type GetSyncProductResponse struct {
	Code   int                    `json:"code"`
	Result GetSyncProductResult   `json:"result"`
}

type GetSyncProductResult struct {
	SyncProduct   SyncProduct `json:"sync_product"`
	SyncVariants  []SyncVariant     `json:"sync_variants"`
}

type SyncVariant struct {
	ID                      int                `json:"id"`
	ExternalID              string             `json:"external_id"`
	SyncProductID           int                `json:"sync_product_id"`
	Name                    string             `json:"name"`
	Synced                  bool               `json:"synced"`
	VariantID               int                `json:"variant_id"`
	RetailPrice             string             `json:"retail_price"`
	Currency                string             `json:"currency"`
	IsIgnored               bool               `json:"is_ignored"`
	SKU                     string             `json:"sku"`
	Product                 SyncVariantProduct `json:"product"`
	Files                   []SyncVariantFile  `json:"files"`
	MainCategoryID          int                `json:"main_category_id"`
	WarehouseProductID      int                `json:"warehouse_product_id"`
	WarehouseProductVariantID int              `json:"warehouse_product_variant_id"`
	Size                    string             `json:"size"`
	Color                   string             `json:"color"`
	AvailabilityStatus      string             `json:"availability_status"`
}

type SyncVariantProduct struct {
	VariantID int    `json:"variant_id"`
	ProductID int    `json:"product_id"`
	Image     string `json:"image"`
	Name      string `json:"name"`
}

type SyncVariantFile struct {
	Type              string                `json:"type"`
	ThumbnailURL      string                `json:"thumbnail_url"`
	PreviewURL        string                `json:"preview_url"`
}
