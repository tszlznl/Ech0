package migration

import (
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/gorm"
)

type oauthBindingsDropMigrator struct{}

func NewOAuthBindingsDropMigrator() Migrator {
	return &oauthBindingsDropMigrator{}
}

func (m *oauthBindingsDropMigrator) Name() string {
	return "oauth_bindings_drop_migrator"
}

func (m *oauthBindingsDropMigrator) Key() string {
	return commonModel.OAuthBindingsDroppedKey
}

func (m *oauthBindingsDropMigrator) CanRerun() bool {
	return false
}

func (m *oauthBindingsDropMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.Exec("DROP TABLE IF EXISTS oauth_bindings").Error
}
