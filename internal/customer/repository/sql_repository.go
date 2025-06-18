package repository

import (
	"context"

	"github.com/nurdsoft/nurd-commerce-core/internal/customer/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/customer/errors"
	dbErrors "github.com/nurdsoft/nurd-commerce-core/shared/db"
	"gorm.io/gorm"
)

type sqlRepository struct {
	gormDB *gorm.DB
}

func (r *sqlRepository) Create(ctx context.Context, customer *entities.Customer) (*entities.Customer, error) {
	err := r.gormDB.Create(customer).Error
	if err != nil {
		if dbErrors.IsAlreadyExistError(err) {
			existingCustomer, findErr := r.FindByEmail(ctx, customer.Email)
			if findErr != nil {
				return nil, findErr
			}
			return existingCustomer, err
		}
		return nil, err
	}

	return customer, nil
}

func (r *sqlRepository) FindByEmail(_ context.Context, email string) (*entities.Customer, error) {
	customer := &entities.Customer{}
	err := r.gormDB.Where("email = ?", email).First(customer).Error
	if err != nil {
		return nil, err
	}

	return customer, nil
}

func (r *sqlRepository) FindByID(_ context.Context, ID string) (*entities.Customer, error) {
	customer := &entities.Customer{}
	err := r.gormDB.Where("id = ?", ID).First(customer).Error
	if err != nil {
		return nil, err
	}

	return customer, nil
}

func (r *sqlRepository) Update(ctx context.Context, details map[string]interface{}, ID string) error {
	result := r.gormDB.WithContext(ctx).Model(&entities.Customer{}).Where("id = ?", ID).Updates(details)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return moduleErrors.NewAPIError("CUSTOMER_NOT_FOUND")
	}

	return nil
}
