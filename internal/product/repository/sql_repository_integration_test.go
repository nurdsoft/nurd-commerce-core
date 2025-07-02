package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nurdsoft/nurd-commerce-core/internal/product/entities"
	"github.com/nurdsoft/nurd-commerce-core/shared/testutils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIntegrationTest(t *testing.T) (Repository, func()) {
	testDB := testutils.SetupTestDB(t, nil)
	gormDB := testDB.GetGormDB()

	productID := uuid.New()
	variant1ID := uuid.New()
	variant2ID := uuid.New()
	variant3ID := uuid.New()

	// Create test product
	err := gormDB.Exec(`
		INSERT INTO products (id, name, description, created_at, updated_at)
		VALUES (?, 'Test Product', 'Test Description', NOW(), NOW())
	`, productID).Error
	require.NoError(t, err)

	// Create test product variants with different attributes
	attributes1 := `{"color": "red", "size": "large"}`
	attributes2 := `{"color": "blue", "size": "medium"}`
	attributes3 := `{"color": "green", "size": "small"}`

	// Variant 1: Red, Large, $50
	err = gormDB.Exec(`
		INSERT INTO product_variants (id, product_id, sku, name, description, price, currency, attributes, created_at, updated_at)
		VALUES (?, ?, ?, 'Red Large Shirt', 'A red large shirt', 50.00, 'USD', ?, NOW(), NOW())
	`, variant1ID, productID, "RED-LARGE-001", attributes1).Error
	require.NoError(t, err)

	// Variant 2: Blue, Medium, $40
	err = gormDB.Exec(`
		INSERT INTO product_variants (id, product_id, sku, name, description, price, currency, attributes, created_at, updated_at)
		VALUES (?, ?, ?, 'Blue Medium Shirt', 'A blue medium shirt', 40.00, 'USD', ?, NOW(), NOW())
	`, variant2ID, productID, "BLUE-MED-002", attributes2).Error
	require.NoError(t, err)

	// Variant 3: Green, Small, $30
	err = gormDB.Exec(`
		INSERT INTO product_variants (id, product_id, sku, name, description, price, currency, attributes, created_at, updated_at)
		VALUES (?, ?, ?, 'Green Small Shirt', 'A green small shirt', 30.00, 'USD', ?, NOW(), NOW())
	`, variant3ID, productID, "GREEN-SMALL-003", attributes3).Error
	require.NoError(t, err)

	repo := New(testDB.GetSQLDB(), gormDB)

	cleanup := func() {
		gormDB.Exec("DELETE FROM product_variants WHERE id IN (?, ?, ?)", variant1ID, variant2ID, variant3ID)
		gormDB.Exec("DELETE FROM products WHERE id = ?", productID)
	}

	return repo, cleanup
}

func TestListVariants(t *testing.T) {
	repo, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := testutils.CreateTestContext(t)

	t.Run("List all variants with default pagination", func(t *testing.T) {
		req := &entities.ListProductVariantsRequest{
			Page:     1,
			PageSize: 10,
		}

		resp, err := repo.ListVariants(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 3)
		assert.Equal(t, 3, resp.Pagination.Total)
		assert.Equal(t, 1, resp.Pagination.TotalPages)
		assert.Equal(t, 1, resp.Pagination.Page)
		assert.Equal(t, 10, resp.Pagination.PageSize)
	})

	t.Run("List variants with pagination", func(t *testing.T) {
		req := &entities.ListProductVariantsRequest{
			Page:     1,
			PageSize: 2,
		}

		resp, err := repo.ListVariants(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 2)
		assert.Equal(t, 3, resp.Pagination.Total)
		assert.Equal(t, 2, resp.Pagination.TotalPages)
		assert.Equal(t, 1, resp.Pagination.Page)
		assert.Equal(t, 2, resp.Pagination.PageSize)
	})

	t.Run("List variants with search filter", func(t *testing.T) {
		search := "red"
		req := &entities.ListProductVariantsRequest{
			Page:     1,
			PageSize: 10,
			Search:   &search,
		}

		resp, err := repo.ListVariants(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, 1, resp.Pagination.Total)
		assert.Contains(t, resp.Data[0].Name, "Red")
	})

	t.Run("List variants with price filter", func(t *testing.T) {
		minPrice := decimal.NewFromInt(35)
		maxPrice := decimal.NewFromInt(45)
		req := &entities.ListProductVariantsRequest{
			Page:     1,
			PageSize: 10,
			MinPrice: &minPrice,
			MaxPrice: &maxPrice,
		}

		resp, err := repo.ListVariants(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, 1, resp.Pagination.Total)
		assert.Zero(t, resp.Data[0].Price.Compare(decimal.NewFromInt(40)))
	})

	t.Run("List variants with JSON attributes filter", func(t *testing.T) {
		req := &entities.ListProductVariantsRequest{
			Page:       1,
			PageSize:   10,
			Attributes: map[string]string{"color": "red", "size": "large"},
		}

		resp, err := repo.ListVariants(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, 1, resp.Pagination.Total)
		assert.Equal(t, "RED-LARGE-001", resp.Data[0].SKU)
	})

	t.Run("List variants with sorting", func(t *testing.T) {
		sortBy := "price"
		sortOrder := "asc"
		req := &entities.ListProductVariantsRequest{
			Page:      1,
			PageSize:  10,
			SortBy:    &sortBy,
			SortOrder: &sortOrder,
		}

		resp, err := repo.ListVariants(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 3)
		assert.Equal(t, 3, resp.Pagination.Total)

		// Check that prices are in ascending order
		assert.Zero(t, resp.Data[0].Price.Compare(decimal.NewFromInt(30))) // Green Small
		assert.Zero(t, resp.Data[1].Price.Compare(decimal.NewFromInt(40))) // Blue Medium
		assert.Zero(t, resp.Data[2].Price.Compare(decimal.NewFromInt(50))) // Red Large
	})

	t.Run("List variants with combined filters", func(t *testing.T) {
		search := "shirt"
		minPrice := decimal.NewFromInt(35)
		req := &entities.ListProductVariantsRequest{
			Page:     1,
			PageSize: 10,
			Search:   &search,
			MinPrice: &minPrice,
		}

		resp, err := repo.ListVariants(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 2) // Blue Medium and Red Large (both >= $35)
		assert.Equal(t, 2, resp.Pagination.Total)
	})

	t.Run("Empty result with no matching search", func(t *testing.T) {
		search := "nonexistent"
		req := &entities.ListProductVariantsRequest{
			Page:     1,
			PageSize: 10,
			Search:   &search,
		}

		resp, err := repo.ListVariants(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 0)
		assert.Equal(t, 0, resp.Pagination.Total)
		assert.Equal(t, 0, resp.Pagination.TotalPages)
	})

	t.Run("Empty result with no attributes", func(t *testing.T) {
		req := &entities.ListProductVariantsRequest{
			Page:       1,
			PageSize:   10,
			Attributes: map[string]string{"color": "black"}, // No variants with black color
		}

		resp, err := repo.ListVariants(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 0)
		assert.Equal(t, 0, resp.Pagination.Total)
		assert.Equal(t, 0, resp.Pagination.TotalPages)
	})

	t.Run("Invalid page number defaults to 1", func(t *testing.T) {
		req := &entities.ListProductVariantsRequest{
			Page:     0, // Invalid
			PageSize: 10,
		}

		resp, err := repo.ListVariants(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 3)
		assert.Equal(t, 1, resp.Pagination.Page)
	})

	t.Run("Invalid page size defaults to 10", func(t *testing.T) {
		req := &entities.ListProductVariantsRequest{
			Page:     1,
			PageSize: 0, // Invalid
		}

		resp, err := repo.ListVariants(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 3)
		assert.Equal(t, 10, resp.Pagination.PageSize)
	})

	t.Run("Page size exceeds maximum is capped at 100", func(t *testing.T) {
		req := &entities.ListProductVariantsRequest{
			Page:     1,
			PageSize: 150, // Exceeds max
		}

		resp, err := repo.ListVariants(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Data, 3)
		assert.Equal(t, 100, resp.Pagination.PageSize)
	})
}
