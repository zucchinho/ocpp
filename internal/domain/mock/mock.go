package mock

import (
	_ "github.com/golang/mock/mockgen/model"
)

//go:generate mockgen -destination=mock.gen.go -package=mock github.com/zucchinho/ocpp/internal/domain EventSource
