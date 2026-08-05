package service

import "context"

// GrowthRegistrationSessionMaxBytes bounds the attribution cookie retained in
// request context and persisted by the registration recorder.
const GrowthRegistrationSessionMaxBytes = 64

type growthRegistrationSessionContextKey struct{}

// WithGrowthRegistrationSession stores a bounded attribution session value in
// the request context. Invalid values are intentionally omitted so they can be
// represented as JSON null by the recorder.
func WithGrowthRegistrationSession(ctx context.Context, session string) context.Context {
	if ctx == nil || session == "" || len([]byte(session)) > GrowthRegistrationSessionMaxBytes {
		return ctx
	}
	return context.WithValue(ctx, growthRegistrationSessionContextKey{}, session)
}

// GrowthRegistrationSessionFromContext returns the captured attribution
// session, if one was present and valid.
func GrowthRegistrationSessionFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	session, ok := ctx.Value(growthRegistrationSessionContextKey{}).(string)
	if !ok || session == "" || len([]byte(session)) > GrowthRegistrationSessionMaxBytes {
		return "", false
	}
	return session, true
}
