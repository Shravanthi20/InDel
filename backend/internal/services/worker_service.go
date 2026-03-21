package services

// worker_service.go - Worker management service
type WorkerService struct{}

func (s *WorkerService) GetProfile(workerID uint) (interface{}, error) {
	return nil, nil
}

func (s *WorkerService) Onboard(phone string, profile map[string]interface{}) (uint, error) {
	return 0, nil
}
