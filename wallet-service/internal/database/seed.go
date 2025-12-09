package database

import (
	"log"
	"time"

	"github.com/jeffleon2/draftea-wallet-service/internal/models"
	"gorm.io/gorm"
)

func SeedWallets(db *gorm.DB) error {
	wallets := []models.Wallet{
		{
			ID:        "w1",
			UserID:    "user_1",
			Balance:   10000.0,
			Email:     "alice@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "w2",
			UserID:    "user_2",
			Balance:   5000.0,
			Email:     "bob@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "w3",
			UserID:    "user_3",
			Balance:   2000.0,
			Email:     "carol@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, wallet := range wallets {
		result := db.Where(models.Wallet{ID: wallet.ID}).FirstOrCreate(&wallet)
		if result.Error != nil {
			return result.Error
		}
	}

	log.Println("✅ Wallets seeded successfully")
	return nil
}
