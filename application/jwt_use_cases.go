package application

import "context"

type JWTUseCaseInterface interface {
	Create(ctx context.Context)
	Check(ctx context.Context)
	Delete(ctx context.Context)
}

type JWTUseCase struct {
}
