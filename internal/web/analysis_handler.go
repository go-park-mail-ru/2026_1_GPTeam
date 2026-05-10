package web

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/context_helper"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/currency_converter"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"go.uber.org/zap"
)

type AnalysisHandler struct {
	budgetApp      application.BudgetUseCase
	transactionApp application.TransactionUseCase
	accountApp     application.AccountUseCase
}

type analysisTransaction struct {
	transaction models.TransactionModel
	rubValue    float64
}

type analysisRange struct {
	period string
	label  string
	start  time.Time
	end    time.Time
}

func NewAnalysisHandler(budgetApp application.BudgetUseCase, transactionApp application.TransactionUseCase, accountApp application.AccountUseCase) *AnalysisHandler {
	return &AnalysisHandler{
		budgetApp:      budgetApp,
		transactionApp: transactionApp,
		accountApp:     accountApp,
	}
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

	rangeInfo := parseAnalysisRange(r.URL.Query().Get("period"), time.Now())
	transactions, err := obj.getTransactionsInRange(r, authUser.Id, rangeInfo.start, rangeInfo.end)
	if err != nil {
		log.Error("failed to load transactions for analysis", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	budgets, summary, err := obj.getBudgetAnalysis(r, authUser, transactions, rangeInfo)
	if err != nil {
		log.Error("failed to build budget analysis", zap.Error(err))
		response := web_helpers.NewServerErrorResponse(context_helper.GetRequestIdFromContext(r.Context()))
		web_helpers.WriteResponseJSON(w, response.Code, response)
		return
	}

	categories, incomeTotal, expenseTotal := buildCategoryAnalysis(transactions)
	timeline := buildTimeline(rangeInfo, transactions)
	summary.IncomeTotal = round2(incomeTotal)
	summary.ExpenseTotal = round2(expenseTotal)
	summary.Savings = round2(incomeTotal - expenseTotal)

	response := web_helpers.NewAnalysisResponse(
		rangeInfo.period,
		rangeInfo.label,
		summary,
		budgets,
		categories,
		timeline,
	)
	web_helpers.WriteResponseJSON(w, response.Code, response)
}

func (obj *AnalysisHandler) getTransactionsInRange(r *http.Request, userId int, start time.Time, end time.Time) ([]analysisTransaction, error) {
	filters := repository.TransactionFilters{
		StartDate: &start,
		EndDate:   &end,
	}
	transactions, err := obj.transactionApp.Search(r.Context(), userId, filters)
	if err != nil {
		if errors.Is(err, repository.NothingInTableError) {
			return []analysisTransaction{}, nil
		}
		return nil, err
	}

	currencyCache := make(map[int]string)
	result := make([]analysisTransaction, 0, len(transactions))
	for _, transaction := range transactions {
		currency, exists := currencyCache[transaction.AccountId]
		if !exists {
			loadedCurrency, currencyErr := obj.accountApp.GetCurrencyByAccountId(r.Context(), transaction.AccountId)
			if currencyErr != nil {
				return nil, currencyErr
			}
			currency = loadedCurrency
			currencyCache[transaction.AccountId] = currency
		}
		result = append(result, analysisTransaction{
			transaction: transaction,
			rubValue:    currency_converter.ConvertToRub(transaction.Value, currency),
		})
	}

	return result, nil
}

func (obj *AnalysisHandler) getBudgetAnalysis(r *http.Request, authUser models.UserModel, transactions []analysisTransaction, rangeInfo analysisRange) ([]web_helpers.AnalysisBudgetItem, web_helpers.AnalysisSummary, error) {
	budgetIDs, err := obj.budgetApp.GetBudgetsOfUser(r.Context(), authUser)
	if err != nil {
		if !errors.Is(err, repository.NothingInTableError) {
			return nil, web_helpers.AnalysisSummary{}, err
		}
		budgetIDs = []int{}
	}

	budgetItems := make([]web_helpers.AnalysisBudgetItem, 0, len(budgetIDs))
	summary := web_helpers.AnalysisSummary{}

	for _, budgetID := range budgetIDs {
		budget, categories, detailErr := obj.budgetApp.GetById(r.Context(), budgetID, authUser)
		if detailErr != nil {
			if errors.Is(detailErr, repository.NothingInTableError) || errors.Is(detailErr, application.UserNotAuthorOfBudgetError) {
				continue
			}
			return nil, web_helpers.AnalysisSummary{}, detailErr
		}
		if !budgetOverlapsRange(budget, rangeInfo.start, rangeInfo.end) {
			continue
		}

		overlapStart, overlapEnd := getBudgetOverlap(budget, rangeInfo.start, rangeInfo.end)
		actual := calculateBudgetActual(transactions, categories, overlapStart, overlapEnd)
		remaining := math.Max(budget.Target-actual, 0)
		progress := 0.0
		if budget.Target > 0 {
			progress = math.Min(actual/budget.Target*100, 999)
		}

		budgetItems = append(budgetItems, web_helpers.AnalysisBudgetItem{
			ID:         budget.Id,
			Title:      strings.TrimSpace(budget.Title),
			Categories: categories,
			Target:     round2(budget.Target),
			Actual:     round2(actual),
			Remaining:  round2(remaining),
			Progress:   round2(progress),
			Currency:   budget.Currency,
		})
		summary.TotalBudgetLimit += budget.Target
		summary.TotalBudgetSpent += actual
	}

	summary.TotalBudgetLimit = round2(summary.TotalBudgetLimit)
	summary.TotalBudgetSpent = round2(summary.TotalBudgetSpent)
	summary.TotalBudgetFree = round2(math.Max(summary.TotalBudgetLimit-summary.TotalBudgetSpent, 0))

	sort.Slice(budgetItems, func(i, j int) bool {
		if budgetItems[i].Progress == budgetItems[j].Progress {
			return budgetItems[i].Title < budgetItems[j].Title
		}
		return budgetItems[i].Progress > budgetItems[j].Progress
	})

	return budgetItems, summary, nil
}

func buildCategoryAnalysis(transactions []analysisTransaction) ([]web_helpers.AnalysisCategoryItem, float64, float64) {
	categoryTotals := make(map[string]float64)
	incomeTotal := 0.0
	expenseTotal := 0.0

	for _, item := range transactions {
		amount := item.rubValue
		if strings.EqualFold(item.transaction.Type, "INCOME") {
			incomeTotal += amount
			continue
		}
		expenseTotal += amount
		category := strings.TrimSpace(item.transaction.Category)
		if category == "" {
			category = "Прочее"
		}
		categoryTotals[category] += amount
	}

	categories := make([]web_helpers.AnalysisCategoryItem, 0, len(categoryTotals))
	for category, amount := range categoryTotals {
		share := 0.0
		if expenseTotal > 0 {
			share = amount / expenseTotal * 100
		}
		categories = append(categories, web_helpers.AnalysisCategoryItem{
			Category: category,
			Amount:   round2(amount),
			Share:    round2(share),
		})
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Amount > categories[j].Amount
	})

	return categories, round2(incomeTotal), round2(expenseTotal)
}

func buildTimeline(rangeInfo analysisRange, transactions []analysisTransaction) []web_helpers.AnalysisTimelineItem {
	type bucket struct {
		label   string
		income  float64
		expense float64
	}

	orderedKeys := make([]string, 0)
	buckets := make(map[string]*bucket)
	ensureBucket := func(key string, label string) *bucket {
		item, exists := buckets[key]
		if !exists {
			item = &bucket{label: label}
			buckets[key] = item
			orderedKeys = append(orderedKeys, key)
		}
		return item
	}

	switch rangeInfo.period {
	case "year":
		cursor := time.Date(rangeInfo.start.Year(), rangeInfo.start.Month(), 1, 0, 0, 0, 0, rangeInfo.start.Location())
		for !cursor.After(rangeInfo.end) {
			key := cursor.Format("2006-01")
			ensureBucket(key, shortMonth(cursor.Month()))
			cursor = cursor.AddDate(0, 1, 0)
		}
	case "quarter":
		cursor := rangeInfo.start
		for !cursor.After(rangeInfo.end) {
			weekStart := startOfWeek(cursor)
			key := weekStart.Format("2006-01-02")
			ensureBucket(key, fmt.Sprintf("%02d %s", weekStart.Day(), shortMonth(weekStart.Month())))
			cursor = weekStart.AddDate(0, 0, 7)
		}
	default:
		cursor := rangeInfo.start
		for !cursor.After(rangeInfo.end) {
			key := cursor.Format("2006-01-02")
			ensureBucket(key, fmt.Sprintf("%02d", cursor.Day()))
			cursor = cursor.AddDate(0, 0, 1)
		}
	}

	for _, item := range transactions {
		date := item.transaction.TransactionDate.In(rangeInfo.start.Location())
		var key string
		switch rangeInfo.period {
		case "year":
			key = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location()).Format("2006-01")
		case "quarter":
			key = startOfWeek(date).Format("2006-01-02")
		default:
			key = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()).Format("2006-01-02")
		}
		bucketItem := buckets[key]
		if bucketItem == nil {
			continue
		}
		if strings.EqualFold(item.transaction.Type, "INCOME") {
			bucketItem.income += item.rubValue
		} else {
			bucketItem.expense += item.rubValue
		}
	}

	result := make([]web_helpers.AnalysisTimelineItem, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		item := buckets[key]
		result = append(result, web_helpers.AnalysisTimelineItem{
			Label:   item.label,
			Income:  round2(item.income),
			Expense: round2(item.expense),
		})
	}
	return result
}

func parseAnalysisRange(rawPeriod string, now time.Time) analysisRange {
	period := strings.ToLower(strings.TrimSpace(rawPeriod))
	if period != "quarter" && period != "year" {
		period = "month"
	}

	location := now.Location()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	label := fmt.Sprintf("%s %d", monthName(start.Month()), start.Year())

	switch period {
	case "quarter":
		quarterStartMonth := time.Month(((int(now.Month())-1)/3)*3 + 1)
		start = time.Date(now.Year(), quarterStartMonth, 1, 0, 0, 0, 0, location)
		end = start.AddDate(0, 3, 0).Add(-time.Nanosecond)
		quarterNumber := ((int(now.Month()) - 1) / 3) + 1
		label = fmt.Sprintf("%d квартал %d", quarterNumber, now.Year())
	case "year":
		start = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, location)
		end = time.Date(now.Year(), time.December, 31, 23, 59, 59, int(time.Second-time.Nanosecond), location)
		label = fmt.Sprintf("%d год", now.Year())
	}

	return analysisRange{
		period: period,
		label:  label,
		start:  start,
		end:    end,
	}
}

func calculateBudgetActual(transactions []analysisTransaction, categories []string, start time.Time, end time.Time) float64 {
	if len(categories) == 0 {
		return 0
	}
	categorySet := make(map[string]struct{}, len(categories))
	for _, category := range categories {
		categorySet[category] = struct{}{}
	}

	actual := 0.0
	for _, item := range transactions {
		transactionDate := item.transaction.TransactionDate
		if transactionDate.Before(start) || transactionDate.After(end) {
			continue
		}
		if _, ok := categorySet[item.transaction.Category]; !ok {
			continue
		}
		if strings.EqualFold(item.transaction.Type, "INCOME") {
			actual -= item.rubValue
		} else {
			actual += item.rubValue
		}
	}
	return round2(math.Max(actual, 0))
}

func budgetOverlapsRange(budget models.BudgetModel, start time.Time, end time.Time) bool {
	budgetEnd := budget.EndAt
	if budgetEnd.IsZero() {
		budgetEnd = end
	}
	return !budget.StartAt.After(end) && !budgetEnd.Before(start)
}

func getBudgetOverlap(budget models.BudgetModel, start time.Time, end time.Time) (time.Time, time.Time) {
	overlapStart := maxTime(start, budget.StartAt)
	overlapEnd := end
	if !budget.EndAt.IsZero() {
		overlapEnd = minTime(end, budget.EndAt)
	}
	return overlapStart, overlapEnd
}

func startOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, -(weekday - 1))
}

func monthName(month time.Month) string {
	months := map[time.Month]string{
		time.January:   "Январь",
		time.February:  "Февраль",
		time.March:     "Март",
		time.April:     "Апрель",
		time.May:       "Май",
		time.June:      "Июнь",
		time.July:      "Июль",
		time.August:    "Август",
		time.September: "Сентябрь",
		time.October:   "Октябрь",
		time.November:  "Ноябрь",
		time.December:  "Декабрь",
	}
	return months[month]
}

func shortMonth(month time.Month) string {
	months := map[time.Month]string{
		time.January:   "янв",
		time.February:  "фев",
		time.March:     "мар",
		time.April:     "апр",
		time.May:       "май",
		time.June:      "июн",
		time.July:      "июл",
		time.August:    "авг",
		time.September: "сен",
		time.October:   "окт",
		time.November:  "ноя",
		time.December:  "дек",
	}
	return months[month]
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func minTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxTime(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
