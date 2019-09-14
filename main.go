package main

import (

	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
//	"strconv"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

)


var db *sql.DB


type Customers struct {
	ID     int `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`

}


func authMiddleware(c *gin.Context) {

	token := c.GetHeader("Authorization")
	if token != "token2019" {
		c.JSON(http.StatusUnauthorized, gin.H{"Error ": "Unauthorization"})
		c.Abort()
		return
	}
	c.Next()
}



func postCustomers(c *gin.Context) {

	t := Customers{}
	err := c.ShouldBindJSON(&t)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error":"JSON input is error ! " + err.Error()})
		return
	}

	row := db.QueryRow("INSERT INTO cust (name, email, status) values ($1,$2,$3) RETURNING id", t.Name, t.Email, t.Status)

	var id int
	err = row.Scan(&id)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error":"Insert is error " + err.Error()})
		return
	}

	//idnum := strconv.Itoa(id)
	t.ID = id
	c.JSON(http.StatusCreated, t)
}


func getAllCustomers(c *gin.Context) {
	
	found_flag := false
	
	a_customer := Customers{}
	all_customer := []Customers{}
	

	stmt, err := db.Prepare("SELECT id, name, email, status FROM cust")

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error":"SQL Preparation is error ! " + err.Error()})
		return
	}

	rows, err := stmt.Query()

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error":"Row Query is error ! " + err.Error()})
		return
	}


	for rows.Next() {

		err = rows.Scan(&a_customer.ID,&a_customer.Name,&a_customer.Email,&a_customer.Status)

		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"Error":"Scan row is error ! " + err.Error()})
			return
		}

		all_customer = append(all_customer, a_customer)
		found_flag = true
	} 
		
	
	if found_flag == false {
		c.JSON(http.StatusOK, gin.H{"message":"Not found customer"})
	} else {
		c.JSON(http.StatusOK, all_customer)
	}

}



func get1Customer(c *gin.Context) {

	id := c.Param("id")
	var a_customer Customers

	stmt, err := db.Prepare("SELECT id, name, email, status FROM cust WHERE id = $1")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error":"SQL Preparation is error ! " + err.Error()})
		return
	}

	row := stmt.QueryRow(id)
	err = row.Scan(&a_customer.ID,&a_customer.Name,&a_customer.Email,&a_customer.Status)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error":fmt.Sprintf("Select id %d error ! : %s", id, err.Error())})
		return
	}
	c.JSON(http.StatusOK, a_customer)
}


func deleteCust(c *gin.Context) {


	id := c.Param("id")

//	idnum := strconv.Itoa(id)


	stmt, err := db.Prepare("DELETE FROM cust WHERE id = $1")

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Delete SQL is error " + err.Error()})
		return
	}


	_, err = stmt.Exec(id)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Execute deletion error!!! " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})

}



func updCustomer(c *gin.Context) {


	id := c.Param("id")
	a := Customers{}
	err := c.ShouldBindJSON(&a)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error":"JSON input is error !" + err.Error()})
		return
	}

	stmt, err := db.Prepare("UPDATE cust SET name=$2,email=$3,status=$4 WHERE id=$1")

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error":"Prepare SQL for update error ! " + err.Error()})
		return
	}

	_, err = stmt.Exec(id, a.Name, a.Email, a.Status)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"Error":"Execute update error!!! " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, a)

}



func main() {

	createTable()

	r := gin.Default()
	r.Use(authMiddleware)



	
	r.POST("/customers", postCustomers)
	r.GET("/customers/:id", get1Customer)
	r.GET("/customers", getAllCustomers)
	r.PUT("/customers/:id", updCustomer)
	r.DELETE("/customers/:id", deleteCust)


	r.Run(":2019")
	defer db.Close()

}



func createTable() {

	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("Failed to connect to database", err)
		return
	}


	createTb := `
	CREATE TABLE IF NOT EXISTS cust (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);

	`

	_, err = db.Exec(createTb)

	if err != nil {

		log.Println("Failed to create table.", err)
		return
	}

	fmt.Println("Created table success.")

}