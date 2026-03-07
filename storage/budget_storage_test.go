package storage

import (
    "sync"
    "testing"
)

func setupBudgetStoreTest() {
    onceBudget = sync.Once{}
    NewBudgetStore()
}

func TestAddBudgetAndGetBudgetByID(t *testing.T) {
    setupBudgetStoreTest()

    id := AddBudget(BudgetInfo{Title: "Trip", Target: 50000, Currency: "RUB", Author: 1})
    if id != 0 {
        t.Fatalf("expected first budget id 0, got %d", id)
    }

    budget, ok := GetBudgetByID(id)
    if !ok {
        t.Fatal("expected budget to exist")
    }
    if budget.Title != "Trip" || budget.Author != 1 {
        t.Fatalf("unexpected budget: %+v", budget)
    }
}

func TestGetBudgetIDsByUserID(t *testing.T) {
    setupBudgetStoreTest()

    AddBudget(BudgetInfo{Title: "A", Author: 1})
    AddBudget(BudgetInfo{Title: "B", Author: 2})
    AddBudget(BudgetInfo{Title: "C", Author: 1})

    ids := GetBudgetIDsByUserID(1)
    if len(ids) != 2 {
        t.Fatalf("expected 2 budget ids for user 1, got %d", len(ids))
    }
}

func TestGetBudgetByIDAndUserID(t *testing.T) {
    setupBudgetStoreTest()

    id := AddBudget(BudgetInfo{Title: "Home", Author: 5})

    _, ok := GetBudgetByIDAndUserID(id, 4)
    if ok {
        t.Fatal("expected access to fail for another user")
    }

    budget, ok := GetBudgetByIDAndUserID(id, 5)
    if !ok {
        t.Fatal("expected access for author")
    }
    if budget.Title != "Home" {
        t.Fatalf("unexpected budget title: %q", budget.Title)
    }
}

func TestDeleteBudgetByIDAndUserID(t *testing.T) {
    setupBudgetStoreTest()

    id := AddBudget(BudgetInfo{Title: "Delete me", Author: 9})

    deleted := DeleteBudgetByIDAndUserID(id, 8)
    if deleted {
        t.Fatal("expected delete to fail for non-author")
    }

    deleted = DeleteBudgetByIDAndUserID(id, 9)
    if !deleted {
        t.Fatal("expected delete to succeed for author")
    }

    _, ok := GetBudgetByID(id)
    if ok {
        t.Fatal("expected budget to be deleted")
    }
}
