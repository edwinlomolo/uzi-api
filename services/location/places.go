package location

import "github.com/3dw1nM0535/uzi-api/model"

type Places interface {
	AutocompletePlace(query string) ([]*model.Place, error)
}
