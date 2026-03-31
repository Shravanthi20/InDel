package insurer

import "github.com/Shravanthi20/InDel/backend/internal/services"

// InsurerHandler encapsulates all endpoint methods and its dependencies
type InsurerHandler struct {
	Service *services.InsurerService
}

func NewInsurerHandler(svc *services.InsurerService) *InsurerHandler {
	return &InsurerHandler{Service: svc}
}
