package storage

import (
	"sync"
	"time"
)

var onceBudget sync.Once
var budgetStore BudgetStore

type BudgetStore struct {
	budgets map[int]BudgetInfo
	mu      sync.RWMutex
}

type BudgetInfo struct {
	Id          int
	Title       string
	Description string
	CreatedAt   time.Time
	StartAt     time.Time
	EndAt       time.Time
	Actual      int
	Target      int
	Currency    string
	Author      int
}

func initBudgetStore() {
	budgetStore = BudgetStore{
		budgets: make(map[int]BudgetInfo),
	}
}

func NewBudgetStore() {
	onceBudget.Do(func() {
		initBudgetStore()
	})
}

func GetBudgetByID(id int) (BudgetInfo, bool) {
	budgetStore.mu.RLock()
	defer budgetStore.mu.RUnlock()
	budget, ok := budgetStore.budgets[id]
	return budget, ok
}

func GetBudgetIDsByUserID(userID int) []int {
	budgetStore.mu.RLock()
	defer budgetStore.mu.RUnlock()
	var ids []int
	for id, budget := range budgetStore.budgets {
		if budget.Author == userID {
			ids = append(ids, id)
		}
	}
	return ids
}

func GetBudgetByIDAndUserID(budgetID int, userID int) (BudgetInfo, bool) {
	budgetStore.mu.RLock()
	defer budgetStore.mu.RUnlock()
	budget, ok := budgetStore.budgets[budgetID]
	if !ok {
		return BudgetInfo{}, false
	}
	if budget.Author != userID {
		return BudgetInfo{}, false
	}
	return budget, true
}

func AddBudget(budget BudgetInfo) int {
	budgetStore.mu.Lock()
	defer budgetStore.mu.Unlock()
	id := len(budgetStore.budgets)
	budget.Id = id
	budgetStore.budgets[id] = budget
	return id
}

func DeleteBudgetByIDAndUserID(budgetID int, userID int) bool {
	budgetStore.mu.Lock()
	defer budgetStore.mu.Unlock()
	budget, ok := budgetStore.budgets[budgetID]
	if !ok {
		return false
	}
	if budget.Author != userID {
		return false
	}
	delete(budgetStore.budgets, budgetID)
	return true
}
