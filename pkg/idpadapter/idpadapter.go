package idpadapter

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	oidc "github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/oauth2"

	etilogger "wwwin-github.cisco.com/eti/sre-go-logger"
)

// AuthFlowType is used to identify if the authentication flow
type AuthFlowType int

const (
	// LoginAuthFlow is used to identify if the auth flow is for an existing user
	// "log in"
	LoginAuthFlow AuthFlowType = iota

	// SignupAuthFlow is used to identify if the auth flow is for a new user
	// "sign up"
	SignupAuthFlow
)

// IdentityProviderAdapter is the configuration for an OAuth2 with OIDC
// Identity Provider
type IdentityProviderAdapter struct {
	log *etilogger.Logger

	label                    string
	ClientID                 string
	clientSecret             string
	Issuer                   string
	Audience                 string
	issuerURL                url.URL
	defaultLoginCallbackURL  url.URL
	defaultSignupCallbackURL url.URL
	issuerLogoutPath         string

	httpClient            *http.Client
	oidcProvider          *oidc.Provider
	oidcConfig            *oidc.Config
	jwksAccessTokenKeySet oidc.KeySet
	oauth2Config          *oauth2.Config
}

// New constructs a new IdentityProviderAdapter. It is expected that this
// struct will be long lived as it will make HTTP requests to the 3rd party
// IDP to cache the JWKS keyset. It will make requests using the provided
// http.Client
func New(ctx context.Context, log *etilogger.Logger, httpClient *http.Client,
	label, id, secret, issuer, audience, loginCB, signupCB, logout string) (
	*IdentityProviderAdapter, error) {

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	ctx = oidc.ClientContext(ctx, httpClient)
	oidcProv, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, errors.Errorf("oidc provider: %s", err)
	}

	oidcConfig := &oidc.Config{
		ClientID:             id,
		SupportedSigningAlgs: []string{oidc.RS256},
	}

	keyset, err := newRemoteKeySet(ctx, oidcProv)
	if err != nil {
		return nil, errors.Errorf("jwks keyset: %s", err)
	}

	oauthConf := &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		Endpoint:     oidcProv.Endpoint(),
		RedirectURL:  "", // intentionally left blank here
		// TODO(saolds): include a refresh token too?
		Scopes: []string{oidc.ScopeOpenID, "email", "profile"},
	}

	loginCBURL, err := url.Parse(loginCB)
	if err != nil {
		return nil, errors.Errorf("login callback expected to be valid url: %s",
			err)
	}

	signupCBURL, err := url.Parse(signupCB)
	if err != nil {
		return nil, errors.Errorf("signup callback expected to be valid url: %s",
			err)
	}

	issuerURL, err := url.Parse(issuer)
	if err != nil {
		return nil, errors.Errorf("issuer expected to be valid url: %s", err)
	}

	i := &IdentityProviderAdapter{
		log:                      log,
		label:                    label,
		ClientID:                 id,
		clientSecret:             secret,
		Issuer:                   issuer,
		Audience:                 audience,
		issuerURL:                *issuerURL,
		defaultLoginCallbackURL:  *loginCBURL,
		defaultSignupCallbackURL: *signupCBURL,
		issuerLogoutPath:         logout,
		httpClient:               httpClient,
		oidcProvider:             oidcProv,
		oidcConfig:               oidcConfig,
		jwksAccessTokenKeySet:    keyset,
		oauth2Config:             oauthConf,
	}

	if i.log != nil {
		i.log.Debug("initialized new identity provider adapter for %s: %s %s",
			i.label, i.Issuer, i.Audience)
	}

	return i, nil
}

// AuthCodeURL will create the link to the identity provider's consent page
// used when getting the code
func (ia *IdentityProviderAdapter) AuthCodeURL(state,
	overrideRedirectURI string, authFlow AuthFlowType) string {

	nonce := uuid.NewV4().String()
	oauth2Client := ia.initOAuth2(overrideRedirectURI, authFlow)
	return oauth2Client.AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce))
}

// Identity holds all of the information from a parsed ID Token from an
// identity provider that should be able to uniquely identify a user
type Identity struct {
	UserID               string // idtoken.sub
	Name                 string // idtoken.name
	Email                string // idtoken.email
	Issuer               string // idtoken.iss
	IDToken              string // raw idtoken
	AccessToken          string // raw access token
	AccessTokenIssuedAt  time.Time
	AccessTokenExpiresAt time.Time
}

// ExchangeCodeAndVerifyTokens will exchange the code for id and access tokens
// overrideRedirectURI is only necessary if the provided defaults callbacks
// were not used when getting the code
func (ia *IdentityProviderAdapter) ExchangeCodeAndVerifyTokens(
	ctx context.Context, code, overrideRedirectURI string,
	authFlow AuthFlowType) (*Identity, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, ia.httpClient)
	oauth2Client := ia.initOAuth2(overrideRedirectURI, authFlow)
	token, err := oauth2Client.Exchange(ctx, code)
	if err != nil {
		return nil, errors.Errorf("exchanging code for token: %s", err)
	}

	return ia.getIdentity(ctx, token)
}

// ExchangeUsernameAndPasswordAndVerifyTokens will do a Resource Owner Password
// Grant Type flow. It is not a recommended approach and should be used as a
// last resort, or in a testing environment.
// It is intended for applications for which no other flow works, as it
// requires your application code to be fully trusted and protected from
// credential-stealing attacks. It is made available primarily to provide a
// consistent and predictable integration pattern for legacy applications that
// can't otherwise be updated to a more secure flow such as the Authorization
// Code flow
func (ia *IdentityProviderAdapter) ExchangeUsernameAndPasswordAndVerifyTokens(
	ctx context.Context, username, password string) (*Identity, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, ia.httpClient)
	oauth2Client := ia.initOAuth2("", LoginAuthFlow)
	token, err := oauth2Client.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, errors.Errorf("exchanging u:p for token: %s", err)
	}

	return ia.getIdentity(ctx, token)
}

// getIdentity will verify the idtoken and return an Identity
func (ia *IdentityProviderAdapter) getIdentity(ctx context.Context,
	token *oauth2.Token) (*Identity, error) {
	idClaims, rawIDToken, err := ia.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, errors.Errorf("verifying id token: %s", err)
	}

	i := &Identity{
		UserID:               idClaims.Subject,
		Name:                 idClaims.Name,
		Email:                idClaims.Email,
		Issuer:               idClaims.Issuer,
		IDToken:              rawIDToken,
		AccessToken:          token.AccessToken,
		AccessTokenIssuedAt:  time.Unix(idClaims.IssuedAt, 0).UTC(),
		AccessTokenExpiresAt: token.Expiry.UTC(),
	}

	return i, nil
}

// LogoutLink will generate the logout url for the 3rd party identity provider
func (ia *IdentityProviderAdapter) LogoutLink(params url.Values) string {
	logoutURL := ia.issuerURL
	logoutURL.Path += "/" + strings.TrimLeft(ia.issuerLogoutPath, "/")
	logoutURL.RawQuery = params.Encode()
	return logoutURL.String()
}

func newRemoteKeySet(ctx context.Context, provider *oidc.Provider) (
	oidc.KeySet, error) {

	var claims struct {
		JWKSURL string `json:"jwks_uri"`
	}

	// for some reason, coreos didn't export the RemoteKeysSet or the JWKSURL
	// used when verifying access tokens
	if err := provider.Claims(&claims); err != nil {
		return nil, err
	}

	return oidc.NewRemoteKeySet(ctx, claims.JWKSURL), nil
}

func (ia *IdentityProviderAdapter) initOAuth2(redirectURI string,
	authFlow AuthFlowType) *oauth2.Config {

	// make a copy to not modify the redirect uri
	cp := &oauth2.Config{}
	*cp = *ia.oauth2Config

	// the provided redirectURI is an override for the default callbacks
	if redirectURI == "" {
		switch authFlow {
		case LoginAuthFlow:
			redirectURI = ia.defaultLoginCallbackURL.String()
		case SignupAuthFlow:
			redirectURI = ia.defaultSignupCallbackURL.String()
		}
	}

	cp.RedirectURL = redirectURI
	return cp
}
