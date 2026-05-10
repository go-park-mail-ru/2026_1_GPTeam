package web_helpers

import "net/http"

type AnalysisSummary struct {
	TotalBudgetLimit float64 `json:"total_budget_limit"`
	TotalBudgetSpent float64 `json:"total_budget_spent"`
	TotalBudgetFree  float64 `json:"total_budget_free"`
	IncomeTotal      float64 `json:"income_total"`
	ExpenseTotal     float64 `json:"expense_total"`
	Savings          float64 `json:"savings"`
}

type AnalysisBudgetItem struct {
	ID         int      `json:"id"`
	Title      string   `json:"title"`
	Categories []string `json:"categories"`
	Target     float64  `json:"target"`
	Actual     float64  `json:"actual"`
	Remaining  float64  `json:"remaining"`
	Progress   float64  `json:"progress"`
	Currency   string   `json:"currency"`
}

type AnalysisCategoryItem struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Share    float64 `json:"share"`
}

type AnalysisTimelineItem struct {
	Label   string  `json:"label"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

type AnalysisResponse struct {
	SimpleResponse
	Period      string                 `json:"period"`
	PeriodLabel string                 `json:"period_label"`
	Summary     AnalysisSummary        `json:"summary"`
	Budgets     []AnalysisBudgetItem   `json:"budgets"`
	Categories  []AnalysisCategoryItem `json:"categories"`
	Timeline    []AnalysisTimelineItem `json:"timeline"`
}

func NewAnalysisResponse(period string, periodLabel string, summary AnalysisSummary, budgets []AnalysisBudgetItem, categories []AnalysisCategoryItem, timeline []AnalysisTimelineItem) AnalysisResponse {
	return AnalysisResponse{
		SimpleResponse: SimpleResponse{
			Code:    http.StatusOK,
			Message: "Ok",
		},
		Period:      period,
		PeriodLabel: periodLabel,
		Summary:     summary,
		Budgets:     budgets,
		Categories:  categories,
		Timeline:    timeline,
	}
}
