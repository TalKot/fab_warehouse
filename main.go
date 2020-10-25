package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Warehouse struct {
	Name     string
	ID       int
	Products []*Product
	Area     [10][10]*Product
}

func (w *Warehouse) Initialize() {
	m := &Product{Name: "Milk", Amount: 10}
	w.Area[5][6] = m
	b := &Product{Name: "Bread", Amount: 10}
	w.Area[1][5] = b
	s := &Product{Name: "Salt", Amount: 10}
	w.Area[6][5] = s
	soap := &Product{Name: "Soap", Amount: 10}
	w.Area[8][9] = soap
	pasta := &Product{Name: "Pasta", Amount: 10}
	w.Area[9][2] = pasta
	w.Products = append(w.Products, m, b, s, soap, pasta)
}

func (w *Warehouse) Order() {

}

func (w *Warehouse) Supply() {

}

type Product struct {
	Name   string
	Amount int
}

// convert from robot?
type Robot struct {
	Name string
	ID   int
}

func (r *Robot) PickFromStock() {
	// do something..
	// r.AlertWarehouse({"asdf " : "asdf"})
}

func (r *Robot) PutToStock() {
	// do something..
	// r.AlertWarehouse({"asdf " : "asdf"})
}

func (r *Robot) AlertWarehouse(event map[string]string) {
	// do something..
}

type Task struct {
	Item       *Product
	ID         int
	StatusOpen bool
}

// Order - a list of products that were ordered by the customer
// Supply - a list of products that were supplied by the supplier
type Actions struct {
	Counter int
	Order   map[int]*Task
	Supply  map[int]*Task
}

func (a *Actions) Init() {
	a.Order = make(map[int]*Task)
	a.Supply = make(map[int]*Task)
}

func (a *Actions) ActionComplete(id int) error {
	if _, ok := a.Order[id]; ok {
		a.Order[id].StatusOpen = false
		return nil
	}

	if _, ok := a.Supply[id]; ok {
		a.Supply[id].StatusOpen = false
		return nil
	}

	//handle somehing...
	return errors.New("Could not find task")
}

func (a *Actions) UpdateOrders(s []string) {
	for _, n := range s {
		a.Counter++
		p := &Product{Name: n, Amount: 1}
		task := &Task{Item: p, ID: a.Counter, StatusOpen: true}
		a.Order[a.Counter] = task
	}
}

func (a *Actions) UpdateSupplies(s []string) {
	for _, n := range s {
		a.Counter++
		p := &Product{Name: n, Amount: 1}
		task := &Task{Item: p, ID: a.Counter, StatusOpen: true}
		a.Supply[a.Counter] = task
	}
}

type dataRequest struct {
	Order  []string
	Supply []string
}

func main() {
	router := gin.Default()
	//Initialize Warehouse
	w := Warehouse{}
	w.Initialize()

	//Initialize actions
	actions := Actions{Counter: 0}
	actions.Init()

	router.POST("/order", func(c *gin.Context) {

		bodyBytes, bodyErr := ioutil.ReadAll(c.Request.Body)
		if bodyErr != nil {
			fmt.Println(bodyErr)
		}
		// Restore the io.ReadCloser to its original state
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		data := dataRequest{}
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			fmt.Println(err)
		}

		if len(data.Order) > 0 {
			actions.UpdateOrders(data.Order)
		}

		c.JSON(http.StatusOK, gin.H{
			"Message": "Order Completed Successfully",
			"Order":   data.Order,
		})

	})

	router.POST("/supply", func(c *gin.Context) {
		bodyBytes, bodyErr := ioutil.ReadAll(c.Request.Body)
		if bodyErr != nil {
			fmt.Println(bodyErr)
		}
		// Restore the io.ReadCloser to its original state
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		data := dataRequest{}
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			fmt.Println(err)
		}

		if len(data.Supply) > 0 {
			actions.UpdateSupplies(data.Supply)
		}

		c.JSON(http.StatusOK, gin.H{
			"Message": "Supply Completed Successfully",
			"Supply":  data.Supply,
		})
	})

	router.POST("/task/:id/complete", func(c *gin.Context) {
		taskID := c.Param("id")
		id, err := strconv.Atoi(taskID)
		if err != nil {
			fmt.Println("Error : ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"Message": err.Error(),
				"Task_ID": taskID,
			})
		}

		ActionErr := actions.ActionComplete(id)
		if ActionErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Message": err.Error(),
				"Task_ID": taskID,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"Message": "Update Item Completed Successfully",
			"Task_ID": taskID,
		})
	})

	router.GET("/next-tasks", func(c *gin.Context) {
		// for i, n := range actions.Order {

		// }
	})

	router.GET("/stock", func(c *gin.Context) {
		data := make(map[string]struct {
			Name   string
			Amount int
		})

		for _, n := range w.Products {
			p := Product{
				Name:   n.Name,
				Amount: n.Amount,
			}
			data[n.Name] = p
		}

		c.JSON(http.StatusOK, data)
	})

	router.Run(":8080")
}
