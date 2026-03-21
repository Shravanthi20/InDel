package razorpay

// RazorpayClient handles Razorpay payments
type RazorpayClient struct{}

func (r *RazorpayClient) CreatePayout(workerID uint, amount float64, UPI string) (string, error) {
	return "", nil
}

func (r *RazorpayClient) CheckPayoutStatus(payoutID string) (string, error) {
	return "", nil
}
