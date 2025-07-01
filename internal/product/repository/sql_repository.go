package repository

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	dbErrors "github.com/nurdsoft/nurd-commerce-core/shared/db"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/shared/errors"
	"gorm.io/gorm"
)

type sqlRepository struct {
	gormDB *gorm.DB
}

func (r *sqlRepository) Create(ctx context.Context, product *entities.Product) (*entities.Product, error) {
	err := r.gormDB.Create(product).Error
	if err != nil {
		if dbErrors.IsAlreadyExistError(err) {
			existingProduct, findErr := r.FindByID(ctx, product.ID.String())
			if findErr != nil {
				return nil, findErr
			}
			return existingProduct, err
		}
		return nil, err
	}

	return product, nil
}

func (r *sqlRepository) FindByID(_ context.Context, ID string) (*entities.Product, error) {
	product := &entities.Product{}
	err := r.gormDB.Where("id = ?", ID).First(product).Error
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *sqlRepository) FindByIDs(ctx context.Context, ids []string) ([]entities.Product, error) {
	var products []entities.Product
	err := r.gormDB.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *sqlRepository) Update(ctx context.Context, details map[string]interface{}, ID string) error {
	result := r.gormDB.WithContext(ctx).Model(&entities.Product{}).Where("id = ?", ID).Updates(details)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return moduleErrors.NewAPIError("PRODUCT_NOT_FOUND")
	}

	return nil
}

func (r *sqlRepository) CreateVariant(ctx context.Context, variant *entities.ProductVariant) (*entities.ProductVariant, error) {
	err := r.gormDB.Create(variant).Error
	if err != nil {
		if dbErrors.IsAlreadyExistError(err) {
			existingVariant, findErr := r.FindVariantBySKU(ctx, variant.SKU)
			if findErr != nil {
				return nil, findErr
			}
			return existingVariant, err
		}
		return nil, err
	}

	return variant, nil
}

func (r *sqlRepository) FindVariantBySKU(_ context.Context, sku string) (*entities.ProductVariant, error) {
	variant := &entities.ProductVariant{}
	err := r.gormDB.Where("sku = ?", sku).First(variant).Error
	if err != nil {
		return nil, err
	}

	return variant, nil
}

func (r *sqlRepository) UpdateVariant(ctx context.Context, details map[string]interface{}, ID string) error {
	result := r.gormDB.WithContext(ctx).Model(&entities.ProductVariant{}).Where("id = ?", ID).Updates(details)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return moduleErrors.NewAPIError("PRODUCT_VARIANT_NOT_FOUND")
	}

	return nil
}

func (r *sqlRepository) FindVariantByID(_ context.Context, id string) (*entities.ProductVariant, error) {
	variant := &entities.ProductVariant{}
	err := r.gormDB.Where("id = ?", id).First(variant).Error
	if err != nil {
		return nil, err
	}

	return variant, nil
}

func (r *sqlRepository) ListVariants(ctx context.Context, req *entities.ListProductVariantsRequest) (*entities.ListProductVariantsResponse, error) {
	var variants []entities.ProductVariant
	var total int64

	query := r.gormDB.WithContext(ctx).Model(&entities.ProductVariant{})

	// Apply search and filters
	if req.Search != nil && *req.Search != "" {
		searchTerm := "%" + *req.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}

	if req.MinPrice != nil {
		query = query.Where("price >= ?", req.MinPrice)
	}
	if req.MaxPrice != nil {
		query = query.Where("price <= ?", req.MaxPrice)
	}

	if len(req.Attributes) > 0 {
		for key, value := range req.Attributes {
			query = query.Where("attributes->>? = ?", key, value)
		}
	}

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	// Apply sorting
	sortBy := "created_at"
	if req.SortBy != nil && *req.SortBy != "" {
		sortBy = *req.SortBy
	}
	sortOrder := "desc"
	if req.SortOrder != nil && *req.SortOrder != "" {
		sortOrder = *req.SortOrder
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	err = query.Order(sortBy + " " + sortOrder).
		Offset(offset).
		Limit(pageSize).
		Find(&variants).Error
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &entities.ListProductVariantsResponse{
		Data: variants,
		Pagination: entities.PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}, nil
}
