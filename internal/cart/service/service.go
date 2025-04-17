package service

import (
	"context"
	"encoding/json"
	"fmt"
	sharedDecimal "github.com/nurdsoft/nurd-commerce-core/shared/decimal"
	salesforce "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/client"
	salesforceEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/inventory/salesforce/entities"
	shipengineEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/shipengine/entities"
	stripeEntities "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/entities"
	"time"

	"github.com/nurdsoft/nurd-commerce-core/internal/address/addressclient"
	addressEntities "github.com/nurdsoft/nurd-commerce-core/internal/address/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/entities"
	moduleErrors "github.com/nurdsoft/nurd-commerce-core/internal/cart/errors"
	"github.com/nurdsoft/nurd-commerce-core/internal/cart/repository"
	productEntities "github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/productclient"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/shared/cache"
	dbErrors "github.com/nurdsoft/nurd-commerce-core/shared/db"
	sharedMeta "github.com/nurdsoft/nurd-commerce-core/shared/meta"
	shipengine "github.com/nurdsoft/nurd-commerce-core/shared/vendors/shipping/shipengine/client"
	stripe "github.com/nurdsoft/nurd-commerce-core/shared/vendors/taxes/stripe/client"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type Service interface {
	UpdateCartItem(ctx context.Context, req *entities.UpdateCartItemRequest) (*entities.CartItem, error)
	GetCartItems(ctx context.Context) (*entities.GetCartItemsResponse, error)
	RemoveCartItem(ctx context.Context, req string) error
	ClearCartItems(ctx context.Context) error
	GetTaxRate(ctx context.Context, req *entities.GetTaxRateRequest) (*entities.GetTaxRateResponse, error)
	GetShippingRate(ctx context.Context, req *entities.GetShippingRateRequest) (*entities.GetShippingRateResponse, error)
	GetShippingRateByID(ctx context.Context, shippingRateID uuid.UUID) (*entities.CartShippingRate, error)
	GetCart(ctx context.Context) (*entities.Cart, error)
}

type service struct {
	repo             repository.Repository
	log              *zap.SugaredLogger
	shipengineClient shipengine.Client
	stripeClient     stripe.Client
	cache            cache.Cache
	productClient    productclient.Client
	addressClient    addressclient.Client
	salesforceClient salesforce.Client
}

func New(
	repo repository.Repository,
	log *zap.SugaredLogger,
	shipengineClient shipengine.Client,
	stripeClient stripe.Client,
	cache cache.Cache,
	productClient productclient.Client,
	addressClient addressclient.Client,
	salesforceClient salesforce.Client,
) Service {
	return &service{
		repo:             repo,
		log:              log,
		shipengineClient: shipengineClient,
		stripeClient:     stripeClient,
		cache:            cache,
		productClient:    productClient,
		addressClient:    addressClient,
		salesforceClient: salesforceClient,
	}
}

// swagger:route POST /cart/items carts UpdateCartItemRequest
//
// # Add Cart Item
// ### Add an item to the cart
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: CartItem Item added to cart successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) UpdateCartItem(ctx context.Context, req *entities.UpdateCartItemRequest) (*entities.CartItem, error) {
	customerID := sharedMeta.XCustomerID(ctx)

	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	// Start a transaction
	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		s.log.Errorf("Error starting transaction: %v", err)
		return nil, moduleErrors.NewAPIError("CART_ERROR_UPDATING_CART_ITEM")
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw after rollback
		} else if err != nil {
			tx.Rollback() // rollback on error
		} else if commitErr := tx.Commit().Error; commitErr != nil {
			s.log.Errorf("Error committing transaction: %v", commitErr)
			err = moduleErrors.NewAPIError("CART_ERROR_UPDATING_CART_ITEM")
		}
	}()

	// Step 1: Retrieve the active cart for the customer
	cart, err := s.repo.GetActiveCart(ctx, customerID)
	if err != nil {
		s.log.Errorf("Error retrieving active cart: %v", err)
		return nil, moduleErrors.NewAPIError("CART_ERROR_GETTING_CART")
	}

	// Step 2: If cart does not exist, create a new one
	if cart == nil {
		// No active cart found, create a new cart
		cart, err = s.repo.CreateNewCart(ctx, tx, customerID)
		if err != nil {
			s.log.Errorf("Error creating new cart: %v", err)
			return nil, moduleErrors.NewAPIError("CART_ERROR_UPDATING_CART_ITEM")
		}
	}

	product, err := s.productClient.GetProduct(ctx, &productEntities.GetProductRequest{
		ProductID: req.Item.ProductID,
	})

	if (err != nil || product == nil) && req.Item.ProductData != nil {
		productData := *req.Item.ProductData
		product, err = s.productClient.CreateProduct(ctx, &productEntities.CreateProductRequest{
			Data: &productEntities.CreateProductRequestBody{
				ID:          &req.Item.ProductID,
				Name:        productData.Name,
				Description: productData.Description,
				ImageURL:    productData.ImageURL,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	if product == nil {
		return nil, moduleErrors.NewAPIError("PRODUCT_NOT_FOUND")
	}

	productVariant, err := s.productClient.GetProductVariant(ctx, &productEntities.GetProductVariantRequest{
		SKU: req.Item.SKU,
	})

	if req.Item.ProductData != nil {
		productData := *req.Item.ProductData
		productVariant, err = s.productClient.CreateProductVariant(ctx, &productEntities.CreateProductVariantRequest{
			ProductID: product.ID,
			Data: &productEntities.CreateProductVariantRequestBody{
				SKU:           req.Item.SKU,
				Name:          productData.Name,
				Description:   productData.Description,
				ImageURL:      productData.ImageURL,
				Price:         productData.Price,
				Currency:      productData.Currency,
				Length:        productData.Length,
				Width:         productData.Width,
				Height:        productData.Height,
				Weight:        productData.Weight,
				Attributes:    productData.Attributes,
				StripeTaxCode: productData.StripeTaxCode,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	if productVariant == nil {
		return nil, moduleErrors.NewAPIError("PRODUCT_VARIANT_NOT_FOUND")
	}

	if err != nil {
		s.log.Errorf("Error fetching product variant: %v", err)
		return nil, moduleErrors.NewAPIError("CART_ITEM_NOT_FOUND")
	}

	// Step 3: Check if the item is already in the cart
	item, err := s.repo.GetCartItem(ctx, cart.Id.String(), productVariant.ID.String())
	if err != nil {
		s.log.Errorf("Error checking for existing cart item: %v", err)
		return nil, moduleErrors.NewAPIError("CART_ERROR_UPDATING_CART_ITEM")
	}

	// Step 4: Prepare changes to be saved
	var resultItem *entities.CartItem
	if item != nil {
		if req.Item.Quantity == 0 {
			// Remove the item from the cart
			err = s.repo.RemoveCartItem(ctx, cart.Id.String(), item.ID.String())
			if err != nil {
				if dbErrors.IsNotFoundError(err) {
					return nil, moduleErrors.NewAPIError("CART_ITEM_NOT_FOUND")
				}
				s.log.Errorf("Error removing cart item: %v", err)
				return item, moduleErrors.NewAPIError("CART_ERROR_UPDATING_CART_ITEM")
			}
		} else {
			// Item already exists, replace quantity and price
			item.Quantity = req.Item.Quantity
			if err = s.repo.UpdateCartItem(ctx, tx, item.ID.String(), item.Quantity); err != nil {
				s.log.Errorf("Error updating cart item quantity: %v", err)
				return nil, moduleErrors.NewAPIError("CART_ERROR_UPDATING_CART_ITEM")
			}
			resultItem = item
		}
	} else if req.Item.Quantity > 0 {
		// Item does not exist, prepare to add
		resultItem, err = s.repo.AddCartItem(ctx, tx, cart.Id.String(), productVariant.ID.String(), req.Item.Quantity)
		if err != nil {
			s.log.Errorf("Error adding item to cart: %v", err)
			return nil, moduleErrors.NewAPIError("CART_ERROR_UPDATING_CART_ITEM")
		} else {
			// only sync with salesforce if the item is successfully added to the cart
			go func() {
				bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
				defer cancel()

				// If product already has a SalesforceID no need to create it again
				if product.SalesforceID != nil {
					return
				}

				var pricebookEntry *salesforceEntities.CreateSFPriceBookEntryResponse
				// Create a salesforce product
				salesforceProduct, err := s.salesforceClient.CreateProduct(bgCtx, &salesforceEntities.CreateSFProductRequest{
					Name: product.Name,
					// TODO: confirm if this is the correct description
					Description: productVariant.SKU,
					ProductCode: product.ID.String(),
					IsActive:    true,
				})
				if err != nil {
					s.log.Errorf("Error creating product on salesforce: %v", err)
				}

				if product != nil {
					// Create a salesforce pricebook entry
					pricebookEntry, err = s.salesforceClient.CreatePriceBookEntry(bgCtx, &salesforceEntities.CreateSFPriceBookEntryRequest{
						Product2ID:   salesforceProduct.ID,
						Pricebook2ID: salesforceEntities.StandardPriceBook,
						// Unit price is 0 because the real price will be available in the order items
						UnitPrice: 0,
					})
					if err != nil {
						s.log.Errorf("Error creating pricebook entry on salesforce: %v", err)
					}
				}

				if pricebookEntry != nil {
					// Save the salesforce product and pricebook entry ids to the database
					err = s.productClient.UpdateProduct(bgCtx, &productEntities.UpdateProductRequest{
						ProductID: product.ID,
						Data: &productEntities.UpdateProductRequestBody{
							SalesforceID:               salesforceProduct.ID,
							SalesforcePricebookEntryId: pricebookEntry.ID,
						},
					})
					if err != nil {
						s.log.Errorf("Error updating product: %v", err)
					}
				}

			}()
		}
	}

	// evict the cache for the shipping rates
	go func() {
		// delete by pattern shipping_rate_<any_address_id>_<customer_id>_<cart_id>
		if err := s.cache.DeleteByPattern(context.Background(), fmt.Sprintf("^shipping_rate_[^_]+_%s_%s$", customerID, cart.Id.String())); err != nil {
			s.log.Errorf("Error deleting shipping rate cache: %v", err)
		}
	}()

	return resultItem, nil

}

// swagger:route GET /cart/items carts GetCartItems
//
// # Get Cart Items
// ### Get the list of items in the cart
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetCartItemsResponse Cart items retrieved successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) GetCartItems(ctx context.Context) (*entities.GetCartItemsResponse, error) {
	customerID := sharedMeta.XCustomerID(ctx)
	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	// Retrieve active cart
	cart, err := s.repo.GetActiveCart(ctx, customerID)
	if err != nil || cart == nil {
		s.log.Errorf("Error retrieving active cart: %v", err)
		return nil, nil
	}

	// Fetch items in the active cart
	items, err := s.repo.GetCartItems(ctx, cart.Id.String())
	if err != nil {
		s.log.Errorf("Error retrieving cart items: %v", err)
		return nil, moduleErrors.NewAPIError("CART_ERROR_GETTING_CART_ITEMS")
	}

	return &entities.GetCartItemsResponse{Items: items}, nil
}

// swagger:route DELETE /cart/items/{item_id} carts RemoveCartItem
//
// # Remove Cart Item
// ### Remove an item from the cart
//
// Parameters:
//
//   + name: item_id
//     in: path
//     description: ID of the item to be removed
//     required: true
//     type: string
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: DefaultResponse Item removed successfully
//	400: DefaultError Bad Request
//	404: DefaultError Not Found
//	500: DefaultError Internal Server Error
func (s *service) RemoveCartItem(ctx context.Context, itemID string) error {
	customerID := sharedMeta.XCustomerID(ctx)
	if customerID == "" {
		return moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	// Retrieve active cart
	cart, err := s.repo.GetActiveCart(ctx, customerID)
	if err != nil {
		s.log.Errorf("Error retrieving active cart: %v", err)
		return moduleErrors.NewAPIError("CART_ERROR_GETTING_CART")
	}
	if cart == nil {
		return moduleErrors.NewAPIError("CART_NOT_FOUND")
	}

	// Remove the item from the cart
	err = s.repo.RemoveCartItem(ctx, cart.Id.String(), itemID)
	if err != nil {
		if dbErrors.IsNotFoundError(err) {
			return moduleErrors.NewAPIError("CART_ITEM_NOT_FOUND")
		}
		s.log.Errorf("Error removing cart item: %v", err)
		return moduleErrors.NewAPIError("CART_ERROR_REMOVING_CART_ITEM")
	} else { // evict the cache for the shipping rates
		go func() {
			// delete by pattern shipping_rate_<any_address_id>_<customer_id>_<cart_id>
			if err := s.cache.DeleteByPattern(context.Background(), fmt.Sprintf("^shipping_rate_[^_]+_%s_%s$", customerID, cart.Id.String())); err != nil {
				s.log.Errorf("Error deleting shipping rate cache: %v", err)
			}
		}()
	}

	return nil
}

// swagger:route DELETE /cart/items carts ClearCartItems
//
// # Clear Cart Items
// ### Clear the list of items in the cart
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: DefaultResponse Cart items cleared successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) ClearCartItems(ctx context.Context) error {
	customerID := sharedMeta.XCustomerID(ctx)
	if customerID == "" {
		return moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	// Retrieve active cart
	cart, err := s.repo.GetActiveCart(ctx, customerID)
	if err != nil {
		s.log.Errorf("Error retrieving active cart: %v", err)
		return moduleErrors.NewAPIError("CART_ERROR_GETTING_CART")
	}
	if cart == nil {
		return moduleErrors.NewAPIError("CART_NOT_FOUND")
	}

	// Mark the cart as cleared
	err = s.repo.UpdateCartStatus(ctx, nil, cart.Id.String(), string(entities.Cleared))
	if err != nil {
		s.log.Errorf("Error updating cart status to cleared: %v", err)
		return moduleErrors.NewAPIError("CART_ERROR_CLEARING_CART")
	}
	return nil
}

// swagger:route POST /cart/tax-rate carts GetTaxRateRequest
//
// # Get Tax Rate for active cart
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetTaxRateResponse Tax rate retrieved successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) GetTaxRate(ctx context.Context, req *entities.GetTaxRateRequest) (*entities.GetTaxRateResponse, error) {
	customerID := sharedMeta.XCustomerID(ctx)
	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	address, err := s.addressClient.GetAddress(ctx, &addressEntities.GetAddressRequest{
		AddressID: req.Body.AddressID,
	})
	if err != nil {
		s.log.Errorf("Error retrieving address: %v", err)
		return nil, err
	}

	// get items from active cart
	getActiveCarItems, err := s.GetCartItems(ctx)
	if err != nil {
		return nil, moduleErrors.NewAPIError("CART_ERROR_GETTING_CART_ITEMS")
	}

	// get customer selected shipping rate
	shippingRate, err := s.repo.GetShippingRate(ctx, req.Body.ShippingRateID)
	if err != nil {
		s.log.Errorf("Error retrieving shipping rate: %v", err)
		return nil, err
	}

	cacheKey := getTaxRateCacheKey(req.Body.AddressID.String(), customerID, shippingRate.Id.String())

	cachedResponse, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedResponse != nil {
		var response entities.GetTaxRateResponse
		if err := json.Unmarshal(cachedResponse.([]byte), &response); err == nil {
			return &response, nil
		}
	}

	var taxItems []stripeEntities.TaxItem
	var totalCartPrice decimal.Decimal

	for _, item := range getActiveCarItems.Items {
		taxItem := stripeEntities.TaxItem{
			// Stripe requires to provide the amount of the product with the no.of pieces being bought
			Price:     item.Price.Mul(decimal.NewFromInt(int64(item.Quantity))),
			Quantity:  item.Quantity,
			Reference: item.SKU,
		}

		if item.StripeTaxCode != nil {
			taxItem.TaxCode = *item.StripeTaxCode
		}

		taxItems = append(taxItems, taxItem)

		totalCartPrice = totalCartPrice.Add(item.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
	}

	toAddress := stripeEntities.Address{
		State:      address.StateCode,
		PostalCode: address.PostalCode,
		Country:    address.CountryCode,
	}

	if address.City != nil {
		toAddress.City = *address.City
	}

	res, err := s.stripeClient.CalculateTax(ctx, &stripeEntities.CalculateTaxRequest{
		ShippingAmount: shippingRate.Amount,
		FromAddress: stripeEntities.Address{
			City:       req.Body.WarehouseAddress.City,
			State:      req.Body.WarehouseAddress.StateCode,
			PostalCode: req.Body.WarehouseAddress.PostalCode,
			Country:    req.Body.WarehouseAddress.CountryCode,
		},
		ToAddress: toAddress,
		TaxItems:  taxItems,
	})
	if err != nil {
		s.log.Errorf("Error calculating tax: %v", err)
		return nil, err
	}

	// convert the tax rate from minor units back to major units for human readability
	res.Tax = res.Tax.Div(decimal.NewFromInt(100))
	res.TotalAmount = res.TotalAmount.Div(decimal.NewFromInt(100))

	// update cart with tax & shipping rate
	err = s.repo.UpdateCartShippingAndTaxRate(
		ctx,
		getActiveCarItems.Items[0].CartID.String(),
		shippingRate.Id,
		decimal.NewFromFloat(res.Tax.InexactFloat64()),
		res.Currency,
		res.Breakdown,
	)
	if err != nil {
		s.log.Errorf("Error updating cart with tax rate: %v", err)
		return nil, moduleErrors.NewAPIError("CART_ERROR_UPDATING_TAX_RATE")
	}

	response := entities.GetTaxRateResponse{
		Tax:          decimal.NewFromFloat(res.Tax.InexactFloat64()),
		Total:        decimal.NewFromFloat(res.TotalAmount.InexactFloat64()),
		ShippingRate: decimal.NewFromFloat(shippingRate.Amount.InexactFloat64()),
		Subtotal:     totalCartPrice,
		Currency:     res.Currency,
	}

	// Cache the response
	responseBytes, err := json.Marshal(response)
	if err == nil {
		_ = s.cache.Set(ctx, cacheKey, responseBytes, 5*time.Minute)
	}

	return &response, nil
}

// swagger:route POST /cart/shipping-rates carts GetShippingRateRequest
//
// # Get Shipping Rates
// ### Get the shipping rates for the cart
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetShippingRateResponse Shipping rate retrieved successfully
//	400: DefaultError Bad Request
//	500: DefaultError Internal Server Error
func (s *service) GetShippingRate(ctx context.Context, req *entities.GetShippingRateRequest) (*entities.GetShippingRateResponse, error) {
	var maxLength, maxWidth, totalHeight, totalWeight decimal.Decimal
	var lengths, widths, allHeights, allWeights []decimal.Decimal
	var cartId uuid.UUID

	customerID, err := uuid.Parse(sharedMeta.XCustomerID(ctx))
	if err != nil {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	address, err := s.addressClient.GetAddress(ctx, &addressEntities.GetAddressRequest{
		AddressID: req.Body.AddressID,
	})
	if err != nil {
		s.log.Errorf("Error retrieving address: %v", err)
		return nil, err
	}

	getActiveCarItems, err := s.GetCartItems(ctx)
	if err != nil || getActiveCarItems == nil || len(getActiveCarItems.Items) == 0 {
		return nil, moduleErrors.NewAPIError("CART_IS_EMPTY")
	}

	for _, item := range getActiveCarItems.Items {
		// TODO revise package calculation
		if item.Length == nil || item.Width == nil || item.Height == nil || item.Weight == nil {
			for i := 0; i < item.Quantity; i++ {
				if item.Length != nil {
					lengths = append(lengths, *item.Length)
				}
				if item.Width != nil {
					widths = append(widths, *item.Width)
				}
				if item.Height != nil {
					allHeights = append(allHeights, *item.Height)
				}
				if item.Weight != nil {
					allWeights = append(allWeights, *item.Weight)
				}
			}
		}
	}

	// assuming all items in the cart belong to the same cart
	cartId = getActiveCarItems.Items[0].CartID

	cacheKey := getShippingRateCacheKey(req.Body.AddressID.String(), customerID.String(), cartId.String())

	cachedResponse, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cachedResponse != nil {
		var response entities.GetShippingRateResponse
		if err := json.Unmarshal(cachedResponse.([]byte), &response); err == nil {
			return &response, nil
		}
	}

	// considering the items are packaged in a stack-wise fashion (one on top of the other)
	// TODO: improve considering the actual packaging style
	maxLength = sharedDecimal.MaxDecimal(lengths...)
	maxWidth = sharedDecimal.MaxDecimal(widths...)
	totalHeight = sharedDecimal.SumDecimals(allHeights...)
	totalWeight = sharedDecimal.SumDecimals(allWeights...)

	toAddress := shipengineEntities.ShippingAddress{
		State:   address.StateCode,
		Zip:     address.PostalCode,
		Country: address.CountryCode,
	}

	if address.City != nil {
		toAddress.City = *address.City
	}

	shipengineEstimates, err := s.shipengineClient.GetRatesEstimate(ctx,
		// From address
		shipengineEntities.ShippingAddress{
			City:    req.Body.WarehouseAddress.City,
			State:   req.Body.WarehouseAddress.StateCode,
			Zip:     req.Body.WarehouseAddress.PostalCode,
			Country: req.Body.WarehouseAddress.CountryCode,
		},
		// To address
		// Assuming all items are being shipped to the same address as a single package
		toAddress,
		// Package dimensions
		shipengineEntities.Dimensions{
			Length: maxLength,
			Width:  maxWidth,
			Height: totalHeight,
			Weight: totalWeight,
		})
	if err != nil {
		return nil, err
	}

	if len(shipengineEstimates) == 0 {
		return nil, moduleErrors.NewAPIError("CART_NO_SHIPPING_RATES_FOUND")
	}

	allRates := make(map[string][]shipengineEntities.EstimateRatesResponse)
	bestRatesPerCarrier := make(map[string]shipengineEntities.EstimateRatesResponse)

	// group rates by carried id. Looks like "se-1140635"
	for _, estimate := range shipengineEstimates {
		if estimate.EstimatedDeliveryDate.Valid {
			allRates[estimate.CarrierID] = append(allRates[estimate.CarrierID], estimate)
		}
	}

	// for each carrier, find the best rate i.e lowest shipping amount
	for carrierID, rates := range allRates {
		var bestRate shipengineEntities.EstimateRatesResponse

		for _, rate := range rates {
			if rate.ShippingAmount.Amount > 0 {
				if bestRate.ShippingAmount.Amount == 0 || rate.ShippingAmount.Amount < bestRate.ShippingAmount.Amount {
					bestRate = rate
				}
			}
		}
		bestRatesPerCarrier[carrierID] = bestRate
	}

	var shippingRates []entities.CartShippingRate

	for _, bestRate := range bestRatesPerCarrier {
		shippingRates = append(shippingRates, entities.CartShippingRate{
			Id:                    uuid.New(),
			CartID:                cartId,
			AddressID:             req.Body.AddressID,
			Amount:                decimal.NewFromFloat(bestRate.ShippingAmount.Amount),
			Currency:              bestRate.ShippingAmount.Currency,
			CarrierName:           bestRate.CarrierFriendlyName,
			CarrierCode:           bestRate.CarrierCode,
			ServiceType:           bestRate.ServiceType,
			ServiceCode:           bestRate.ServiceCode,
			EstimatedDeliveryDate: bestRate.EstimatedDeliveryDate.Time,
		})
	}

	// save the shipping rates to the database
	err = s.repo.CreateCartShippingRates(ctx, shippingRates)
	if err != nil {
		s.log.Errorf("Error saving shipping rate: %v", err)
		return nil, moduleErrors.NewAPIError("CART_ERROR_UPDATING_SHIPPING_RATE")
	}

	response := &entities.GetShippingRateResponse{Rates: shippingRates}

	// Cache the response
	responseBytes, err := json.Marshal(response)
	if err == nil {
		_ = s.cache.Set(ctx, cacheKey, responseBytes, 5*time.Minute)
	}

	return response, nil
}

func (s *service) GetShippingRateByID(ctx context.Context, shippingRateID uuid.UUID) (*entities.CartShippingRate, error) {
	shippingRate, err := s.repo.GetShippingRate(ctx, shippingRateID)
	if err != nil {
		s.log.Errorf("Error retrieving shipping rate: %v", err)
		return nil, moduleErrors.NewAPIError("CART_ERROR_GETTING_SHIPPING_RATE")
	}

	return shippingRate, nil
}

func (s *service) GetCart(ctx context.Context) (*entities.Cart, error) {
	customerID := sharedMeta.XCustomerID(ctx)
	if customerID == "" {
		return nil, moduleErrors.NewAPIError("CUSTOMER_ID_REQUIRED")
	}

	cart, err := s.repo.GetActiveCart(ctx, customerID)
	if err != nil {
		s.log.Errorf("Error retrieving active cart: %v", err)
		return nil, moduleErrors.NewAPIError("CART_ERROR_GETTING_CART")
	}

	return cart, nil
}

func getShippingRateCacheKey(addressID, customerID, cartID string) string {
	return fmt.Sprintf("shipping_rate_%s_%s_%s", addressID, customerID, cartID)
}

func getTaxRateCacheKey(addressID, customerID, shippingRateID string) string {
	return fmt.Sprintf("tax_rate_%s_%s_%s", addressID, customerID, shippingRateID)
}
