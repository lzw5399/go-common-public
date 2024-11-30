package fcontext

import "context"

type ignoreTenantKey struct{}

func IgnoreTenantFromContext(ctx context.Context) bool {
	ignoreTenant, ok := ctx.Value(ignoreTenantKey{}).(bool)
	if !ok {
		return false
	}
	return ignoreTenant
}

func IgnoreTenantWithContext(ctx context.Context, ignoreTenant bool) context.Context {
	return context.WithValue(ctx, ignoreTenantKey{}, ignoreTenant)
}
