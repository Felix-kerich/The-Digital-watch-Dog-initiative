package repository

import (
	"strconv"
	"time"

	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/utils"
	"gorm.io/gorm"
)

type analyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository creates a new analytics repository instance
func NewAnalyticsRepository() AnalyticsRepository {
	return &analyticsRepository{
		db: utils.DB,
	}
}

func (r *analyticsRepository) GetTransactionSummary(filter map[string]interface{}) (map[string]interface{}, error) {
	query := r.db.Model(&models.Transaction{})

	// Apply filters
	if entityID, ok := filter["entityId"].(string); ok && entityID != "" {
		query = query.Where("entity_id = ?", entityID)
	}
	if year, ok := filter["year"].(string); ok && year != "" {
		if yearInt, err := strconv.Atoi(year); err == nil {
			startDate := time.Date(yearInt, time.January, 1, 0, 0, 0, 0, time.UTC)
			endDate := time.Date(yearInt+1, time.January, 1, 0, 0, 0, 0, time.UTC)
			query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate)
		}
	}
	if transactionType, ok := filter["type"].(string); ok && transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}

	// Get transaction counts by status
	var statusResults []struct {
		Status string
		Count  int
		Total  float64
	}
	query.Select("status, COUNT(*) as count, SUM(amount) as total").
		Group("status").
		Scan(&statusResults)

	// Get total count of transactions by type
	var typeResults []struct {
		Type  string
		Count int
		Total float64
	}
	query.Select("type, COUNT(*) as count, SUM(amount) as total").
		Group("type").
		Scan(&typeResults)

	// Get monthly trends
	var monthlyResults []struct {
		Month  time.Time
		Count  int
		Total  float64
		Status string
	}
	monthQuery := r.db.Model(&models.Transaction{})
	if entityID, ok := filter["entityId"].(string); ok && entityID != "" {
		monthQuery = monthQuery.Where("entity_id = ?", entityID)
	}
	if transactionType, ok := filter["type"].(string); ok && transactionType != "" {
		monthQuery = monthQuery.Where("type = ?", transactionType)
	}

	monthQuery.Raw(`
		SELECT
			DATE_FORMAT(created_at, '%Y-%m-01') as month,
			COUNT(*) as count,
			SUM(amount) as total,
			status
		FROM
			transactions
		WHERE
			created_at >= DATE_SUB(NOW(), INTERVAL 12 MONTH)
		GROUP BY
			DATE_FORMAT(created_at, '%Y-%m'),
			status
		ORDER BY
			month
	`).Scan(&monthlyResults)

	// Get flagged transactions count
	var flaggedCount int64
	flaggedQuery := r.db.Model(&models.Transaction{}).Where("ai_flagged = ?", true)
	if entityID, ok := filter["entityId"].(string); ok && entityID != "" {
		flaggedQuery = flaggedQuery.Where("entity_id = ?", entityID)
	}
	if year, ok := filter["year"].(string); ok && year != "" {
		if yearInt, err := strconv.Atoi(year); err == nil {
			startDate := time.Date(yearInt, time.January, 1, 0, 0, 0, 0, time.UTC)
			endDate := time.Date(yearInt+1, time.January, 1, 0, 0, 0, 0, time.UTC)
			flaggedQuery = flaggedQuery.Where("created_at BETWEEN ? AND ?", startDate, endDate)
		}
	}
	flaggedQuery.Count(&flaggedCount)

	return map[string]interface{}{
		"summary":       statusResults,
		"byType":        typeResults,
		"monthlyTrends": monthlyResults,
		"flaggedCount":  flaggedCount,
	}, nil
}

func (r *analyticsRepository) GetUserActivitySummary(filter map[string]interface{}) (map[string]interface{}, error) {
	query := r.db.Model(&models.AuditLog{})

	// Apply filters
	if entityID, ok := filter["entityId"].(string); ok && entityID != "" {
		query = query.Where("entity_id = ?", entityID)
	}
	if userID, ok := filter["userId"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if fromDate, ok := filter["fromDate"].(string); ok && fromDate != "" {
		from, err := time.Parse("2006-01-02", fromDate)
		if err == nil {
			query = query.Where("timestamp >= ?", from)
		}
	}
	if toDate, ok := filter["toDate"].(string); ok && toDate != "" {
		to, err := time.Parse("2006-01-02", toDate)
		if err == nil {
			to = to.Add(24 * time.Hour)
			query = query.Where("timestamp < ?", to)
		}
	}

	// Get activity counts by action
	var actionResults []struct {
		Action string
		Count  int
	}
	query.Select("action, COUNT(*) as count").
		Group("action").
		Scan(&actionResults)

	// Get activity counts by user
	var userResults []struct {
		UserID    string
		UserName  string
		UserEmail string
		Count     int
	}
	r.db.Raw(`
		SELECT
			u.id as user_id,
			u.name as user_name,
			u.email as user_email,
			COUNT(a.id) as count
		FROM
			audit_logs a
			JOIN users u ON a.user_id = u.id
		GROUP BY
			a.user_id
		ORDER BY
			count DESC
		LIMIT 10
	`).Scan(&userResults)

	// Get recent activity
	var recentActivity []models.AuditLog
	query.Order("timestamp DESC").Limit(20).Find(&recentActivity)

	return map[string]interface{}{
		"byAction":       actionResults,
		"byUser":         userResults,
		"recentActivity": recentActivity,
	}, nil
}

func (r *analyticsRepository) GetFundUtilizationReport(fundID, entityID, fiscalYear string) (map[string]interface{}, error) {
	fundsQuery := r.db.Model(&models.Fund{})

	if fundID != "" {
		fundsQuery = fundsQuery.Where("id = ?", fundID)
	}
	if entityID != "" {
		fundsQuery = fundsQuery.Where("entity_id = ?", entityID)
	}
	if fiscalYear != "" {
		if fy, err := strconv.Atoi(fiscalYear); err == nil {
			fundsQuery = fundsQuery.Where("fiscal_year = ?", fy)
		}
	}

	var funds []struct {
		ID          string
		Name        string
		Amount      float64
		FiscalYear  int
		EntityID    string
		EntityName  string
		Utilized    float64
		UtilizedPct float64
	}

	fundsQuery.Raw(`
		SELECT
			f.id,
			f.name,
			f.amount,
			f.fiscal_year,
			f.entity_id,
			e.name as entity_name,
			COALESCE(SUM(t.amount), 0) as utilized,
			CASE WHEN f.amount > 0 THEN (COALESCE(SUM(t.amount), 0) / f.amount) * 100 ELSE 0 END as utilized_pct
		FROM
			funds f
			LEFT JOIN entities e ON f.entity_id = e.id
			LEFT JOIN transactions t ON f.id = t.fund_id AND t.status = 'COMPLETED'
		WHERE
			f.status = 'ACTIVE'
			AND (f.id = ? OR ? IS NULL)
			AND (f.entity_id = ? OR ? IS NULL)
			AND (f.fiscal_year = ? OR ? IS NULL)
		GROUP BY
			f.id
		ORDER BY
			utilized_pct DESC
	`, fundID, fundID, entityID, entityID, fiscalYear, fiscalYear).Scan(&funds)

	response := map[string]interface{}{
		"funds": funds,
	}

	// Get detailed breakdown by transaction type if a single fund is specified
	if fundID != "" {
		var typeBreakdown []struct {
			Type       string
			Count      int
			Total      float64
			Percentage float64
		}

		r.db.Raw(`
			SELECT
				type,
				COUNT(*) as count,
				SUM(amount) as total,
				CASE WHEN (SELECT amount FROM funds WHERE id = ?) > 0 
					THEN (SUM(amount) / (SELECT amount FROM funds WHERE id = ?)) * 100 
					ELSE 0 
				END as percentage
			FROM
				transactions
			WHERE
				fund_id = ?
				AND status = 'COMPLETED'
			GROUP BY
				type
		`, fundID, fundID, fundID).Scan(&typeBreakdown)

		response["typeBreakdown"] = typeBreakdown

		// Get monthly trends for the fund
		var monthlyTrends []struct {
			Month      time.Time
			Amount     float64
			Type       string
			Cumulative float64
		}

		r.db.Raw(`
			SELECT
				DATE_FORMAT(created_at, '%Y-%m-01') as month,
				SUM(amount) as amount,
				type,
				@running_total := @running_total + SUM(amount) as cumulative
			FROM
				transactions
			CROSS JOIN (SELECT @running_total := 0) rt
			WHERE
				fund_id = ?
				AND status = 'COMPLETED'
			GROUP BY
				DATE_FORMAT(created_at, '%Y-%m'),
				type
			ORDER BY
				month
		`, fundID).Scan(&monthlyTrends)

		response["monthlyTrends"] = monthlyTrends
	}

	return response, nil
}
