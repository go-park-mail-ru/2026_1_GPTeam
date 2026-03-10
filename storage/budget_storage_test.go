package storage_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"main/storage"
)

func setupBudgetStoreTest(t *testing.T) {
	t.Helper()
	storage.NewBudgetStore()
}

func TestAddBudgetAndGetBudgetByID(t *testing.T) {
	setupBudgetStoreTest(t)

	createdAt := time.Now().UTC().Truncate(time.Second)
	startAt := createdAt.Add(time.Hour)
	endAt := createdAt.Add(2 * time.Hour)

	id := storage.AddBudget(storage.BudgetInfo{
		Title:       "Test budget",
		Description: "test description",
		CreatedAt:   createdAt,
		StartAt:     startAt,
		EndAt:       endAt,
		Actual:      1500,
		Target:      3000,
		Currency:    "RUB",
		Author:      101,
	})

	budget, ok := storage.GetBudgetByID(id)
	require.True(t, ok)

	assert.Equal(t, id, budget.Id)
	assert.Equal(t, "Test budget", budget.Title)
	assert.Equal(t, "test description", budget.Description)
	assert.Equal(t, createdAt, budget.CreatedAt)
	assert.Equal(t, startAt, budget.StartAt)
	assert.Equal(t, endAt, budget.EndAt)
	assert.Equal(t, 1500, budget.Actual)
	assert.Equal(t, 3000, budget.Target)
	assert.Equal(t, "RUB", budget.Currency)
	assert.Equal(t, 101, budget.Author)
}

func TestGetBudgetIDsByUserID_ReturnsOnlyUserBudgetIDs(t *testing.T) {
	setupBudgetStoreTest(t)

	id1 := storage.AddBudget(storage.BudgetInfo{
		Title:       "Budget 1",
		Description: "first",
		CreatedAt:   time.Now().UTC(),
		StartAt:     time.Now().UTC(),
		EndAt:       time.Now().UTC().Add(time.Hour),
		Actual:      100,
		Target:      200,
		Currency:    "RUB",
		Author:      1,
	})
	id2 := storage.AddBudget(storage.BudgetInfo{
		Title:       "Budget 2",
		Description: "second",
		CreatedAt:   time.Now().UTC(),
		StartAt:     time.Now().UTC(),
		EndAt:       time.Now().UTC().Add(time.Hour),
		Actual:      300,
		Target:      500,
		Currency:    "RUB",
		Author:      1,
	})
	storage.AddBudget(storage.BudgetInfo{
		Title:       "Foreign budget",
		Description: "third",
		CreatedAt:   time.Now().UTC(),
		StartAt:     time.Now().UTC(),
		EndAt:       time.Now().UTC().Add(time.Hour),
		Actual:      999,
		Target:      1000,
		Currency:    "RUB",
		Author:      2,
	})

	ids := storage.GetBudgetIDsByUserID(1)
	require.Len(t, ids, 2)
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
}

func TestGetBudgetByIDAndUserID_ReturnsOwnedBudget(t *testing.T) {
	setupBudgetStoreTest(t)

	id := storage.AddBudget(storage.BudgetInfo{
		Title:       "Owned budget",
		Description: "owned",
		CreatedAt:   time.Now().UTC(),
		StartAt:     time.Now().UTC(),
		EndAt:       time.Now().UTC().Add(time.Hour),
		Actual:      10,
		Target:      20,
		Currency:    "RUB",
		Author:      77,
	})

	budget, ok := storage.GetBudgetByIDAndUserID(id, 77)
	require.True(t, ok)
	assert.Equal(t, id, budget.Id)
	assert.Equal(t, 77, budget.Author)
}

func TestDeleteBudgetByIDAndUserID_DeletesOwnedBudget(t *testing.T) {
	setupBudgetStoreTest(t)

	id := storage.AddBudget(storage.BudgetInfo{
		Title:       "Owned budget",
		Description: "owned",
		CreatedAt:   time.Now().UTC(),
		StartAt:     time.Now().UTC(),
		EndAt:       time.Now().UTC().Add(time.Hour),
		Actual:      10,
		Target:      20,
		Currency:    "RUB",
		Author:      77,
	})

	ok := storage.DeleteBudgetByIDAndUserID(id, 77)
	assert.True(t, ok)

	_, exists := storage.GetBudgetByID(id)
	assert.False(t, exists)
}

func TestDeleteBudgetByIDAndUserID_RejectsWrongUser(t *testing.T) {
	setupBudgetStoreTest(t)

	id := storage.AddBudget(storage.BudgetInfo{
		Title:       "Protected budget",
		Description: "protected",
		CreatedAt:   time.Now().UTC(),
		StartAt:     time.Now().UTC(),
		EndAt:       time.Now().UTC().Add(time.Hour),
		Actual:      10,
		Target:      20,
		Currency:    "RUB",
		Author:      77,
	})

	ok := storage.DeleteBudgetByIDAndUserID(id, 88)
	assert.False(t, ok)

	_, exists := storage.GetBudgetByID(id)
	assert.True(t, exists)
}

func TestGetBudgetByID_ReturnsFalseForMissingBudget(t *testing.T) {
	setupBudgetStoreTest(t)

	_, ok := storage.GetBudgetByID(999999)
	assert.False(t, ok)
}
