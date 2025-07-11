package base_de_datos

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func ConectarDB(datosBase CredencialesBase) *sql.DB {
	var errorVariable error
	datos := ""+ datosBase.Usuario + ":" + datosBase.Password + "@tcp(" + datosBase.Direccion_Puerto + ")/" + datosBase.BaseDeDatos
	baseDeDatos, errorVariable := sql.Open("Hola", datos)
	if errorVariable != nil {
		log.Fatal("Error al iniciar la conexión.")
	}
	errorVariable = baseDeDatos.Ping()
	if errorVariable != nil {
		log.Fatal("Error al conectar a la base de datos.")
	}
	log.Println("Base de datos conectada.")
	return baseDeDatos
}