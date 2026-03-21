package idempotency

// Key represents an idempotency key
type Key struct {
	Key    string
	Result interface{}
}

func (k *Key) Store() error {
	// Store in database with TTL
	return nil
}

func (k *Key) Get(key string) (interface{}, error) {
	// Retrieve from database
	return nil, nil
}
