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
	milk := &Product{Name: "Milk", Amount: 10}
	w.Area[5][6] = milk
	bread := &Product{Name: "Bread", Amount: 10}
	w.Area[1][5] = bread
	salt := &Product{Name: "Salt", Amount: 10}
	w.Area[6][5] = salt
	soap := &Product{Name: "Soap", Amount: 10}
	w.Area[8][9] = soap
	pasta := &Product{Name: "Pasta", Amount: 10}
	w.Area[9][2] = pasta
	w.Products = append(w.Products, milk, bread, salt, soap, pasta)
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
	// r.AlertWarehouse({"event " : "something_else"})
}

func (r *Robot) PutToStock() {
	// do something..
	// r.AlertWarehouse({"event " : "something"})
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

	//handle error somehow...
	return errors.New("Could not find task")
}
func (a *Actions) GetUniqueID() int {
	a.Counter++
	return a.Counter
}

func (a *Actions) UpdateOrders(s []string) {
	for _, name := range s {

		id := a.GetUniqueID()
		task := &Task{
			Item: &Product{
				Name:   name,
				Amount: 1,
			},
			ID:         id,
			StatusOpen: true,
		}
		a.Order[a.Counter] = task
	}
}

func (a *Actions) UpdateSupplies(s []string) {
	for _, name := range s {
		id := a.GetUniqueID()
		task := &Task{
			Item: &Product{
				Name:   name,
				Amount: 1,
			},
			ID:         id,
			StatusOpen: true,
		}
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
				"TaskID":  taskID,
			})
		}

		ActionErr := actions.ActionComplete(id)
		if ActionErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Message": err.Error(),
				"TaskID":  taskID,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"Message": "Update Item Completed Successfully",
			"TaskID":  taskID,
		})
	})

	router.GET("/next-tasks", func(c *gin.Context) {
		type data struct {
			ID       int
			Action   string
			Product  string
			Location [][]int
		}

		var DataRes []data

		for _, n := range actions.Order {
			if n.StatusOpen == true {
				DataRes = append(DataRes, data{
					ID:      n.ID,
					Action:  "put_to_stock",
					Product: n.Item.Name,
					// location: [1][1]

				})
			}
		}

		for _, n := range actions.Supply {
			if n.StatusOpen == true {
				DataRes = append(DataRes, data{
					ID:      n.ID,
					Action:  "pick_from_stock",
					Product: n.Item.Name,
					// location: [1][1]

				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"data": DataRes,
		})

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
