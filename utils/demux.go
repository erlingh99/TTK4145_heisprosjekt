package utils

import "elevatorproject/config"


func Demux(orders [config.N_FLOORS][config.N_BUTTONS]bool) ([config.N_FLOORS][2]bool, [config.N_FLOORS]bool) {

	hallOrders := [config.N_FLOORS][2]bool{}

	cabOrders := [config.N_FLOORS]bool{}

	for k := range orders {
		copy(hallOrders[k][:], orders[k][:2])
		cabOrders[k] = orders[k][2]
	}

	return hallOrders, cabOrders
}