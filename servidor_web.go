package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
)

// Variables globales
var jwtKey = []byte("keysecreta")

var db *sql.DB

// Estrucuras
type Datos struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Marca struct {
	Id     int    `json:"id_marca"`
	Nombre string `json:"nombre"`
}

// Un HTML sencillo para un login
const loginPage = `
<!DOCTYPE html>
<html>
<head>
	<title>Login JWT</title>
</head>
<body>
	<h2>Iniciar Sesión</h2>
	<form action="/Tokendinamico" method="POST">
		<label>Usuario:</label><br>
		<input type="text" name="username"><br>
		<label>Contraseña:</label><br>
		<input type="password" name="password"><br><br>
		<input type="submit" value="Entrar">
	</form>
</body>
</html>
`

// Funciones para base de datos y la creación y validación del token
func conectarBaseDeDatos() {
	var err error
	dsn := "root:1234@tcp(localhost:3306)/joyeria"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error al abrir la conexión:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error al conectar con la base de datos:", err)
	}

	fmt.Println("Conexión exitosa a MySQL")
}

func generateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"Username": username,
		"exp":      time.Now().Add(2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(tokenHeader, "Bearer ") {
			http.Error(w, "sin autorizacion", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(tokenHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma")
			}
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// Valida el usuario y password. Para prueba usuario=usuario y password=1234. Muestra el token generado.
func validateJWT(w http.ResponseWriter, r *http.Request) {
	var req Datos

	// POST con form o JSON
	if r.Header.Get("Content-Type") == "application/json" {
		// Si viene JSON
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Error al decodificar JSON", http.StatusBadRequest)
			return
		}
	} else {
		// Si viene form
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error al leer formulario", http.StatusBadRequest)
			return
		}
		req.Username = r.FormValue("username")
		req.Password = r.FormValue("password")
	}

	if req.Username == "usuario" && req.Password == "1234" {
		token, err := generateJWT(req.Username)
		if err != nil {
			http.Error(w, "Error al generar el token", http.StatusInternalServerError)
			return
		}
		// Devuelve el token como JSON y como texto
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	} else {
		http.Error(w, "Usuario o contraseña incorrectos", http.StatusUnauthorized)
	}
}

// Handlers o controladores
func listarHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM marca")
	if err != nil {
		http.Error(w, "Error al consultar", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var marcas []Marca

	for rows.Next() {
		var m Marca
		if err := rows.Scan(&m.Id, &m.Nombre); err != nil {
			http.Error(w, "Error al escanear fila", http.StatusInternalServerError)
			return
		}
		marcas = append(marcas, m)
	}

	// Especificar el tipo de contenido JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(marcas)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("login").Parse(loginPage)
	tmpl.Execute(w, nil)
}

// Main
func main() {
	conectarBaseDeDatos()
	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/Tokendinamico", validateJWT)
	http.HandleFunc("/listar", middleware(listarHandler))

	fmt.Println("Servidor escuchando en http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
