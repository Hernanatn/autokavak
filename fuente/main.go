package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hernanatn/autokavak/data"
	"github.com/hernanatn/autokavak/gsheets"
	"golang.org/x/sys/windows"
	"google.golang.org/api/sheets/v4"

	"github.com/hernanatn/aplicacion.go"
	"github.com/hernanatn/aplicacion.go/consola"
	"github.com/hernanatn/aplicacion.go/consola/cadena"
	"github.com/hernanatn/aplicacion.go/consola/color"
)

var programa aplicacion.Aplicacion
var servicio *gsheets.Servicio

var (
	DOLAR_KAVAK  float64
	CONDICION_FP any
)

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

	respuesta, err := servicio.Spreadsheets.Values.Get(data.ID_LIBRO_PUBLICAR, data.RANGO_DOLAR_KAVAK).Do()
	if err != nil {
		return err
	}

	DOLAR_KAVAK, err = gsheets.Celda{Valor: respuesta.Values[0][0]}.Flotante()
	if err != nil {
		return errors.Join(err, fmt.Errorf("no se pudo leer el dolar_KAVAK. %v", respuesta.Values[0]...))
	}

	respuesta, err = servicio.Spreadsheets.Values.Get(data.ID_LIBRO_PUBLICAR, data.RANGO_CONDICION_CARGA_FP).Do()
	if err != nil {
		return fmt.Errorf("no se pudo leer condicion first publish %v", respuesta.Values[0]...)
	}
	CONDICION_FP = respuesta.Values[0][0]

	respuesta, err = servicio.Spreadsheets.Values.Get(data.ID_LIBRO_PUBLICAR, data.RANGO_ESCALERA).Do()
	if err != nil {
		return err
	}
	escalera := gsheets.Celda{Valor: respuesta.Values[0][0]}.Cadena()
	if escalera != "FALSE" {
		return fmt.Errorf("escalera de PIX Activada. %v", respuesta.Values[0]...)
	}

	respuesta, err = servicio.Spreadsheets.Values.Get(data.ID_LIBRO_PUBLICAR, data.RANGO_SUCCESS_IDX).Do()
	if err != nil {
		return err
	}
	inventory := gsheets.Celda{Valor: respuesta.Values[0][0]}.Cadena()
	if inventory != "Success" {
		return fmt.Errorf("no corrió Inventory Index %v", respuesta.Values[0]...)
	}

	respuesta, err = servicio.Spreadsheets.Values.Get(data.ID_LIBRO_PUBLICAR, data.RANGO_SUCCESS_PDB).Do()
	if err != nil {
		return err
	}
	pdb := gsheets.Celda{Valor: respuesta.Values[0][0]}.Cadena()
	if pdb != "Success" {
		return fmt.Errorf("no corrió Pricing Database %v", respuesta.Values[0]...)
	}

	respuesta, err = servicio.Spreadsheets.Values.Get(data.ID_LIBRO_PUBLICAR, data.RANGO_PISO_TECHO).Do()
	if err != nil {
		return err
	}
	pisoTecho := gsheets.Celda{Valor: respuesta.Values[0][0]}.Cadena()
	if pisoTecho != "TRUE" {
		return fmt.Errorf("Piso/Techo no seteado %v", respuesta.Values[0]...)
	}

	respuesta, err = servicio.Spreadsheets.Values.Get(data.ID_LIBRO_PUBLICAR, data.RANGO_A_ESTRATEGICO).Do()
	if err != nil {
		return err
	}
	rangoAE := gsheets.Celda{Valor: respuesta.Values[0][0]}.Cadena()
	if rangoAE != "SI" {
		return fmt.Errorf("'Ajuste Estratégico' debería estar vacío y no lo está %v", respuesta.Values[0]...)
	}

	respuesta, err = servicio.Spreadsheets.Values.Get(data.ID_LIBRO_ESCALERITA, data.RANGO_FECHA_PDB).Do()
	if err != nil {
		return err
	}
	rangoFPDB := gsheets.Celda{Valor: respuesta.Values[0][0]}.Cadena()
	fecha := strings.Split(rangoFPDB, "-")
	año, err0 := strconv.Atoi(fecha[0])
	m, err1 := strconv.Atoi(fecha[1])
	dia, err2 := strconv.Atoi(fecha[2])

	if e := errors.Join(err0, err1, err2); e != nil {
		return errors.Join(e, fmt.Errorf("no se pudo leer la fecha de Pricing Database %v", respuesta.Values[0]...))
	}

	mes := time.Month(m)
	añoHoy, mesHoy, hoy := time.Now().Date()
	añoPasado := añoHoy - 1
	mesPasado := mesHoy - 1
	ayer := hoy - 1

	if (año != añoHoy && mesHoy != 1 && hoy != 1) || (año < añoPasado) || (mes != mesHoy && hoy > 1) || (mes < mesPasado) || dia < ayer {
		return fmt.Errorf("'Pricing' Database no registra las compras de ayer %v", respuesta.Values[0]...)
	}

	fechaEsperada := fmt.Sprintf("%.04d-%.02d-%.02d", añoHoy, mesHoy, ayer)
	respuesta, err = servicio.Spreadsheets.Values.Get(data.ID_LIBRO_ESCALERITA, fmt.Sprintf("%v:%v", data.RANGO_FECHA_PDB, "K")).Do()
	if err != nil {
		return fmt.Errorf("no se pudo leer la fecha de Pricing Database %v", respuesta.Values[0]...)
	}
	var validas int = 0
	for i, fecha := range respuesta.Values {
		fechaS := gsheets.Celda{Valor: fecha[0]}.Cadena()
		if fechaS == fechaEsperada {
			validas++
		}
		if validas > 9 {
			break
		}
		if i > 25 {
			return fmt.Errorf("'Pricing Database' registra menos de 10 compras de ayer %v", respuesta.Values[0]...)
		}
	}
	if validas < 10 {
		return fmt.Errorf("'Pricing Database' registra menos de 10 compras de ayer %v", respuesta.Values[0]...)
	}

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
		"Herramienta de gestión de precios y publicaciones de Kavak",
		make([]string, 0),
		aplicacion.NuevaConsola(os.Stdin, os.Stdout),
	).
		RegistrarInicio(inicializar).
		RegistrarLimpieza(finalizar).
		RegistrarFinal(limpiar)

	if strings.Contains(runtime.GOOS, "windows") {
		stdout := windows.Handle(os.Stdout.Fd())
		var originalMode uint32

		windows.GetConsoleMode(stdout, &originalMode)
		windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	}
}

func main() {

	var publicar aplicacion.Comando = aplicacion.NuevoComando(
		"publicar",
		"publicar",
		[]string{"-p"},
		"Primer tarea. Publicar Autos no Publicados. Verifica si hubo actualización exitosa en la base de datos, que la data esté actualizada al día anterior y construye el CSV para subir en el backoffice.",
		aplicacion.Accion(
			func(con aplicacion.Consola, opciones aplicacion.Opciones, parametros aplicacion.Parametros, argumentos ...any) (res any, cod aplicacion.CodigoError, err error) {
				if err != nil {
					return nil, aplicacion.ERROR, err
				}
				var csvSalida [][]string

				const idLibro string = data.ID_LIBRO_PUBLICAR

				// <HACER/> VALIDAR ULTIMA FECHA
				var actualizado bool
				// <HACER/> VALIDAR `SUCCESS` EN CARGA DE BDD
				var exito bool

				if !(actualizado && exito) {

					//return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirAdvertencia(cadena.CadenaFmt("La lista de autos no está actualizada. Ultima fecha registrada: %v | Estado de actualización BDD: %v\nAbortando programa..", actualizado, exito)), nil))
				}
				con.ImprimirLinea(cadena.Subtitulo("Analizando Autos sin precio."))
				const rangoAutosNuevos string = data.RANGO_PUBLICAR_AJUSTE

				respuesta, err := servicio.Spreadsheets.Values.Get(idLibro, rangoAutosNuevos).Do()
				if err != nil {
					return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirFatal(cadena.CadenaFmt("No se pudo leer el rango solicitado. Id hoja: %s | Rango: %s ", idLibro, rangoAutosNuevos), err))
				}

				var autosNuevos gsheets.TablaValores = gsheets.NuevaTabla(respuesta.Values)

				if len(autosNuevos) < 1 {

					return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirFatal(cadena.CadenaFmt("No se encontró data en el rango solicitado. Id hoja: %s | Rango: %s ", idLibro, rangoAutosNuevos), nil))
				}

				if len(autosNuevos) < 10 {
					continuar, err := con.Leer("Hay menos de diez autos nuevos.\n¿Continuar? [S/n]\n")
					if err != nil {
						return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirLinea(aplicacion.Cadena(cadena.Fatal("", err))))
					}

					if strings.ToLower(continuar.S()) != "s" {
						return nil, aplicacion.ERROR, con.ImprimirLinea("Ejecución abortada por el Usuario")
					}

					respuesta, err = servicio.Spreadsheets.Values.Get(idLibro, rangoAutosNuevos).Do()
					if err != nil {
						return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirFatal(cadena.CadenaFmt("No se pudo leer el rango solicitado. Id hoja: %s | Rango: %s ", idLibro, rangoAutosNuevos), err))
					}
					autosNuevos = gsheets.NuevaTabla(respuesta.Values)
				}

				for fila, auto := range autosNuevos {
					gmf, err := auto[data.COLUMNA_GM].Flotante()
					if err != nil {
						if auto[data.COLUMNA_GM].Cadena() == "" {
							continue
						}
						return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirFatal(cadena.CadenaFmt("No se encontró un valor válido para GM: %s", auto), err))
					}
					gmf /= 100
					if gmf < data.LIMITE_GM {
						nuevoGM, err := con.Leer(consola.Cadena(cadena.Advertencia(fmt.Sprintf("El auto %v, tiene GM < %f: %f. Ingrese ajuste", auto[1:5], data.LIMITE_GM, gmf), nil)))
						if err != nil {
							return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirLinea(aplicacion.Cadena(cadena.Fatal("No se pudo leer el ajuste", err))))
						}
						aj, err := strconv.ParseFloat(nuevoGM.S(), 64)
						if err != nil {
							return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirLinea(aplicacion.Cadena(cadena.Fatal("No se pudo leer el ajuste", err))))
						}
						ajuste := &sheets.ValueRange{
							Values: [][]interface{}{
								{fmt.Sprintf("%.0f%%", aj)},
							},
						}
						_, err = servicio.Spreadsheets.Values.Update(data.ID_LIBRO_PUBLICAR, fmt.Sprintf("%s%s%d", data.NOMBRE_HOJA_PUBLICAR, data.COLUMNA_AJUSTE_S, fila+3+(len(autosNuevos)-len(autosNuevos))), ajuste).
							ValueInputOption("USER_ENTERED").
							Do()
						if err != nil {
							return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirLinea(aplicacion.Cadena(cadena.Fatal("No se pudo escribir el ajuste a la hoja", err))))
						}
						autosNuevos[fila][data.COLUMNA_AJUSTE] = gsheets.Celda{Valor: ajuste.Values[0][0]}
					}
				}
				con.ImprimirLinea(consola.Cadena(cadena.Exito("Todos los ajustes necesarios cargados")))
				con.ImprimirSeparador()
				con.ImprimirLinea(cadena.Subtitulo("Filtrando autos que se pueden publicar"))

				var autosFiltrados gsheets.TablaValores
				for fila, auto := range autosNuevos {
					puedePublicar, err := auto[data.COLUMNA_PUEDE_PUBLICAR].Flotante()
					if err != nil {
						return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirFatal(cadena.CadenaFmt("No se encontró un valor válido para ¿Puede Publicar?: %s", auto), err))
					}
					if puedePublicar == 1 {
						autosFiltrados = append(autosFiltrados, autosNuevos[fila])
					}
				}

				con.ImprimirLinea(consola.Cadena(cadena.Exito(fmt.Sprintf("Autos publicados: %d", len(autosFiltrados)))).Colorear(color.AzulFondo))
				con.ImprimirSeparador()
				encabezados := []string{"#Stock", "Marca", "Modelo", " Año ", "Kilom.", "PIX F", "Sell Price", "GM", " .% "}
				filas := [][]string{}

				for fila, auto := range autosFiltrados {
					var ajuste float64 = 0
					pixr2s, err := auto[data.COLUMNA_PIX_R2S].Flotante()
					a := auto[data.COLUMNA_AJUSTE].Cadena()
					if len(a) > 0 {
						var e error
						ajuste, e = auto[data.COLUMNA_AJUSTE].Flotante()
						err = errors.Join(err, e)
					}
					if err != nil {
						return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirFatal(cadena.CadenaFmt("No se encontró un valor válido para PIX_R2S o AJUSTE: %s", auto), err))
					}

					pixFinal := pixr2s + ajuste
					autosFiltrados[fila][data.COLUMNA_PIX_F] = gsheets.Celda{Valor: fmt.Sprintf("%.0f%%", pixFinal)}

					pm, err := auto[data.COLUMNA_PM].Flotante()
					if err != nil {
						return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirFatal(cadena.CadenaFmt("No se encontró un valor válido para PIX_R2S o AJUSTE: %s", auto), err))
					}

					pv := math.Round((pm*(1+pixFinal/100))/10_000) * 10_000
					autosFiltrados[fila][data.COLUMNA_PV] = gsheets.Celda{Valor: fmt.Sprintf("%.0f", pv)}

					bodytype := auto[data.COLUMNA_BT].Cadena()

					ptusd, err := auto[data.COLUMNA_PT_USD].Flotante()
					if err != nil {
						return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirFatal(cadena.CadenaFmt("No se encontró un valor válido para Precio de Toma USD: %s", auto), err))
					}

					dmusd, err := auto[data.COLUMNA_DM_USD].Flotante()
					if err != nil {
						return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirFatal(cadena.CadenaFmt("No se encontró un valor válido para Precio de Toma USD: %s", auto), err))
					}

					var iva float64 = 1.21
					switch bodytype {
					case "Utilitarian", "Pick Up", "Van":
						iva = 1.105
					}

					gmf := pv/iva/DOLAR_KAVAK - ptusd - dmusd
					autosFiltrados[fila][data.COLUMNA_GM_F] = gsheets.Celda{Valor: fmt.Sprintf("%.0f", math.Round(gmf))}

					gmp := (gmf * DOLAR_KAVAK / pv) * 100
					autosFiltrados[fila][data.COLUMNA_GM_P] = gsheets.Celda{Valor: fmt.Sprintf("%.0f%%", math.Round(gmp))}
					var tabla gsheets.Fila = auto[data.COLUMNA_STOCK_ID : data.COLUMNA_KM+1]
					tabla = append(tabla, auto[data.COLUMNA_PIX_F:data.COLUMNA_GM_P+1]...)

					filas = append(filas, tabla.Cadenas())
					//con.ImprimirLinea(cadena.Cadena(fmt.Sprintf("| %-8s | %-15s | %-23s | %-7s | %-8s | %-6s | %-12s | %-7s | %-6s |", tabla...)))
				}
				//con.ImprimirLinea("+----------+-----------------+-------------------------+---------+----------+--------+--------------+---------+--------+")
				con.ImprimirLinea(consola.Cadena(cadena.Tabla(encabezados, filas)))
				respuesta, err = servicio.Spreadsheets.Values.Get(data.ID_LIBRO_PUBLICAR, data.RANGO_CANTIDAD_HISTORICOS).Do()
				if err != nil {
					return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirLinea(aplicacion.Cadena(cadena.Fatal("No se pudo leer de historicos", err))))

				}

				var cantidadHistoricos int
				if canth, ok := respuesta.Values[0][0].(string); ok {
					cantidadHistoricos, err = strconv.Atoi(canth)
					if err != nil {
						return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirLinea(aplicacion.Cadena(cadena.Fatal("No se pudo leer de historicos", err))))
					}
				} else {
					return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirLinea(aplicacion.Cadena(cadena.Fatal("No se pudo leer de historicos", err))))
				}
				if !(cantidadHistoricos > 0) {
					return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirLinea(aplicacion.Cadena(cadena.Fatal("No se pudo leer de historicos", err))))
				}

				fila := cantidadHistoricos + 1

				valoresAutosFiltrados := &sheets.ValueRange{
					Values: autosFiltrados.Anys(),
				}

				_, err = servicio.Spreadsheets.Values.Update(data.ID_LIBRO_PUBLICAR, fmt.Sprintf("%s%s%d:%s", data.NOMBRE_HOJA_HISTORICOS, "A", fila, "AC"), valoresAutosFiltrados).
					ValueInputOption("USER_ENTERED").
					Do()
				if err != nil {
					return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirLinea(aplicacion.Cadena(cadena.Fatal("No se pudo escribir el ajuste a la hoja", err))))
				}

				con.ImprimirLinea("")
				con.ImprimirLinea(consola.Cadena(cadena.Exito("Copiado a Historico PIXR2S")))

				tablaFP, err := gsheets.EsperarYLeerCondicion(servicio, data.ID_LIBRO_PUBLICAR, data.RANGO_CARGA_FP, data.INTENTOS_MAXIMOS, data.RETARDO_BASE, [2]int{0, 0}, CONDICION_FP)
				if err != nil {
					return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirLinea(aplicacion.Cadena(cadena.Fatal("No se pudo leer Carga First Publish", err))))
				}

				con.ImprimirLinea(consola.Cadena(cadena.Exito("Carga First Publish actualizado.")).Colorear(color.AzulFondo))
				con.ImprimirSeparador()
				con.ImprimirLinea(cadena.Subtitulo("Creando CSV."))

				csvSalida, err = gsheets.ConvertirValoresGSaCSV(tablaFP)
				if err != nil {
					return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirLinea(aplicacion.Cadena(cadena.Fatal("No se pudo convertir los resultados de Carga First Publish", err))))
				}

				var nombreArchivo string
				if _, ok := parametros["-salida"]; ok {
					nA, okS := (parametros["-salida"]).(string)
					if okS {
						nombreArchivo = nA
					} else {

						nombreArchivo = data.NOMBRE_CSV_PUBLICAR
					}
				} else {
					nombreArchivo = data.NOMBRE_CSV_PUBLICAR
				}
				archivo, err := os.Create(nombreArchivo)
				if err != nil {
					panic(err)
				}
				defer archivo.Close()

				escritor := csv.NewWriter(archivo)
				defer escritor.Flush()

				err = escritor.WriteAll(csvSalida)
				if err != nil {
					return nil, aplicacion.ERROR, errors.Join(err, con.ImprimirFatal(consola.Cadena("No se pudo construir el archivo CSV"), err))
				}

				con.ImprimirLinea(consola.Cadena("CSV Creado con éxtio.").Colorear(color.AzulFuente))
				con.ImprimirLinea(aplicacion.Cadena(cadena.Exito("¡Todo listo!")).Colorear(color.AzulFondo))
				return nil, aplicacion.EXITO, nil
			}),
		make([]string, 0),
	)
	res, err := programa.
		RegistrarComando(publicar).
		Correr(os.Args[1:]...)
	if err != nil {
		fmt.Print(cadena.Fatal("La aplicación terminó forsozamente: ", err))
	}
	if res != nil {
		// [HACER] ver si sirve que suba el output hasta main...
	}
	// LEER SUCCESS
}
