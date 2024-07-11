package models

import "github.com/csams/common-inventory/pkg/authn/api"

type Input interface {
	Validate() []error
}

type Model interface {
	GetId() IDType
	GetResourceType() string
}

type Output interface {
	SetHref(string)
}

type Transformer[I Input, M Model, O Output] interface {
	NewInput() I
	NewModel() M
	Create(I, *api.Identity) M
	Update(I, M, *api.Identity)
	ToOutput(M) O
}
