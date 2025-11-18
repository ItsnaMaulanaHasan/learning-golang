package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var (
	host     = os.Getenv("DB_HOST") // Use the container name
	user     = os.Getenv("DB_USER")
	password = os.Getenv("DB_PASSWORD")
	dbname   = os.Getenv("DB_NAME")
)

var db *sql.DB
var App *gin.Engine

func init() {
	// Connect to the database
	psqlInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=require",
		host, user, password, dbname)
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Successfully connected to the database!")

	// Initialize Gin router
	App = gin.New()

	// Load HTML templates
	App.LoadHTMLGlob("templates/*")

	// Define routes
	App.GET("/", func(c *gin.Context) {
		// Fetch all users and render the index page
		users, err := getUsersFromDB()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{"users": users})
	})

	App.GET("/users", func(c *gin.Context) {
		// Fetch all users and return as JSON
		users, err := getUsersFromDB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, users)
	})

	App.POST("/users", createUser)
	App.PUT("/users/:id", updateUser)
	App.DELETE("/users/:id", deleteUser)

	// Start the server
	App.Run(":8080")
}

// Handler to create a new user
func createUser(c *gin.Context) {
	var user struct {
		Name       string `json:"name"`
		Department string `json:"department"`
		Email      string `json:"email"`
	}
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Printf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID int
	err := db.QueryRow(`INSERT INTO users (name, department, email) VALUES ($1, $2, $3) RETURNING id`,
		user.Name, user.Department, user.Email).Scan(&userID)
	if err != nil {
		log.Printf("Failed to insert user into database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": userID})
}

// Handler to update a user
func updateUser(c *gin.Context) {
	id := c.Param("id")
	var user struct {
		Name       string `json:"name"`
		Department string `json:"department"`
		Email      string `json:"email"`
	}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("UPDATE users SET name = $1, department = $2, email = $3 WHERE id = $4",
		user.Name, user.Department, user.Email, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

// Handler to delete a user
func deleteUser(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

// Fetch all users from the database
func getUsersFromDB() ([]map[string]interface{}, error) {
	rows, err := db.Query("SELECT id, name, department, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var name, department, email string
		err = rows.Scan(&id, &name, &department, &email)
		if err != nil {
			return nil, err
		}
		users = append(users, gin.H{"id": id, "name": name, "department": department, "email": email})
	}

	return users, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	App.ServeHTTP(w, r)
}
