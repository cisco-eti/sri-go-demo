package idpadapter

import (
	"context"
	"encoding/json"
	"time"

	oidc "github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// AccessTokenClaims holds all of the claims that a standard JWT would have.
// But depending on the idp, the access token might not be a JWT.
type AccessTokenClaims struct {
	// standard
	Audience  string `json:"aud"`
	ExpiresAt int64  `json:"exp"`
	JTI       string `json:"jti"`
	IssuedAt  int64  `json:"iat"`
	Issuer    string `json:"iss"`
	NotBefore int64  `json:"nbf"`
	Subject   string `json:"sub"`

	// used by okta. maybe others?
	UID      string `json:"uid"`
	ClientID string `json:"cid"`
}

// VerifyJWTAccessToken will take an AccessToken formatted as a JWT and verify
// the signature, expiration, and the claims.
// https://auth0.com/docs/tokens/access-tokens/validate-access-tokens
// https://auth0.com/docs/tokens/json-web-tokens/validate-json-web-tokens
func (ia *IdentityProviderAdapter) VerifyJWTAccessToken(ctx context.Context,
	jwt string) (*AccessTokenClaims, error) {

	ctx = oidc.ClientContext(ctx, ia.httpClient)

	// uses a cached key set provided by the idp's jwks url
	payload, err := ia.jwksAccessTokenKeySet.VerifySignature(ctx, jwt)
	if err != nil {
		return nil, errors.Errorf("verifying jwt signature: %s", err)
	}

	var accessClaims AccessTokenClaims
	if err := json.Unmarshal(payload, &accessClaims); err != nil {
		return nil, errors.Errorf("unmarshaling claims: %v", err)
	}

	unixNow := time.Now().UTC().Unix()
	if accessClaims.IssuedAt > unixNow {
		return nil, errors.New("invalid jwt iat")
	}

	if accessClaims.ExpiresAt < unixNow {
		return nil, errors.New("invalid jwt expired")
	}

	if accessClaims.NotBefore > unixNow {
		return nil, errors.New("invalid jwt nbf")
	}

	if accessClaims.Issuer != ia.Issuer {
		return nil, errors.New("invalid jwt iss")
	}

	if accessClaims.Audience != ia.Audience {
		return nil, errors.New("invalid jwt aud")
	}

	if accessClaims.ClientID != "" && accessClaims.ClientID != ia.ClientID {
		return nil, errors.New("invalid jwt cid")
	}

	return &accessClaims, nil
}

// IDTokenClaims holds all of the claims that a standard JWT would have for an
// OIDC ID Token.
// https://openid.net/specs/openid-connect-basic-1_0-22.html#id_token
type IDTokenClaims struct {
	Audience  string `json:"aud"`
	ExpiresAt int64  `json:"exp"`
	JTI       string `json:"jti"`
	IssuedAt  int64  `json:"iat"`
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

// VerifyIDToken will take an ID Token and verify the signature, expiration,
// and the claims.
func (ia *IdentityProviderAdapter) VerifyIDToken(ctx context.Context,
	token *oauth2.Token) (*IDTokenClaims, string, error) {

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, "", errors.New("no id_token field in oauth2 token")
	}

	ctx = oidc.ClientContext(ctx, ia.httpClient)
	idToken, err := ia.oidcProvider.Verifier(ia.oidcConfig).Verify(ctx,
		rawIDToken)
	if err != nil {
		return nil, "", errors.Errorf("verifying jwt signature: %s", err)
	}

	idClaims := IDTokenClaims{}
	if err := idToken.Claims(&idClaims); err != nil {
		return nil, "", errors.Errorf("unmarshaling claims: %v", err)
	}

	unixNow := time.Now().UTC().Unix()
	if idClaims.IssuedAt > unixNow {
		return nil, "", errors.New("invalid jwt iat")
	}

	if idClaims.ExpiresAt < unixNow {
		return nil, "", errors.New("invalid jwt exp")
	}

	if idClaims.Issuer != ia.Issuer {
		return nil, "", errors.New("invalid jwt iss")
	}

	if idClaims.Audience != ia.ClientID {
		return nil, "", errors.New("invalid jwt aud")
	}

	return &idClaims, rawIDToken, nil
}
