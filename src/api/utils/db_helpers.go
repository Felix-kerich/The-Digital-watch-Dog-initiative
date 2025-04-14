package utils

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// IsUniqueConstraintError checks if an error is a MySQL unique constraint violation
func IsUniqueConstraintError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}

// WithTimeout wraps a database operation with a timeout
func WithTimeout(timeout time.Duration, operation func(ctx context.Context, tx *gorm.DB) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return operation(ctx, tx)
	})
}

// GenerateUUID generates a new UUID string
func GenerateUUID() string {
	return uuid.New().String()
}

// FindByID is a generic function to find a record by ID
func FindByID(id string, result interface{}) error {
	return DB.First(result, "id = ?", id).Error
}

// FindAll is a generic function to find all records with pagination
func FindAll(result interface{}, page, pageSize int, order string) error {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if order == "" {
		order = "created_at desc"
	}

	offset := (page - 1) * pageSize
	return DB.Order(order).Limit(pageSize).Offset(offset).Find(result).Error
}

// Create is a generic function to create a record
func Create(value interface{}) error {
	return DB.Create(value).Error
}

// Update is a generic function to update a record
func Update(value interface{}) error {
	return DB.Save(value).Error
}

// DeleteByID is a generic function to delete a record by ID
func DeleteByID(model interface{}, id string) error {
	return DB.Delete(model, "id = ?", id).Error
}

// Exists checks if a record exists based on a condition
func Exists(model interface{}, condition string, args ...interface{}) (bool, error) {
	var count int64
	err := DB.Model(model).Where(condition, args...).Count(&count).Error
	return count > 0, err
}

// Upsert performs an insert or update operation
func Upsert(value interface{}, uniqueColumns []string, updateColumns []string) error {
	// Convert string slices to clause.Column slices
	columns := make([]clause.Column, len(uniqueColumns))
	for i, col := range uniqueColumns {
		columns[i] = clause.Column{Name: col}
	}

	return DB.Clauses(clause.OnConflict{
		Columns:   columns,
		DoUpdates: clause.AssignmentColumns(updateColumns),
	}).Create(value).Error
}

// Count returns the count of records based on conditions
func Count(model interface{}, condition string, args ...interface{}) (int64, error) {
	var count int64
	var err error
	if condition == "" {
		err = DB.Model(model).Count(&count).Error
	} else {
		err = DB.Model(model).Where(condition, args...).Count(&count).Error
	}
	return count, err
}
