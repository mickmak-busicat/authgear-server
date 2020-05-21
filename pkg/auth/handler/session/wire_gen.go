// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package session

import (
	"github.com/skygeario/skygear-server/pkg/auth"
	auth2 "github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	redis2 "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/pq"
	redis3 "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	pq2 "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
	"net/http"
)

// Injectors from wire.go:

func newResolveHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	insecureCookieConfig := auth.ProvideSessionInsecureCookieConfig(m)
	context := auth.ProvideContext(r)
	tenantConfiguration := auth.ProvideTenantConfig(context, m)
	cookieConfiguration := session.ProvideSessionCookieConfiguration(r, insecureCookieConfig, tenantConfiguration)
	provider := time.NewProvider()
	requestID := auth.ProvideLoggingRequestID(r)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	store := redis.ProvideStore(context, tenantConfiguration, provider, factory)
	eventStore := redis2.ProvideEventStore(context, tenantConfiguration)
	accessEventProvider := &auth2.AccessEventProvider{
		Store: eventStore,
	}
	sessionProvider := session.ProvideSessionProvider(r, store, accessEventProvider, tenantConfiguration)
	resolver := &session.Resolver{
		CookieConfiguration: cookieConfiguration,
		Provider:            sessionProvider,
		Time:                provider,
	}
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	authorizationStore := &pq.AuthorizationStore{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	grantStore := redis3.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	resolverSessionProvider := oauth.ProvideResolverProvider(sessionProvider)
	oauthResolver := &oauth.Resolver{
		Authorizations: authorizationStore,
		AccessGrants:   grantStore,
		OfflineGrants:  grantStore,
		Sessions:       resolverSessionProvider,
		Time:           provider,
	}
	authAccessEventProvider := auth2.AccessEventProvider{
		Store: eventStore,
	}
	authinfoStore := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	middleware := &auth2.Middleware{
		IDPSessionResolver:         resolver,
		AccessTokenSessionResolver: oauthResolver,
		AccessEvents:               authAccessEventProvider,
		AuthInfoStore:              authinfoStore,
		Time:                       provider,
		TxContext:                  txContext,
	}
	anonymousProvider := anonymous.ProvideProvider(sqlBuilder, sqlExecutor)
	handler := provideResolveHandler(middleware, factory, provider, anonymousProvider)
	return handler
}

func newListHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	tenantConfiguration := auth.ProvideTenantConfig(context, m)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	store := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	provider := time.NewProvider()
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	userprofileStore := userprofile.ProvideStore(provider, sqlBuilder, sqlExecutor)
	requestID := auth.ProvideLoggingRequestID(r)
	reservedNameChecker := auth.ProvideReservedNameChecker(m)
	typeCheckerFactory := loginid.ProvideTypeCheckerFactory(tenantConfiguration, reservedNameChecker)
	checker := loginid.ProvideChecker(tenantConfiguration, typeCheckerFactory)
	normalizerFactory := loginid.ProvideNormalizerFactory(tenantConfiguration)
	loginidProvider := loginid.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration, checker, normalizerFactory)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	hookProvider := hook.ProvideHookProvider(context, sqlBuilder, sqlExecutor, requestID, tenantConfiguration, txContext, provider, store, userprofileStore, loginidProvider, factory)
	sessionStore := redis.ProvideStore(context, tenantConfiguration, provider, factory)
	insecureCookieConfig := auth.ProvideSessionInsecureCookieConfig(m)
	cookieConfiguration := session.ProvideSessionCookieConfiguration(r, insecureCookieConfig, tenantConfiguration)
	manager := session.ProvideSessionManager(sessionStore, provider, tenantConfiguration, cookieConfiguration)
	grantStore := redis3.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	sessionManager := &oauth.SessionManager{
		Store: grantStore,
		Time:  provider,
	}
	authSessionManager := &auth2.SessionManager{
		AuthInfoStore:       store,
		UserProfileStore:    userprofileStore,
		Hooks:               hookProvider,
		IDPSessions:         manager,
		AccessTokenSessions: sessionManager,
	}
	requireAuthz := handler.NewRequireAuthzFactory(factory)
	httpHandler := provideListHandler(txContext, authSessionManager, requireAuthz)
	return httpHandler
}

func newGetHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	tenantConfiguration := auth.ProvideTenantConfig(context, m)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	store := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	provider := time.NewProvider()
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	userprofileStore := userprofile.ProvideStore(provider, sqlBuilder, sqlExecutor)
	requestID := auth.ProvideLoggingRequestID(r)
	reservedNameChecker := auth.ProvideReservedNameChecker(m)
	typeCheckerFactory := loginid.ProvideTypeCheckerFactory(tenantConfiguration, reservedNameChecker)
	checker := loginid.ProvideChecker(tenantConfiguration, typeCheckerFactory)
	normalizerFactory := loginid.ProvideNormalizerFactory(tenantConfiguration)
	loginidProvider := loginid.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration, checker, normalizerFactory)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	hookProvider := hook.ProvideHookProvider(context, sqlBuilder, sqlExecutor, requestID, tenantConfiguration, txContext, provider, store, userprofileStore, loginidProvider, factory)
	sessionStore := redis.ProvideStore(context, tenantConfiguration, provider, factory)
	insecureCookieConfig := auth.ProvideSessionInsecureCookieConfig(m)
	cookieConfiguration := session.ProvideSessionCookieConfiguration(r, insecureCookieConfig, tenantConfiguration)
	manager := session.ProvideSessionManager(sessionStore, provider, tenantConfiguration, cookieConfiguration)
	grantStore := redis3.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	sessionManager := &oauth.SessionManager{
		Store: grantStore,
		Time:  provider,
	}
	authSessionManager := &auth2.SessionManager{
		AuthInfoStore:       store,
		UserProfileStore:    userprofileStore,
		Hooks:               hookProvider,
		IDPSessions:         manager,
		AccessTokenSessions: sessionManager,
	}
	validator := auth.ProvideValidator(m)
	requireAuthz := handler.NewRequireAuthzFactory(factory)
	httpHandler := provideGetHandler(txContext, authSessionManager, validator, requireAuthz)
	return httpHandler
}

func newRevokeHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	tenantConfiguration := auth.ProvideTenantConfig(context, m)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	store := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	provider := time.NewProvider()
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	userprofileStore := userprofile.ProvideStore(provider, sqlBuilder, sqlExecutor)
	requestID := auth.ProvideLoggingRequestID(r)
	reservedNameChecker := auth.ProvideReservedNameChecker(m)
	typeCheckerFactory := loginid.ProvideTypeCheckerFactory(tenantConfiguration, reservedNameChecker)
	checker := loginid.ProvideChecker(tenantConfiguration, typeCheckerFactory)
	normalizerFactory := loginid.ProvideNormalizerFactory(tenantConfiguration)
	loginidProvider := loginid.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration, checker, normalizerFactory)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	hookProvider := hook.ProvideHookProvider(context, sqlBuilder, sqlExecutor, requestID, tenantConfiguration, txContext, provider, store, userprofileStore, loginidProvider, factory)
	sessionStore := redis.ProvideStore(context, tenantConfiguration, provider, factory)
	insecureCookieConfig := auth.ProvideSessionInsecureCookieConfig(m)
	cookieConfiguration := session.ProvideSessionCookieConfiguration(r, insecureCookieConfig, tenantConfiguration)
	manager := session.ProvideSessionManager(sessionStore, provider, tenantConfiguration, cookieConfiguration)
	grantStore := redis3.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	sessionManager := &oauth.SessionManager{
		Store: grantStore,
		Time:  provider,
	}
	authSessionManager := &auth2.SessionManager{
		AuthInfoStore:       store,
		UserProfileStore:    userprofileStore,
		Hooks:               hookProvider,
		IDPSessions:         manager,
		AccessTokenSessions: sessionManager,
	}
	validator := auth.ProvideValidator(m)
	requireAuthz := handler.NewRequireAuthzFactory(factory)
	httpHandler := provideRevokeHandler(txContext, authSessionManager, validator, requireAuthz)
	return httpHandler
}

func newRevokeAllHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	tenantConfiguration := auth.ProvideTenantConfig(context, m)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	store := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	provider := time.NewProvider()
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	userprofileStore := userprofile.ProvideStore(provider, sqlBuilder, sqlExecutor)
	requestID := auth.ProvideLoggingRequestID(r)
	reservedNameChecker := auth.ProvideReservedNameChecker(m)
	typeCheckerFactory := loginid.ProvideTypeCheckerFactory(tenantConfiguration, reservedNameChecker)
	checker := loginid.ProvideChecker(tenantConfiguration, typeCheckerFactory)
	normalizerFactory := loginid.ProvideNormalizerFactory(tenantConfiguration)
	loginidProvider := loginid.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration, checker, normalizerFactory)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	hookProvider := hook.ProvideHookProvider(context, sqlBuilder, sqlExecutor, requestID, tenantConfiguration, txContext, provider, store, userprofileStore, loginidProvider, factory)
	sessionStore := redis.ProvideStore(context, tenantConfiguration, provider, factory)
	insecureCookieConfig := auth.ProvideSessionInsecureCookieConfig(m)
	cookieConfiguration := session.ProvideSessionCookieConfiguration(r, insecureCookieConfig, tenantConfiguration)
	manager := session.ProvideSessionManager(sessionStore, provider, tenantConfiguration, cookieConfiguration)
	grantStore := redis3.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	sessionManager := &oauth.SessionManager{
		Store: grantStore,
		Time:  provider,
	}
	authSessionManager := &auth2.SessionManager{
		AuthInfoStore:       store,
		UserProfileStore:    userprofileStore,
		Hooks:               hookProvider,
		IDPSessions:         manager,
		AccessTokenSessions: sessionManager,
	}
	requireAuthz := handler.NewRequireAuthzFactory(factory)
	httpHandler := provideRevokeAllHandler(txContext, authSessionManager, requireAuthz)
	return httpHandler
}

// wire.go:

func provideResolveHandler(
	m *auth2.Middleware,
	lf logging.Factory,
	t time.Provider,
	ap *anonymous.Provider,
) http.Handler {
	return m.Handle(&ResolveHandler{
		TimeProvider:  t,
		LoggerFactory: lf,
		Anonymous:     ap,
	})
}

func provideListHandler(tx db.TxContext, sm sessionListManager, requireAuthz handler.RequireAuthz) http.Handler {
	h := &ListHandler{
		txContext:      tx,
		sessionManager: sm,
	}
	return requireAuthz(h, h)
}

func provideGetHandler(tx db.TxContext, sm sessionGetManager, v *validation.Validator, requireAuthz handler.RequireAuthz) http.Handler {
	h := &GetHandler{
		validator:      v,
		txContext:      tx,
		sessionManager: sm,
	}
	return requireAuthz(h, h)
}

func provideRevokeHandler(tx db.TxContext, sm sessionRevokeManager, v *validation.Validator, requireAuthz handler.RequireAuthz) http.Handler {
	h := &RevokeHandler{
		validator:      v,
		txContext:      tx,
		sessionManager: sm,
	}
	return requireAuthz(h, h)
}

func provideRevokeAllHandler(tx db.TxContext, sm sessionRevokeAllManager, requireAuthz handler.RequireAuthz) http.Handler {
	h := &RevokeAllHandler{
		txContext:      tx,
		sessionManager: sm,
	}
	return requireAuthz(h, h)
}
