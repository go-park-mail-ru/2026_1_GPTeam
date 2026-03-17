package models_test

import (
	"testing"
	"time"

	models2 "github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2026_1_GPTeam/models"
)

func setupBudgetStoreTest(t *testing.T) {
	t.Helper()
	models.NewBudgetStore()
}

func makeBudget(author int) models2.BudgetModel {
	now := time.Now().UTC().Truncate(time.Second)
	return models2.BudgetModel{
		Title:       "Test budget",
		Description: "test description",
		CreatedAt:   now,
		StartAt:     now.Add(time.Hour),
		EndAt:       now.Add(2 * time.Hour),
		Actual:      1500,
		Target:      3000,
		Currency:    "RUB",
		Author:      author,
	}
}

func TestAddBudgetAndGetBudgetByID(t *testing.T) {
	t.Parallel()
	setupBudgetStoreTest(t)

	budget := makeBudget(101)
	id := models.AddBudget(budget)

	got, ok := models.GetBudgetByID(id)
	require.True(t, ok)

	assert.Equal(t, id, got.Id)
	assert.Equal(t, budget.Title, got.Title)
	assert.Equal(t, budget.Description, got.Description)
	assert.Equal(t, budget.CreatedAt, got.CreatedAt)
	assert.Equal(t, budget.StartAt, got.StartAt)
	assert.Equal(t, budget.EndAt, got.EndAt)
	assert.Equal(t, budget.Actual, got.Actual)
	assert.Equal(t, budget.Target, got.Target)
	assert.Equal(t, budget.Currency, got.Currency)
	assert.Equal(t, budget.Author, got.Author)
}

func TestGetBudgetByID(t *testing.T) {
	t.Parallel()
	setupBudgetStoreTest(t)

	id := models.AddBudget(makeBudget(10))

	cases := []struct {
		name   string
		id     int
		wantOK bool
	}{
		{
			name:   "существующий бюджет",
			id:     id,
			wantOK: true,
		},
		{
			name:   "несуществующий ID",
			id:     999999,
			wantOK: false,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, ok := models.GetBudgetByID(c.id)
			assert.Equal(t, c.wantOK, ok)
		})
	}
}

func TestGetBudgetIDsByUserID(t *testing.T) {
	t.Parallel()
	setupBudgetStoreTest(t)

	id1 := models.AddBudget(makeBudget(1))
	id2 := models.AddBudget(makeBudget(1))
	models.AddBudget(makeBudget(2))

	cases := []struct {
		name    string
		userID  int
		wantIDs []int
		wantLen int
	}{
		{
			name:    "пользователь с двумя бюджетами",
			userID:  1,
			wantIDs: []int{id1, id2},
			wantLen: 2,
		},
		{
			name:    "пользователь с одним бюджетом",
			userID:  2,
			wantLen: 1,
		},
		{
			name:    "пользователь без бюджетов",
			userID:  999,
			wantLen: 0,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			ids := models.GetBudgetIDsByUserID(c.userID)
			assert.Len(t, ids, c.wantLen)
			for _, wantID := range c.wantIDs {
				assert.Contains(t, ids, wantID)
			}
		})
	}
}

func TestGetBudgetByIDAndUserID(t *testing.T) {
	t.Parallel()
	setupBudgetStoreTest(t)

	id := models.AddBudget(makeBudget(77))

	cases := []struct {
		name     string
		budgetID int
		userID   int
		wantOK   bool
	}{
		{
			name:     "владелец получает свой бюджет",
			budgetID: id,
			userID:   77,
			wantOK:   true,
		},
		{
			name:     "чужой userID → false",
			budgetID: id,
			userID:   88,
			wantOK:   false,
		},
		{
			name:     "несуществующий budgetID → false",
			budgetID: 999999,
			userID:   77,
			wantOK:   false,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, ok := models.GetBudgetByIDAndUserID(c.budgetID, c.userID)
			assert.Equal(t, c.wantOK, ok)
			if c.wantOK {
				assert.Equal(t, c.budgetID, got.Id)
				assert.Equal(t, c.userID, got.Author)
			} else {
				assert.Equal(t, models2.BudgetModel{}, got)
			}
		})
	}
}

func TestDeleteBudgetByIDAndUserID(t *testing.T) {
	t.Parallel()
	setupBudgetStoreTest(t)

	cases := []struct {
		name         string
		authorID     int
		deleteUserID int
		budgetID     int
		useRealID    bool
		wantOK       bool
		wantExists   bool
	}{
		{
			name:         "владелец удаляет свой бюджет",
			authorID:     77,
			deleteUserID: 77,
			useRealID:    true,
			wantOK:       true,
			wantExists:   false,
		},
		{
			name:         "чужой userID → false, бюджет остаётся",
			authorID:     77,
			deleteUserID: 88,
			useRealID:    true,
			wantOK:       false,
			wantExists:   true,
		},
		{
			name:         "несуществующий budgetID → false",
			authorID:     77,
			deleteUserID: 77,
			budgetID:     999999,
			useRealID:    false,
			wantOK:       false,
			wantExists:   false,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			id := models.AddBudget(makeBudget(c.authorID))
			targetID := id
			if !c.useRealID {
				targetID = c.budgetID
			}

			ok := models.DeleteBudgetByIDAndUserID(targetID, c.deleteUserID)
			assert.Equal(t, c.wantOK, ok)

			if c.useRealID {
				_, exists := models.GetBudgetByID(id)
				assert.Equal(t, c.wantExists, exists)
			}
		})
	}
}
