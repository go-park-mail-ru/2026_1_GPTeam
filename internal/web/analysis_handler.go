package web

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/context_helper"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

type AnalysisHandler struct {
	analysisApp application.AnalysisUseCase
}

func NewAnalysisHandler(analysisApp application.AnalysisUseCase) *AnalysisHandler {
	return &AnalysisHandler{analysisApp: analysisApp}
}

func (obj *AnalysisHandler) Get(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLoggerWithRequestId(r.Context())
	log.Info("get analysis request")

	authUser, ok := web_helpers.GetAuthUser(r)
	if !ok {
		log.Warn("user unauthorized")
		response := web_helpers.NewUnauthorizedErrorResponse()
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	query := r.URL.Query()
	startDate, err := application.ParseAnalysisStartDate(query.Get("start_date"), time.Now().Location())
	if err != nil {
		log.Warn("invalid analysis start_date", zap.Error(err))
		response := web_helpers.NewBadRequestErrorResponse("Некорректная дата начала периода")
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	result, err := obj.analysisApp.Get(r.Context(), authUser, application.AnalysisParams{
		Period:    application.ParseAnalysisPeriod(query.Get("period")),
		StartDate: startDate,
		Now:       time.Now(),
	})
	if err != nil {
		log.Error("failed to build analysis", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	response := web_helpers.NewAnalysisResponse(
		string(result.Period),
		result.PeriodLabel,
		result.PeriodStart,
		result.PeriodEnd,
		toWebAnalysisSummary(result.Summary),
		toWebAnalysisBudgets(result.Budgets),
		toWebAnalysisCategories(result.Categories),
		toWebAnalysisTimeline(result.Timeline),
	)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func toWebAnalysisSummary(summary application.AnalysisSummary) web_helpers.AnalysisSummary {
	return web_helpers.AnalysisSummary{
		TotalBudgetLimit: summary.TotalBudgetLimit,
		TotalBudgetSpent: summary.TotalBudgetSpent,
		TotalBudgetFree:  summary.TotalBudgetFree,
		IncomeTotal:      summary.IncomeTotal,
		ExpenseTotal:     summary.ExpenseTotal,
		Savings:          summary.Savings,
	}
}

func toWebAnalysisBudgets(items []application.AnalysisBudgetItem) []web_helpers.AnalysisBudgetItem {
	result := make([]web_helpers.AnalysisBudgetItem, 0, len(items))
	for _, item := range items {
		result = append(result, web_helpers.AnalysisBudgetItem{
			ID:         item.ID,
			Title:      item.Title,
			Categories: item.Categories,
			Target:     item.Target,
			Actual:     item.Actual,
			Remaining:  item.Remaining,
			Progress:   item.Progress,
			Currency:   item.Currency,
		})
	}
	return result
}

func toWebAnalysisCategories(items []application.AnalysisCategoryItem) []web_helpers.AnalysisCategoryItem {
	result := make([]web_helpers.AnalysisCategoryItem, 0, len(items))
	for _, item := range items {
		result = append(result, web_helpers.AnalysisCategoryItem{
			Category: item.Category,
			Amount:   item.Amount,
			Share:    item.Share,
		})
	}
	return result
}

func toWebAnalysisTimeline(items []application.AnalysisTimelineItem) []web_helpers.AnalysisTimelineItem {
	result := make([]web_helpers.AnalysisTimelineItem, 0, len(items))
	for _, item := range items {
		result = append(result, web_helpers.AnalysisTimelineItem{
			Label:   item.Label,
			Income:  item.Income,
			Expense: item.Expense,
		})
	}
	return result
}
