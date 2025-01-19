package gsheets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"reflect"
	"time"

	"golang.org/x/oauth2/jwt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Servicio = sheets.Service
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

func EsperarYLeer(servicio *sheets.Service, idhoja, rango string, intentosMaximos int, retardoBase time.Duration) ([][]any, error) {
	//intentosMaximos := 6
	//retardoBase := time.Second * 1 / 2

	var ultimosValores [][]interface{}

	for intento := 0; intento < intentosMaximos; intento++ {
		resp, err := servicio.Spreadsheets.Values.Get(idhoja, rango).
			ValueRenderOption("FORMATTED_VALUE").Do()
		if err != nil {
			return nil, err
		}

		valoresActuales := resp.Values

		switch {
		case ultimosValores != nil && reflect.DeepEqual(ultimosValores, valoresActuales):
			return valoresActuales, nil
		}

		ultimosValores = valoresActuales

		retardo := retardoBase * time.Duration(math.Pow(2, float64(intento)))
		time.Sleep(retardo)
	}

	return nil, fmt.Errorf("la hoja de calculo no se estabilizó después de %d intentos", intentosMaximos)
}

func EsperarYLeerCondicion(servicio *sheets.Service, idhoja, rango string, intentosMaximos int, retardoBase time.Duration, celda [2]int, condicion any) ([][]any, error) {
	//intentosMaximos := 6
	//retardoBase := time.Second * 1 / 2

	for intento := 0; intento < intentosMaximos; intento++ {
		resp, err := servicio.Spreadsheets.Values.Get(idhoja, rango).
			ValueRenderOption("FORMATTED_VALUE").Do()
		if err != nil {
			return nil, err
		}

		valoresActuales := resp.Values

		switch {
		case !reflect.DeepEqual(valoresActuales[celda[0]][celda[1]], condicion):
			return valoresActuales, nil
		}

		retardo := retardoBase * time.Duration(math.Pow(2, float64(intento)))
		time.Sleep(retardo)
	}

	return nil, fmt.Errorf("la hoja de calculo no se estabilizó después de %d intentos", intentosMaximos)
}

func ConvertirValoresGSaCSV(valores [][]any) ([][]string, error) {
	var csvSalida [][]string

	for f, fila := range valores {
		var estaFila []string
		for c, celda := range fila {
			cs, ok := celda.(string)
			if !ok {
				return csvSalida, fmt.Errorf("NO SE PUDO COERCIONAR UN VALOR A CSV: %d:%d | %v", f, c, celda)
			}
			estaFila = append(estaFila, cs)
		}
		csvSalida = append(csvSalida, estaFila)
	}

	return csvSalida, nil
}
