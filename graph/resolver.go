package graph

import "github.com/gaubeur/desafio-posgraduacao-golang-fullcycle/desafio-clean-architecture/internal/database"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	OrderDB *database.Order
}
