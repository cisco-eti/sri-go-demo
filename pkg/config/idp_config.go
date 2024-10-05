package config

import "os"

type IDPConfig struct {
	Label            string
	ClientID         string
	ClientSecret     string
	Issuer           string
	Audience         string
	LoginCallback    string
	SignupCallback   string
	IssuerLogoutPath string
}

func ReadIDPConfig() (IDPConfig, error) {
	i := IDPConfig{
		Label:            lookupenv("IDP_LABEL"),
		ClientID:         lookupenv("IDP_CLIENT_ID"),
		ClientSecret:     lookupenv("IDP_CLIENT_SECRET"),
		Issuer:           lookupenv("IDP_ISSUER"),
		Audience:         lookupenv("IDP_AUDIENCE"),
		LoginCallback:    lookupenv("IDP_LOGIN_CALLBACK"),
		SignupCallback:   lookupenv("IDP_SIGNUP_CALLBACK"),
		IssuerLogoutPath: lookupenv("IDP_ISSUER_LOGOUT_PATH"),
	}

	return i, nil
}

func lookupenv(envVar string) string {
	e, exist := os.LookupEnv(envVar)
	if !exist || e == "" {
		return ""
	}
	return e
}
