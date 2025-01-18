package gsheets

import (
	"context"
	"encoding/json"
	"log"

	"golang.org/x/oauth2/jwt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Credenciales struct {
	Type                        string `json:type`
	Project_id                  string `json:project_id`
	Private_key_id              string `json:private_key_id`
	Private_key                 string `json:private_key`
	Client_email                string `json:client_email`
	Client_id                   string `json:client_id`
	Auth_uri                    string `json:auth_uri`
	Token_uri                   string `json:token_uri`
	Auth_provider_x509_cert_url string `json:auth_provider_x509_cert_url`
	Client_x509_cert_url        string `json:client_x509_cert_url`
	Universe_domain             string `json:universe_domain`
}

func (c Credenciales) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		panic(69) //<HACER/>
	}

	return string(b)
}

func ObtenerServicioGSheets(c Credenciales, dominio []string, correo string) (*sheets.Service, error) {
	// Your credentials should be obtained from the Google
	// Developer Console (https://console.developers.google.com).

	ctx := context.Background()

	configuracion := &jwt.Config{
		Email:      c.Client_email,
		PrivateKey: []byte(c.Private_key),
		Scopes:     dominio,
		TokenURL:   c.Token_uri,
		Subject:    correo,
	}

	cliente := configuracion.Client(ctx)

	servicio, err := sheets.NewService(ctx, option.WithHTTPClient(cliente))
	if err != nil {
		log.Fatalf("No se pudi instancia el cliente de sheets: %v", err)
		return nil, err
	}
	return servicio, nil
}
