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
