package utils

import "elevatorproject/config"


func Mux(hallOrders [config.N_FLOORS][2]bool, cabOrders [config.N_FLOORS]bool) [config.N_FLOORS][config.N_BUTTONS]bool {

	out := [config.N_FLOORS][config.N_BUTTONS]bool{}

	for k := range out {
		copy(out[k][:], append(hallOrders[k][:], cabOrders[k]))
	}

	return out
}