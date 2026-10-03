package model

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"rainbow-backend/internal/config"
)

func OpenDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&Admin{},
		&User{},
		&SceneDomain{},
		&ContentItem{},
		&ScenePageConfig{},
	); err != nil {
		return err
	}

	if err := migrateSceneDomains(db); err != nil {
		return err
	}
	if err := migrateContentSceneCode(db); err != nil {
		return err
	}
	if err := migrateScenePageConfigs(db); err != nil {
		return err
	}
	if err := migrateUsers(db); err != nil {
		return err
	}

	return nil
}

func migrateSceneDomains(db *gorm.DB) error {
	if err := db.Exec(
		"UPDATE scene_domains SET host = LOWER(TRIM(host)) WHERE host IS NOT NULL AND host <> ''",
	).Error; err != nil {
		return fmt.Errorf("normalize scene domain host: %w", err)
	}

	if err := db.Exec(
		"UPDATE scene_domains SET scene_code = LOWER(TRIM(scene_code)) WHERE scene_code IS NOT NULL AND scene_code <> ''",
	).Error; err != nil {
		return fmt.Errorf("normalize scene domain scene_code: %w", err)
	}

	return nil
}

func migrateContentSceneCode(db *gorm.DB) error {
	if err := addContentItemColumnIfMissing(db, "image_urls", "tags", "JSON NULL COMMENT '图片链接列表'"); err != nil {
		return err
	}

	if err := db.Exec(
		"UPDATE content_items SET scene_code = ? WHERE scene_code IS NULL OR scene_code = ''",
		"default",
	).Error; err != nil {
		return fmt.Errorf("backfill content scene_code: %w", err)
	}

	indexes, err := findLegacyContentDateIndexes(db)
	if err != nil {
		return err
	}
	for _, indexName := range indexes {
		if err := db.Migrator().DropIndex(&ContentItem{}, indexName); err != nil {
			return fmt.Errorf("drop legacy content index %s: %w", indexName, err)
		}
	}

	if err := db.Exec(
		"ALTER TABLE content_items MODIFY COLUMN scene_code VARCHAR(64) NOT NULL DEFAULT 'default'",
	).Error; err != nil {
		return fmt.Errorf("enforce content scene_code constraint: %w", err)
	}
	if err := db.Exec(
		"UPDATE content_items SET image_urls = JSON_ARRAY() WHERE image_urls IS NULL",
	).Error; err != nil {
		return fmt.Errorf("backfill content image_urls: %w", err)
	}

	return nil
}

func addContentItemColumnIfMissing(db *gorm.DB, columnName string, afterColumn string, columnDef string) error {
	if db.Migrator().HasColumn(&ContentItem{}, columnName) {
		return nil
	}

	query := fmt.Sprintf("ALTER TABLE content_items ADD COLUMN %s %s", columnName, columnDef)
	if afterColumn != "" {
		query += fmt.Sprintf(" AFTER %s", afterColumn)
	}

	if err := db.Exec(query).Error; err != nil {
		return fmt.Errorf("add content item %s: %w", columnName, err)
	}

	return nil
}

func migrateScenePageConfigs(db *gorm.DB) error {
	if err := db.Exec(
		"UPDATE scene_page_configs SET scene_code = LOWER(TRIM(scene_code)) WHERE scene_code IS NOT NULL AND scene_code <> ''",
	).Error; err != nil {
		return fmt.Errorf("normalize scene page config scene_code: %w", err)
	}

	if err := addScenePageConfigColumnIfMissing(db, &ScenePageConfig{}, "default_bg_url", "bac_img", "VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '默认背景图URL'"); err != nil {
		return err
	}
	if err := addScenePageConfigColumnIfMissing(db, &ScenePageConfig{}, "default_music", "default_bg_url", "VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '默认背景音乐URL'"); err != nil {
		return err
	}
	if err := addScenePageConfigColumnIfMissing(db, &ScenePageConfig{}, "image_urls", "tags_default", "JSON NULL COMMENT '图片链接列表'"); err != nil {
		return err
	}
	if err := addScenePageConfigColumnIfMissing(db, &ScenePageConfig{}, "video_url", "image_urls", "VARCHAR(1024) NULL COMMENT '视频URL'"); err != nil {
		return err
	}
	if err := addScenePageConfigColumnIfMissing(db, &ScenePageConfig{}, "avatar_url", "video_url", "VARCHAR(1024) NULL COMMENT '头像URL'"); err != nil {
		return err
	}

	if err := db.Exec(
		"UPDATE scene_page_configs SET tags_default = JSON_ARRAY() WHERE tags_default IS NULL",
	).Error; err != nil {
		return fmt.Errorf("backfill scene page config tags_default: %w", err)
	}
	if err := db.Exec(
		"UPDATE scene_page_configs SET default_bg_url = '' WHERE default_bg_url IS NULL",
	).Error; err != nil {
		return fmt.Errorf("backfill scene page config default_bg_url: %w", err)
	}
	if err := db.Exec(
		"UPDATE scene_page_configs SET default_music = '' WHERE default_music IS NULL",
	).Error; err != nil {
		return fmt.Errorf("backfill scene page config default_music: %w", err)
	}
	if err := db.Exec(
		"UPDATE scene_page_configs SET image_urls = JSON_ARRAY() WHERE image_urls IS NULL",
	).Error; err != nil {
		return fmt.Errorf("backfill scene page config image_urls: %w", err)
	}
	if err := db.Exec(
		"UPDATE scene_page_configs SET video_url = '' WHERE video_url IS NULL",
	).Error; err != nil {
		return fmt.Errorf("backfill scene page config video_url: %w", err)
	}
	if err := db.Exec(
		"UPDATE scene_page_configs SET avatar_url = '' WHERE avatar_url IS NULL",
	).Error; err != nil {
		return fmt.Errorf("backfill scene page config avatar_url: %w", err)
	}

	if err := db.Exec(
		"ALTER TABLE scene_page_configs MODIFY COLUMN scene_code VARCHAR(128) NOT NULL",
	).Error; err != nil {
		return fmt.Errorf("enforce scene page config scene_code constraint: %w", err)
	}

	return nil
}

func migrateUsers(db *gorm.DB) error {
	columns := []struct {
		name        string
		afterColumn string
		definition  string
	}{
		{
			name:        "scene_code",
			afterColumn: "last_login_at",
			definition:  "JSON NULL COMMENT '用户场景编码列表'",
		},
		{
			name:        "nickname",
			afterColumn: "scene_code",
			definition:  "VARCHAR(128) NULL COMMENT '用户昵称'",
		},
		{
			name:        "avatar_url",
			afterColumn: "nickname",
			definition:  "VARCHAR(1024) NULL COMMENT '用户头像URL'",
		},
		{
			name:        "birthday",
			afterColumn: "avatar_url",
			definition:  "DATE NULL COMMENT '出生年月'",
		},
		{
			name:        "gender",
			afterColumn: "birthday",
			definition:  "INT NULL COMMENT '性别'",
		},
		{
			name:        "occupation",
			afterColumn: "gender",
			definition:  "INT NULL COMMENT '职业'",
		},
	}

	for _, column := range columns {
		if err := addUserColumnIfMissing(db, column.name, column.afterColumn, column.definition); err != nil {
			return err
		}
	}

	if err := db.Exec(
		"UPDATE users SET scene_code = JSON_ARRAY() WHERE scene_code IS NULL",
	).Error; err != nil {
		return fmt.Errorf("backfill users scene_code: %w", err)
	}

	return nil
}

func addUserColumnIfMissing(db *gorm.DB, columnName, afterColumn, columnDefinition string) error {
	if db.Migrator().HasColumn(&User{}, columnName) {
		return nil
	}

	query := fmt.Sprintf("ALTER TABLE users ADD COLUMN %s %s", columnName, columnDefinition)
	if afterColumn != "" {
		query += fmt.Sprintf(" AFTER %s", afterColumn)
	}

	if err := db.Exec(query).Error; err != nil {
		return fmt.Errorf("add user %s: %w", columnName, err)
	}

	return nil
}

func addScenePageConfigColumnIfMissing(db *gorm.DB, model any, columnName string, afterColumn string, columnDef string) error {
	if db.Migrator().HasColumn(model, columnName) {
		return nil
	}

	query := fmt.Sprintf("ALTER TABLE scene_page_configs ADD COLUMN %s %s", columnName, columnDef)
	if afterColumn != "" {
		query += fmt.Sprintf(" AFTER %s", afterColumn)
	}

	if err := db.Exec(query).Error; err != nil {
		return fmt.Errorf("add scene page config %s: %w", columnName, err)
	}

	return nil
}

func findLegacyContentDateIndexes(db *gorm.DB) ([]string, error) {
	type row struct {
		IndexName string
	}

	var rows []row
	err := db.Raw(`
		SELECT INDEX_NAME
		FROM information_schema.statistics
		WHERE table_schema = DATABASE()
		  AND table_name = ?
		  AND non_unique = 0
		GROUP BY INDEX_NAME
		HAVING COUNT(*) = 1
		   AND SUM(CASE WHEN column_name = 'date' THEN 1 ELSE 0 END) = 1
		   AND INDEX_NAME <> ?
	`, (&ContentItem{}).TableName(), "idx_content_scene_date").Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("query legacy content indexes: %w", err)
	}

	result := make([]string, 0, len(rows))
	for _, item := range rows {
		if item.IndexName == "" || item.IndexName == "PRIMARY" {
			continue
		}
		result = append(result, item.IndexName)
	}

	return result, nil
}

func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
