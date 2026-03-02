package storage

import (
	"strconv"
	"sync"
	"time"
)

var onceBudget sync.Once
var budgetStore BudgetStore

type BudgetStore struct {
	budgets map[string]BudgetInfo
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
}

func initBudgetStore() {
	budgetStore = BudgetStore{
		budgets: make(map[string]BudgetInfo),
	}
}

func NewBudgetStore() {
	onceBudget.Do(func() {
		initBudgetStore()
	})
}

func DoBudgetWithLock(f func()) {
	budgetStore.mu.Lock()
	defer budgetStore.mu.Unlock()
	f()
}

func GetBudgetByID(id string) (BudgetInfo, bool) {
	budgetStore.mu.RLock()
	defer budgetStore.mu.RUnlock()
	budget, ok := budgetStore.budgets[id]
	return budget, ok
}

func AddBudget(budget BudgetInfo) string {
	budgetStore.mu.Lock()
	defer budgetStore.mu.Unlock()
	id := len(budgetStore.budgets)
	budget.Id = id
	key := strconv.Itoa(id)
	budgetStore.budgets[key] = budget
	return key
}
