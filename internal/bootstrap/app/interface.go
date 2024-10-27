package app

import "context"

type DependencyChecker interface {
	Check(ctx context.Context) error
}
