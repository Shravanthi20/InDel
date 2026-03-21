package firebase

// OTPService handles Firebase OTP
type OTPService struct{}

func (o *OTPService) SendOTP(phone string) error {
	return nil
}

func (o *OTPService) VerifyOTP(phone string, otp string) (bool, error) {
	return false, nil
}
