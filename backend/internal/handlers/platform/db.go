package platform

import (
	"os"

	"github.com/Shravanthi20/InDel/backend/internal/services"
	"github.com/Shravanthi20/InDel/backend/pkg/razorpay"
	"gorm.io/gorm"
)

var platformDB *gorm.DB
var platformCoreOps *services.CoreOpsService

// SetDB registers DB handle and initialized core service for platform handlers.
func SetDB(db *gorm.DB) {
	platformDB = db
	
	// Ensure core ops service is initialized with Razorpay client
	platformCoreOps = services.NewCoreOpsService(db)
	
	apiKey := os.Getenv("RAZORPAY_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("Test_Key_ID")
	}
	apiSecret := os.Getenv("RAZORPAY_API_SECRET")
	if apiSecret == "" {
		apiSecret = os.Getenv("Test_Key_Secret")
	}
	rzpClient := razorpay.NewRazorpayClient(apiKey, apiSecret)
	platformCoreOps.SetRazorpayClient(rzpClient)
}

func hasDB() bool {
	return platformDB != nil
}
