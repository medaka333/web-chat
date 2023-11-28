package initializers

import "web-chat/models"

func SyncDatabase() {
	// DBの初期化
	DB.AutoMigrate(&models.Users{}, &models.Rooms{}, &models.Friends{}, &models.Groups{}, &models.Chat_history{})
}
