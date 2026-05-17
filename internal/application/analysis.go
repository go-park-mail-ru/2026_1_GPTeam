package application

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/repository"
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/currency_converter"
)

type AnalysisPeriod string

const (
	AnalysisPeriodMonth   AnalysisPeriod = "month"
	AnalysisPeriodQuarter AnalysisPeriod = "quarter"
	AnalysisPeriodYear    AnalysisPeriod = "year"
)

type AnalysisUseCase interface {
	Get(ctx context.Context, user models.UserModel, params AnalysisParams) (AnalysisResult, error)
}

type AnalysisParams struct {
	Period    AnalysisPeriod
	StartDate *time.Time
	Now       time.Time
}

type AnalysisRange struct {
	Period AnalysisPeriod
	Label  string
	Start  time.Time
	End    time.Time
}

type AnalysisSummary struct {
	TotalBudgetLimit float64
	TotalBudgetSpent float64
	TotalBudgetFree  float64
	IncomeTotal      float64
	ExpenseTotal     float64
	Savings          float64
}

type AnalysisBudgetItem struct {
	ID         int
	Title      string
	Categories []string
	Target     float64
	Actual     float64
	Remaining  float64
	Progress   float64
	Currency   string
}

type AnalysisCategoryItem struct {
	Category string
	Amount   float64
	Share    float64
}

type AnalysisTimelineItem struct {
	Label   string
	Income  float64
	Expense float64
}

type AnalysisResult struct {
	Period      AnalysisPeriod
	PeriodLabel string
	PeriodStart time.Time
	PeriodEnd   time.Time
	Summary     AnalysisSummary
	Budgets     []AnalysisBudgetItem
	Categories  []AnalysisCategoryItem
	Timeline    []AnalysisTimelineItem
}

type Analysis struct {
	budgetApp      BudgetUseCase
	transactionApp TransactionUseCase
	accountApp     AccountUseCase
}

type analysisTransaction struct {
	transaction models.TransactionModel
	rubValue    float64
}

func NewAnalysis(budgetApp BudgetUseCase, transactionApp TransactionUseCase, accountApp AccountUseCase) *Analysis {
	return &Analysis{
		budgetApp:      budgetApp,
		transactionApp: transactionApp,
		accountApp:     accountApp,
	}
}

func (obj *Analysis) Get(ctx context.Context, user models.UserModel, params AnalysisParams) (AnalysisResult, error) {
	rangeInfo := BuildAnalysisRange(params.Period, params.StartDate, params.Now)
	transactions, err := obj.getTransactionsInRange(ctx, user.Id, rangeInfo.Start, rangeInfo.End)
	if err != nil {
		return AnalysisResult{}, err
	}

	budgets, summary, err := obj.getBudgetAnalysis(ctx, user, transactions, rangeInfo)
	if err != nil {
		return AnalysisResult{}, err
	}

	categories, incomeTotal, expenseTotal := buildCategoryAnalysis(transactions)
	timeline := buildTimeline(rangeInfo, transactions)
	summary.IncomeTotal = roundAnalysis(incomeTotal)
	summary.ExpenseTotal = roundAnalysis(expenseTotal)
	summary.Savings = roundAnalysis(incomeTotal - expenseTotal)

	return AnalysisResult{
		Period:      rangeInfo.Period,
		PeriodLabel: rangeInfo.Label,
		PeriodStart: rangeInfo.Start,
		PeriodEnd:   rangeInfo.End,
		Summary:     summary,
		Budgets:     budgets,
		Categories:  categories,
		Timeline:    timeline,
	}, nil
}

func ParseAnalysisPeriod(rawPeriod string) AnalysisPeriod {
	switch AnalysisPeriod(strings.ToLower(strings.TrimSpace(rawPeriod))) {
	case AnalysisPeriodQuarter:
		return AnalysisPeriodQuarter
	case AnalysisPeriodYear:
		return AnalysisPeriodYear
	default:
		return AnalysisPeriodMonth
	}
}

func ParseAnalysisStartDate(rawStartDate string, location *time.Location) (*time.Time, error) {
	value := strings.TrimSpace(rawStartDate)
	if value == "" {
		return nil, nil
	}
	if location == nil {
		location = time.Local
	}
	parsed, err := time.ParseInLocation("2006-01-02", value, location)
	if err == nil {
		return &parsed, nil
	}
	parsed, err = time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	parsed = parsed.In(location)
	start := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, location)
	return &start, nil
}

func BuildAnalysisRange(period AnalysisPeriod, startDate *time.Time, now time.Time) AnalysisRange {
	if now.IsZero() {
		now = time.Now()
	}
	period = ParseAnalysisPeriod(string(period))
	location := now.Location()

	if startDate != nil {
		start := time.Date(startDate.In(location).Year(), startDate.In(location).Month(), startDate.In(location).Day(), 0, 0, 0, 0, location)
		end := endByPeriod(start, period)
		return AnalysisRange{
			Period: period,
			Label:  fmt.Sprintf("%s — %s", formatDateRu(start), formatDateRu(end)),
			Start:  start,
			End:    end,
		}
	}

	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	label := fmt.Sprintf("%s %d", monthNameAnalysis(start.Month()), start.Year())

	switch period {
	case AnalysisPeriodQuarter:
		quarterStartMonth := time.Month(((int(now.Month())-1)/3)*3 + 1)
		start = time.Date(now.Year(), quarterStartMonth, 1, 0, 0, 0, 0, location)
		end = start.AddDate(0, 3, 0).Add(-time.Nanosecond)
		quarterNumber := ((int(now.Month()) - 1) / 3) + 1
		label = fmt.Sprintf("%d квартал %d", quarterNumber, now.Year())
	case AnalysisPeriodYear:
		start = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, location)
		end = time.Date(now.Year(), time.December, 31, 23, 59, 59, int(time.Second-time.Nanosecond), location)
		label = fmt.Sprintf("%d год", now.Year())
	}

	return AnalysisRange{
		Period: period,
		Label:  label,
		Start:  start,
		End:    end,
	}
}

func IsBudgetOverlappingRange(budget models.BudgetModel, start time.Time, end time.Time) bool {
	budgetEnd := budget.EndAt
	if budgetEnd.IsZero() {
		budgetEnd = end
	}
	return !budget.StartAt.After(end) && !budgetEnd.Before(start)
}

func GetBudgetRangeIntersection(budget models.BudgetModel, start time.Time, end time.Time) (time.Time, time.Time) {
	overlapStart := maxAnalysisTime(start, budget.StartAt)
	overlapEnd := end
	if !budget.EndAt.IsZero() {
		overlapEnd = minAnalysisTime(end, budget.EndAt)
	}
	return overlapStart, overlapEnd
}

func (obj *Analysis) getTransactionsInRange(ctx context.Context, userId int, start time.Time, end time.Time) ([]analysisTransaction, error) {
	filters := repository.TransactionFilters{
		StartDate: &start,
		EndDate:   &end,
	}
	transactions, err := obj.transactionApp.Search(ctx, userId, filters)
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
			loadedCurrency, currencyErr := obj.accountApp.GetCurrencyByAccountId(ctx, transaction.AccountId)
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

func (obj *Analysis) getBudgetAnalysis(ctx context.Context, authUser models.UserModel, transactions []analysisTransaction, rangeInfo AnalysisRange) ([]AnalysisBudgetItem, AnalysisSummary, error) {
	budgetIDs, err := obj.budgetApp.GetBudgetsOfUser(ctx, authUser)
	if err != nil {
		if !errors.Is(err, repository.NothingInTableError) {
			return nil, AnalysisSummary{}, err
		}
		budgetIDs = []int{}
	}

	budgetItems := make([]AnalysisBudgetItem, 0, len(budgetIDs))
	summary := AnalysisSummary{}

	for _, budgetID := range budgetIDs {
		budget, categories, detailErr := obj.budgetApp.GetById(ctx, budgetID, authUser)
		if detailErr != nil {
			if errors.Is(detailErr, repository.NothingInTableError) || errors.Is(detailErr, UserNotAuthorOfBudgetError) {
				continue
			}
			return nil, AnalysisSummary{}, detailErr
		}
		if !IsBudgetOverlappingRange(budget, rangeInfo.Start, rangeInfo.End) {
			continue
		}

		overlapStart, overlapEnd := GetBudgetRangeIntersection(budget, rangeInfo.Start, rangeInfo.End)
		actual := calculateBudgetActual(transactions, categories, overlapStart, overlapEnd)
		remaining := math.Max(budget.Target-actual, 0)
		progress := 0.0
		if budget.Target > 0 {
			progress = math.Min(actual/budget.Target*100, 999)
		}

		budgetItems = append(budgetItems, AnalysisBudgetItem{
			ID:         budget.Id,
			Title:      strings.TrimSpace(budget.Title),
			Categories: categories,
			Target:     roundAnalysis(budget.Target),
			Actual:     roundAnalysis(actual),
			Remaining:  roundAnalysis(remaining),
			Progress:   roundAnalysis(progress),
			Currency:   budget.Currency,
		})
		summary.TotalBudgetLimit += budget.Target
		summary.TotalBudgetSpent += actual
	}

	summary.TotalBudgetLimit = roundAnalysis(summary.TotalBudgetLimit)
	summary.TotalBudgetSpent = roundAnalysis(summary.TotalBudgetSpent)
	summary.TotalBudgetFree = roundAnalysis(math.Max(summary.TotalBudgetLimit-summary.TotalBudgetSpent, 0))

	sort.Slice(budgetItems, func(i, j int) bool {
		if budgetItems[i].Target == budgetItems[j].Target {
			return budgetItems[i].Title < budgetItems[j].Title
		}
		return budgetItems[i].Target > budgetItems[j].Target
	})

	return budgetItems, summary, nil
}

func buildCategoryAnalysis(transactions []analysisTransaction) ([]AnalysisCategoryItem, float64, float64) {
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

	categories := make([]AnalysisCategoryItem, 0, len(categoryTotals))
	for category, amount := range categoryTotals {
		share := 0.0
		if expenseTotal > 0 {
			share = amount / expenseTotal * 100
		}
		categories = append(categories, AnalysisCategoryItem{
			Category: category,
			Amount:   roundAnalysis(amount),
			Share:    roundAnalysis(share),
		})
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Amount > categories[j].Amount
	})

	return categories, roundAnalysis(incomeTotal), roundAnalysis(expenseTotal)
}

func buildTimeline(rangeInfo AnalysisRange, transactions []analysisTransaction) []AnalysisTimelineItem {
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

	switch rangeInfo.Period {
	case AnalysisPeriodYear:
		cursor := time.Date(rangeInfo.Start.Year(), rangeInfo.Start.Month(), 1, 0, 0, 0, 0, rangeInfo.Start.Location())
		for !cursor.After(rangeInfo.End) {
			key := cursor.Format("2006-01")
			ensureBucket(key, shortMonthAnalysis(cursor.Month()))
			cursor = cursor.AddDate(0, 1, 0)
		}
	case AnalysisPeriodQuarter:
		cursor := rangeInfo.Start
		for !cursor.After(rangeInfo.End) {
			weekStart := startOfWeekAnalysis(cursor)
			key := weekStart.Format("2006-01-02")
			ensureBucket(key, fmt.Sprintf("%02d %s", weekStart.Day(), shortMonthAnalysis(weekStart.Month())))
			cursor = weekStart.AddDate(0, 0, 7)
		}
	default:
		cursor := rangeInfo.Start
		for !cursor.After(rangeInfo.End) {
			key := cursor.Format("2006-01-02")
			ensureBucket(key, fmt.Sprintf("%02d", cursor.Day()))
			cursor = cursor.AddDate(0, 0, 1)
		}
	}

	for _, item := range transactions {
		date := item.transaction.TransactionDate.In(rangeInfo.Start.Location())
		var key string
		switch rangeInfo.Period {
		case AnalysisPeriodYear:
			key = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location()).Format("2006-01")
		case AnalysisPeriodQuarter:
			key = startOfWeekAnalysis(date).Format("2006-01-02")
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

	result := make([]AnalysisTimelineItem, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		item := buckets[key]
		result = append(result, AnalysisTimelineItem{
			Label:   item.label,
			Income:  roundAnalysis(item.income),
			Expense: roundAnalysis(item.expense),
		})
	}
	return result
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
	return roundAnalysis(math.Max(actual, 0))
}

func endByPeriod(start time.Time, period AnalysisPeriod) time.Time {
	switch period {
	case AnalysisPeriodQuarter:
		return start.AddDate(0, 3, 0).Add(-time.Nanosecond)
	case AnalysisPeriodYear:
		return start.AddDate(1, 0, 0).Add(-time.Nanosecond)
	default:
		return start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	}
}

func startOfWeekAnalysis(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, -(weekday - 1))
}

func monthNameAnalysis(month time.Month) string {
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

func shortMonthAnalysis(month time.Month) string {
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

func formatDateRu(value time.Time) string {
	return fmt.Sprintf("%02d.%02d.%d", value.Day(), value.Month(), value.Year())
}

func roundAnalysis(value float64) float64 {
	return math.Round(value*100) / 100
}

func minAnalysisTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxAnalysisTime(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
