package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

var db *sql.DB

func main() {

	ConnectDB()
	CreateTab()
	defer db.Close()

	r := gin.Default()
	r.Use(authMiddleware)
	r.POST("/customers", createCustHandler)
	r.GET("/customers/:id", getCustHandler)
	r.GET("/customers", getAllTCustHandler)
	r.PUT("/customers/:id", updateCustHandler)
	r.DELETE("/customers/:id", deleteCustHandler)
	r.Run(":2019")

}

func ConnectDB() {
	url := os.Getenv("DATABASE_URL")
	var err error
	db, err = sql.Open("postgres", url)
	if err != nil {
		log.Println("Connet to database error : ", err)
	}
}

func CreateTab() {

	createTb := `
		CREATE TABLE IF NOT EXISTS customer (
			id SERIAL PRIMARY KEY,
			name TEXT,
			email TEXT,
			status	TEXT
		);
		`
	_, err := db.Exec(createTb)
	if err != nil {
		log.Println("Can't create table : ", err)
		return
	}
	fmt.Println("Create table success")

}

func authMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "token2019" {
		log.Println("Middleware: Unauthorize")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
	}
	c.Next()
}

func createCustHandler(c *gin.Context) {
	var cust customer
	err := c.ShouldBindJSON(&cust)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	stmt, err := db.Prepare("INSERT INTO customer (name,email, status) values ($1, $2,$3) RETURNING id")
	if err != nil {
		log.Println("Can't prepare query one row statment", err)
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		c.Abort()
		return
	}
	row := stmt.QueryRow(cust.Name, cust.Email, cust.Status)
	err = row.Scan(&cust.ID)
	if err != nil {
		fmt.Println("Can't scan id", err)
		return
	}

	c.JSON(http.StatusCreated, cust)
}

func getCustHandler(c *gin.Context) {
	var err error
	cust := customer{}
	id := c.Param("id")
	cust.ID, err = strconv.Atoi(id)

	stmt, err := db.Prepare("SELECT name,email, status FROM customer where id=$1")
	if err != nil {
		log.Println("Can't prepare query customer : ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	rows := stmt.QueryRow(id)
	err = rows.Scan(&cust.Name, &cust.Email, &cust.Status)
	if err != nil {
		log.Println("Can't Scan row into variable", err)
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, cust)

}

func getAllTCustHandler(c *gin.Context) {
	stmt, err := db.Prepare("SELECT id, name,email, status FROM customer order by id")
	if err != nil {
		log.Println("Can't prepare query all customer statment : ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Println("Can't query all customer : ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	custs := []customer{}
	for rows.Next() {
		c1 := customer{}
		err := rows.Scan(&c1.ID, &c1.Name, &c1.Email, &c1.Status)
		if err != nil {
			log.Println("getAllTCustHandler : Can't Scan row into variable", err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			c.Abort()
			return
		}

		custs = append(custs, c1)
	}

	c.JSON(http.StatusOK, custs)
}

func updateCustHandler(c *gin.Context) {
	cust := customer{}
	id := c.Param("id")
	iid, err := strconv.Atoi(id)

	err = c.ShouldBindJSON(&cust)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		c.Abort()
		return
	}

	cust.ID = iid

	stmt, err := db.Prepare("UPDATE customer SET name=$2,email=$3,status=$4 WHERE id=$1;")
	if err != nil {
		log.Println("Can't prepare statment update", err)
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		c.Abort()
		return
	}
	if _, err := stmt.Exec(iid, cust.Name, cust.Email, cust.Status); err != nil {
		log.Println("Error execute update: ", err)
		return
	}
	c.JSON(http.StatusOK, cust)
}

func deleteCustHandler(c *gin.Context) {
	id := c.Param("id")
	iid, err := strconv.Atoi(id)

	stmt, err := db.Prepare("DELETE FROM customer WHERE id = $1")
	if err != nil {
		log.Println("Can't prepare statment delete", err)
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		c.Abort()
		return
	}
	if _, err := stmt.Exec(iid); err != nil {
		log.Println("Error execute delete: ", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
}
