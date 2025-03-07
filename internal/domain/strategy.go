package domain

import (
	"github.com/cfioretti/calculator/pkg/domain"
)

type Strategy func(data map[string]string) domain.Pan
