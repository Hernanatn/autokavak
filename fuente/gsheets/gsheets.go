package gsheets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
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
		panic(69)
	}
	return string(b)
}

type Celda struct {
	Valor interface{}
}

func nuevaCelda(v interface{}) Celda {
	return Celda{Valor: v}
}

func (c Celda) Cadena() string {
	if c.Valor == nil {
		return ""
	}
	if s, ok := c.Valor.(string); ok {
		return s
	}
	return fmt.Sprint(c.Valor)
}

func (c Celda) Flotante() (float64, error) {
	if c.Valor == nil {
		return 0, fmt.Errorf("valor nulo")
	}

	str := c.Cadena()
	str = strings.Replace(str, "%", "", -1)
	str = strings.Replace(str, ",", ".", -1)

	return strconv.ParseFloat(str, 64)
}

func (c Celda) FlotanteODefault(def float64) float64 {
	val, err := c.Flotante()
	if err != nil {
		return def
	}
	return val
}

func (c Celda) Porcentaje() (float64, error) {
	val, err := c.Flotante()
	if err != nil {
		return 0, err
	}
	return val / 100, nil
}

type Fila []Celda

func nuevaFila(filaRaw []interface{}) Fila {
	fila := make(Fila, len(filaRaw))
	for i, v := range filaRaw {
		fila[i] = nuevaCelda(v)
	}
	return fila
}

func (f Fila) Cadenas() []string {

	var cadenas []string
	for _, celda := range f {
		cadenas = append(cadenas, celda.Cadena())
	}
	return cadenas
}

type TablaValores []Fila

func NuevaTabla(valores [][]interface{}) TablaValores {
	tabla := make(TablaValores, len(valores))
	for i, fila := range valores {
		tabla[i] = nuevaFila(fila)
	}
	return tabla
}

func (f Fila) obtenerCelda(col int) Celda {
	if col >= len(f) {
		return Celda{Valor: nil}
	}
	return f[col]
}

func (t TablaValores) Anys() [][]any {
	var matriz [][]any
	for _, f := range t {
		var fila []any
		for _, c := range f {
			fila = append(fila, c.Valor)
		}
		matriz = append(matriz, fila)
	}

	return matriz
}

func ObtenerServicioGSheets(c Credenciales, dominio []string, correo string) (*sheets.Service, error) {
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
		log.Fatalf("No se pudo instanciar el cliente de sheets: %v", err)
		return nil, err
	}
	return servicio, nil
}

func EsperarYLeer(servicio *sheets.Service, idhoja, rango string, intentosMaximos int, retardoBase time.Duration) (TablaValores, error) {
	var ultimosValores [][]interface{}

	for intento := 0; intento < intentosMaximos; intento++ {
		resp, err := servicio.Spreadsheets.Values.Get(idhoja, rango).
			ValueRenderOption("FORMATTED_VALUE").Do()
		if err != nil {
			return nil, err
		}

		valoresActuales := resp.Values

		if ultimosValores != nil && reflect.DeepEqual(ultimosValores, valoresActuales) {
			return NuevaTabla(valoresActuales), nil
		}

		ultimosValores = valoresActuales

		retardo := retardoBase * time.Duration(math.Pow(2, float64(intento)))
		time.Sleep(retardo)
	}

	return nil, fmt.Errorf("la hoja de calculo no se estabilizó después de %d intentos", intentosMaximos)
}

func EsperarYLeerCondicion(servicio *sheets.Service, idhoja, rango string, intentosMaximos int, retardoBase time.Duration, celda [2]int, condicion any) (TablaValores, error) {
	for intento := 0; intento < intentosMaximos; intento++ {
		resp, err := servicio.Spreadsheets.Values.Get(idhoja, rango).
			ValueRenderOption("FORMATTED_VALUE").Do()
		if err != nil {
			return nil, err
		}

		valoresActuales := resp.Values

		if !reflect.DeepEqual(valoresActuales[celda[0]][celda[1]], condicion) {
			return NuevaTabla(valoresActuales), nil
		}

		retardo := retardoBase * time.Duration(math.Pow(2, float64(intento)))
		time.Sleep(retardo)
	}

	return nil, fmt.Errorf("la hoja de calculo no se estabilizó después de %d intentos", intentosMaximos)
}

func ConvertirValoresGSaCSV(valores TablaValores) ([][]string, error) {
	var csvSalida [][]string

	for f, fila := range valores {
		var estaFila []string
		for c, celda := range fila {
			cs := celda.Cadena()
			if cs == "" {
				return csvSalida, fmt.Errorf("NO SE PUDO COERCIONAR UN VALOR A CSV: %d:%d | %v", f, c, celda.Valor)
			}
			estaFila = append(estaFila, cs)
		}
		csvSalida = append(csvSalida, estaFila)
	}

	return csvSalida, nil
}
