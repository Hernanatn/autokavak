# autoprice
Herramienta de linea de comandos escrita en Go que realiza las tareas del maikintosh en Price

<!-- <img src="https://img.shields.io/badge/hecho_por-Ch'aska-253545?style=for-the-badge" alt="hecho_por_Chaska" height="25px"/> -->
 <img src="https://img.shields.io/badge/Go-1.22-blue?style=for-the-badge&logo=go&logoColor=white" alt="C++" height="25px"/> <a href=https://www.raylib.com>
<img src="https://img.shields.io/badge/Versión-0.0.1--alpha-orange?style=for-the-badge" alt="version" height="25px"/></a> <img src="https://img.shields.io/badge/Licencia-HLQSLCEO-lightgrey?style=for-the-badge" alt="licencia" height="25px"/>

### Versión 0.0.1-alpha

Primer y probablemente única versión de este programa, es una pequeña herramienta para ayudar a un amigo. Todos los secretos y la data sensible están debidamente protegidos. Para poder usar la herramienta se debe compilar desde fuente, y se deben agregar las credenciales de cuenta de servicio y un módulo `data` dentro del directorio homónimo, de modo que la base debe presentar la siguiente estructura:

fuente  
├── main.go  
├── gsheets  
│   └── gsheets.go  
├── ***data***  
│   ├── ***data.go***  
│   └── ***credenciales.json***  

`data.go` debe definir las siguientes constantes:
```go

//go:embed credenciales.json
var CREDENCIALES []byte

const CORREO string // ID del correo que se usará como alias para las acciones que realice la cuenta de servicio
const IDHOJA_PUBLICAR string // ID de la hoja a leer
const IDHOJA_AJUSTAR string // ID de la hoja a leer

```

La implementación de la herramienta está profundamente acoplada a la estructura de la hoja que se utiliza como fuente. Sin embargo, ofrecer la estructura de esta hoja violaría el secreto profesional y provablemente sería un ataque a la propiedad de Price. Esta situación no debiera presentar un problema ya que esta herramienta tiene un solo usuario pretendido, el cual es autor de la hoja necesaria - y de todas formas no presenta valor para ningun usuario salvo el depto. de [?] de Price. Se publica únicamente con el proposito de proveer un ejemplo de herramienta real y funcional creada con [`aplicacion.go`](https://github.com/hernanatn/aplicacion.go).


#### Licencia
Ud. puede hacer lo que se le cante con el código, el cual se provee como está y por el cual ni su autor, Hernán A. Teszkiewicz Novick, ni Ch'aska S.R.L., ni Price se hacen responsables.

Ver [Licencia](LICENSE)
