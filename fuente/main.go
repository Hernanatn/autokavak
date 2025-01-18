package main

import (
	"encoding/json"
	"fmt"
	"kavak/data"
	"kavak/gsheets"
	"os"
	"strings"

	"github.com/hernanatn/aplicacion.go"
	"github.com/hernanatn/aplicacion.go/consola"
	"github.com/hernanatn/aplicacion.go/consola/cadena"
	"google.golang.org/api/sheets/v4"
)

var programa aplicacion.Aplicacion
var servicio *sheets.Service

func inicializar(a aplicacion.Aplicacion, args ...string) error {
	var err error
	var CREDENCIALES gsheets.Credenciales
	err = json.Unmarshal(data.CREDENCIALES, &CREDENCIALES)
	if err != nil {
		a.ImprimirFatal("No se pudo leer las credenciales", err)
		return err
	}

	var dominios []string = []string{
		"https://www.googleapis.com/auth/spreadsheets",
		"https://www.googleapis.com/auth/drive",
	}
	servicio, err = gsheets.ObtenerServicioGSheets(CREDENCIALES, dominios, data.CORREO)
	if err != nil {
		a.ImprimirFatal("No se pudo crear el servicio de Google Sheets", err)
		return err
	}

	//var opciones googleapi.CallOption

	return nil
}

func finalizar(a aplicacion.Aplicacion, args ...string) error {
	a.ImprimirLinea(aplicacion.Cadena("¡Adiós!"))
	return nil
}

func limpiar(a aplicacion.Aplicacion, args ...string) error {
	return nil
}

func init() {
	programa = aplicacion.NuevaAplicacion(
		"Kavak",
		"kavak <opciones> [comando]",
		"aplicación que reemplaza el trabajo de mike en kavak.",
		make([]string, 0),
		aplicacion.NuevaConsola(os.Stdin, os.Stdout),
	).
		RegistrarInicio(inicializar).
		RegistrarLimpieza(finalizar).
		RegistrarFinal(limpiar)
}

func main() {

	var ajustarGM aplicacion.Comando = aplicacion.NuevoComando(
		"ajustar",
		"ajustar",
		[]string{},
		"Primer tarea. Ajustar GM.",
		aplicacion.Accion(
			func(con aplicacion.Consola, opciones aplicacion.Opciones, parametros aplicacion.Parametros, argumentos ...any) (res any, cod aplicacion.CodigoError, err error) {
				var autos []string
				con.ImprimirLinea(aplicacion.Cadena("ajustar"))

				const idHoja string = data.ID_HOJA
				const rango string = "A1:D2"
				respuesta, err := servicio.Spreadsheets.Values.Get(idHoja, rango).Do()
				if err != nil {
					con.ImprimirFatal(consola.Cadena(fmt.Sprintf("No se pudo leer el rango solicitado. Id hoja: %s | Rango: %s ", idHoja, rango)), nil)
					return nil, aplicacion.ERROR, err
				}

				if len(respuesta.Values) < 1 {
					con.ImprimirAdvertencia(consola.Cadena(fmt.Sprintf("No se encontró data en el rango solicitado. Id hoja: %s | Rango: %s ", idHoja, rango)), nil)
					return nil, aplicacion.ERROR, err
				}

				for _, fila := range respuesta.Values {
					con.ImprimirLinea(consola.Cadena(fmt.Sprintf("%s", fila)))
				}

				if len(autos) < 10 {
					continuar, err := con.Leer("Hay menos de diez autos nuevos.\n¿Continuar? [S/n]\n")
					if err != nil {
						e := con.ImprimirLinea(aplicacion.Cadena(cadena.Fatal("Autos. 14/31.", err)))
						if e != nil {
							return nil, aplicacion.ERROR, e
						}
					}

					if strings.ToLower(continuar.S()) != "s" {
						e := con.ImprimirLinea("Ejecución abortada por el Usuario")
						if e != nil {
							return nil, aplicacion.ERROR, e
						}
					}
				}
				return nil, aplicacion.EXITO, nil
			}),
		make([]string, 0),
	)

	res, err := programa.
		RegistrarComando(ajustarGM).
		Correr(os.Args[1:]...)

	if err != nil {
		fmt.Print(cadena.Fatal(" ", err))
	}
	if res != nil {
		// [HACER] ver si sirve que suba el output hasta main...
	}
	// LEER SUCCESS
}
