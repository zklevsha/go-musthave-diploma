package processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/zklevsha/go-musthave-diploma/internal/helpers"
	"github.com/zklevsha/go-musthave-diploma/internal/interfaces"
	"github.com/zklevsha/go-musthave-diploma/internal/structs"
)

type Processor struct {
	Delay   time.Duration
	Ctx     context.Context
	Wg      *sync.WaitGroup
	Storage interfaces.Storage
	Accrual string
}

func (p Processor) Start() {
	log.Println("INFO processor have started")
	defer p.Wg.Done()
	ticker := time.NewTicker(p.Delay)
	for {
		select {
		case <-p.Ctx.Done():
			log.Println("INFO processor has received  a ctx.Done()'. Exiting...")
			return
		case <-ticker.C:
			p.processOrders()
		}
	}
}

func (p Processor) processOrders() {
	log.Println("INFO processor starting process orders")
	log.Println("INFO processor getting list of unprocessed orders")
	orders, err := p.Storage.GetUnprocessedOrders()
	if err != nil {
		log.Printf("ERROR processor failed to get list of orders: %s",
			err.Error())
		return
	}
	if len(orders) == 0 {
		log.Printf("INFO processor there are no unprocessed orders. Sleeping for %s", p.Delay)
	}
	log.Printf("INFO processor received %d unprocessed orders. Begin proccessing.", len(orders))
	for _, o := range orders {
		log.Printf("INFO processor processing order %d", o)
		err := p.processOrder(o)
		if err != nil {
			if errors.Is(err, structs.ErrToManyRequests) {
				log.Println("ERROR processor received to many requests from accrual system. " +
					"Stopping order processing")
				break
			}
			log.Printf("ERROR processor failed to process order %d: %s", o, err.Error())
		}
		log.Printf("INFO processor finished processing order %d", o)
	}
	log.Printf("INFO processor finished processing orders. Sleeping for %s", p.Delay)
}

func (p Processor) processOrder(id int) error {
	order, err := p.GetOrderAccrual(id)
	if err != nil {
		return err
	}
	log.Printf("INFO processor received order status from accrual: %v", order)
	if order.Status == "INVALID" || order.Status == "PROCESSED" {
		log.Printf("INFO proccessor updating order status "+
			"(PROCESSING -> %s)", order.Status)
		rowsAffected, err := p.Storage.SetOrderStatus(id, order.Status)
		if err != nil {
			e := fmt.Sprintf("ERROR processor failed to update order status: %s", err.Error())
			return errors.New(e)
		}
		if rowsAffected != 1 {
			e := fmt.Sprintf("ERROR processor failed to update order status: "+
				"invalid number of affected rows: %d", rowsAffected)
			return errors.New(e)
		}
	}
	if order.Status == "PROCESSED" && order.Accrual != nil {
		log.Printf("INFO proccessor setting order`s accural value (%f)",
			*order.Accrual)
		rowsAffected, err := p.Storage.SetOrderAccrual(id, *order.Accrual)
		if err != nil {
			e := fmt.Sprintf("ERROR processor failed to update order accural: %s", err.Error())
			return errors.New(e)
		}
		if rowsAffected != 1 {
			e := fmt.Sprintf("ERROR processor failed to update order accrual: "+
				"invalid number of affected rows: %d", rowsAffected)
			return errors.New(e)
		}
	}
	return nil
}

func (p Processor) GetOrderAccrual(orderID int) (structs.Order, error) {
	url := fmt.Sprintf("%s/api/orders/%d", p.Accrual, orderID)
	resp, err := http.Get(url)
	if err != nil {
		return structs.Order{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return structs.Order{}, structs.ErrToManyRequests
	}
	var order structs.Order
	err = json.NewDecoder(resp.Body).Decode(&order)
	if err != nil {
		return structs.Order{}, err
	}
	// round to 2 digits
	var fixed float32
	if order.Accrual != nil {
		fixed = helpers.ToFixed(*order.Accrual, 2)
		order.Accrual = &fixed
	}
	return order, nil
}
