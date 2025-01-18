package main

import (
	"encoding/json"
	"fmt"
	"kavak/data"
	"kavak/gsheets"
	"os"
	"slices"
	"strings"

	"github.com/hernanatn/aplicacion.go"
	"github.com/hernanatn/aplicacion.go/consola/cadena"
	"google.golang.org/api/sheets/v4"
)

var programa aplicacion.Aplicacion
var cliente *sheets.Service

func inicializar(a aplicacion.Aplicacion, args ...string) error {
	if len(args) <= 0 || len(args) > 0 && !slices.Contains(args, "--sin-ini") {
	}
	var c *gsheets.Credenciales
	json.Unmarshal(data.CREDENCIALES, c)

	var dominios []string = []string{}
	cliente, err := gsheets.ObtenerClienteGSheets(c, dominios, "mike@mike.mike")
	if err != nil {
		return err
	}
	return nil
}

func finalizar(a aplicacion.Aplicacion, args ...string) error {
	a.ImprimirLinea(aplicacion.Cadena("¡Adiós!"))
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
		RegistrarFinal(finalizar)

	var CREDENCIALES gsheets.Credenciales
	err := json.Unmarshal(data.CREDENCIALES, &CREDENCIALES)
	if err != nil {
		panic(69)
	}
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

				cliente.Spreadsheets.Get("[ID_HOJA]")

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
