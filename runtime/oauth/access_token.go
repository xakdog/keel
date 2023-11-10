package oauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/samber/lo"
	"github.com/segmentio/ksuid"
	"github.com/teamkeel/keel/runtime/common"
	"github.com/teamkeel/keel/runtime/runtimectx"
)

const (
	KeelIssuer                             = "https://keel.so"
	DefaultAccessTokenExpiry time.Duration = time.Hour * 24
)

var (
	ErrInvalidToken     = common.NewAuthenticationFailedMessageErr("cannot be parsed or verified as a valid JWT")
	ErrTokenExpired     = common.NewAuthenticationFailedMessageErr("token has expired")
	ErrIdentityNotFound = common.NewAuthenticationFailedMessageErr("identity not found")
)

// https://pkg.go.dev/github.com/golang-jwt/jwt/v4#RegisteredClaims
type AccessTokenClaims struct {
	jwt.RegisteredClaims
}

func GenerateAccessToken(ctx context.Context, identityId string) (string, time.Duration, error) {
	expiry := DefaultAccessTokenExpiry
	config, err := runtimectx.GetAuthConfig(ctx)
	if err == nil {
		if config != nil && config.Keel != nil {
			expiry = time.Duration(config.Keel.TokenDuration) * time.Second
		}
	}

	token, err := generateToken(ctx, identityId, []string{}, expiry)
	if err != nil {
		return "", 0, err
	}

	return token, expiry, nil
}

func generateToken(ctx context.Context, sub string, aud []string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			Audience:  aud,
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    KeelIssuer,
		},
	}

	privateKey, err := runtimectx.GetPrivateKey(ctx)
	if err != nil {
		return "", err
	}

	if privateKey == nil {
		return "", fmt.Errorf("no private key set")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("cannot create signed jwt: %w", err)
	}
	return tokenString, nil
}

func ValidateAccessToken(ctx context.Context, tokenString string, audienceClaim string) (string, string, error) {
	ctx, span := tracer.Start(ctx, "Validate access token")
	defer span.End()

	privateKey, err := runtimectx.GetPrivateKey(ctx)
	if err != nil {
		return "", "", err
	}

	if privateKey == nil {
		return "", "", errors.New("no private key set")
	}

	var token *jwt.Token
	claims := &AccessTokenClaims{}

	token, err = jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	var validationErr *jwt.ValidationError
	if errors.As(err, &validationErr) && validationErr.Errors == jwt.ValidationErrorExpired {
		return "", "", ErrTokenExpired
	}

	if err != nil {
		return "", "", ErrInvalidToken
	}

	if !claims.VerifyExpiresAt(time.Now().UTC(), true) {
		return "", "", ErrTokenExpired
	}

	if audienceClaim != "" {
		if !lo.Contains(claims.Audience, audienceClaim) {
			return "", "", ErrInvalidToken
		}
	}

	if err != nil || !token.Valid {
		return "", "", ErrInvalidToken
	}

	identifier, err := ksuid.Parse(claims.Subject)
	if err != nil {
		return "", "", err
	}

	return identifier.String(), claims.Issuer, nil
}
