package store

import (
	"time"

	"multi-cluster-dashboard/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MetricsStore handles database operations for metrics and alerts
type MetricsStore struct {
	db *gorm.DB
}

// NewMetricsStore creates a new metrics store with SQLite
func NewMetricsStore(dbPath string) (*MetricsStore, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&models.MetricSnapshot{}, &models.Alert{}); err != nil {
		return nil, err
	}

	return &MetricsStore{db: db}, nil
}

// SaveSnapshot saves a metrics snapshot
func (s *MetricsStore) SaveSnapshot(snapshot *models.MetricSnapshot) error {
	snapshot.Timestamp = time.Now()
	return s.db.Create(snapshot).Error
}

// GetSnapshots retrieves snapshots for a cluster within a time range
func (s *MetricsStore) GetSnapshots(cluster string, since time.Duration) ([]models.MetricSnapshot, error) {
	var snapshots []models.MetricSnapshot
	cutoff := time.Now().Add(-since)
	
	err := s.db.Where("cluster = ? AND timestamp > ?", cluster, cutoff).
		Order("timestamp ASC").
		Find(&snapshots).Error
	
	return snapshots, err
}

// GetLatestSnapshots retrieves the most recent n snapshots for a cluster
func (s *MetricsStore) GetLatestSnapshots(cluster string, limit int) ([]models.MetricSnapshot, error) {
	var snapshots []models.MetricSnapshot
	
	err := s.db.Where("cluster = ?", cluster).
		Order("timestamp DESC").
		Limit(limit).
		Find(&snapshots).Error
	
	return snapshots, err
}

// CleanupOldSnapshots removes snapshots older than the specified duration
func (s *MetricsStore) CleanupOldSnapshots(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return s.db.Where("timestamp < ?", cutoff).Delete(&models.MetricSnapshot{}).Error
}

// SaveAlert saves a new alert
func (s *MetricsStore) SaveAlert(alert *models.Alert) error {
	alert.Timestamp = time.Now()
	return s.db.Create(alert).Error
}

// GetActiveAlerts retrieves all unresolved alerts
func (s *MetricsStore) GetActiveAlerts() ([]models.Alert, error) {
	var alerts []models.Alert
	err := s.db.Where("resolved = ?", false).
		Order("timestamp DESC").
		Find(&alerts).Error
	return alerts, err
}

// GetAlertsByCluster retrieves alerts for a specific cluster
func (s *MetricsStore) GetAlertsByCluster(cluster string) ([]models.Alert, error) {
	var alerts []models.Alert
	err := s.db.Where("cluster = ?", cluster).
		Order("timestamp DESC").
		Find(&alerts).Error
	return alerts, err
}

// ResolveAlert marks an alert as resolved
func (s *MetricsStore) ResolveAlert(id uint) error {
	return s.db.Model(&models.Alert{}).Where("id = ?", id).Update("resolved", true).Error
}

// GetRecentAlerts retrieves alerts from the last n hours
func (s *MetricsStore) GetRecentAlerts(hours int) ([]models.Alert, error) {
	var alerts []models.Alert
	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	err := s.db.Where("timestamp > ?", cutoff).
		Order("timestamp DESC").
		Find(&alerts).Error
	
	return alerts, err
}

// Close closes the database connection
func (s *MetricsStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
