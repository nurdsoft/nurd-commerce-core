package repository

import (
	"context"
	"github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/address/errors"
	dbErrors "github.com/nurdsoft/nurd-commerce-core/shared/db"
	"gorm.io/gorm"
)

type sqlRepository struct {
	gormDB *gorm.DB
}

func (r *sqlRepository) CreateAddress(ctx context.Context, address *entities.Address) (*entities.Address, error) {
	err := r.gormDB.Transaction(func(tx *gorm.DB) error {
		if address.IsDefault {
			if err := tx.WithContext(ctx).Model(&entities.Address{}).Where("customer_id = ?", address.CustomerID).Update("is_default", false).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(address).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return address, nil
}

func (r *sqlRepository) GetAddresses(ctx context.Context, customerID string) ([]entities.Address, error) {
	var addresses []entities.Address
	err := r.gormDB.WithContext(ctx).Where("customer_id = ?", customerID).Find(&addresses).Error
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

func (r *sqlRepository) GetAddress(ctx context.Context, customerID, addressID string) (*entities.Address, error) {
	address := &entities.Address{}
	err := r.gormDB.WithContext(ctx).Where("customer_id = ? AND id = ?", customerID, addressID).First(address).Error
	if err != nil {
		if dbErrors.IsNotFoundError(err) {
			return nil, moduleErrors.NewAPIError("ADDRESS_NOT_FOUND")
		}
		return nil, err
	}

	return address, nil
}

func (r *sqlRepository) UpdateAddress(ctx context.Context, address *entities.Address) (*entities.Address, error) {
	err := r.gormDB.Transaction(func(tx *gorm.DB) error {
		if address.IsDefault {
			if err := tx.WithContext(ctx).Model(&entities.Address{}).Where("customer_id = ?", address.CustomerID).Update("is_default", false).Error; err != nil {
				return err
			}
		}

		if err := tx.WithContext(ctx).Model(&entities.Address{}).Where("customer_id = ? AND id = ?", address.CustomerID, address.ID).Updates(address).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	updatedAddress := &entities.Address{}
	err = r.gormDB.WithContext(ctx).Where("customer_id = ? AND id = ?", address.CustomerID, address.ID).First(updatedAddress).Error
	if err != nil {
		return nil, err
	}

	return updatedAddress, nil
}

func (r *sqlRepository) DeleteAddress(ctx context.Context, customerID, addressID string) error {
	err := r.gormDB.WithContext(ctx).Where("customer_id = ? AND id = ?", customerID, addressID).Delete(&entities.Address{}).Error
	if err != nil {
		if dbErrors.IsNotFoundError(err) {
			return moduleErrors.NewAPIError("ADDRESS_NOT_FOUND")
		}
		return err
	}

	return nil
}

func (r *sqlRepository) UpdateAddressField(ctx context.Context, customerID, addressID string, details map[string]interface{}) error {
	err := r.gormDB.WithContext(ctx).Model(&entities.Address{}).Where("customer_id = ? AND id = ?", customerID, addressID).Updates(details).Error
	if err != nil {
		return err
	}

	return nil
}
