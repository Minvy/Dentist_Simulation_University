package main

import (
	"fmt"
	"math/rand"
	"time"
)

/*
	3.b
	I don't think my implementations to part_2 or part_3, may result in a deadlock as my main method tests the algorithm quite rigorously
	at random intervals with random batches.
*/

func main() {
	dent := make(chan chan int)
	hwait := make(chan chan int, 100)
	lwait := make(chan chan int, 5)
	wait := make(chan chan int, 5)

	go dentist(wait, dent)
	go assistant(hwait, lwait, wait)

	high := 3
	low := 5

	for i := 0; i < high; i++ {
		go patient(hwait, dent, i)
	}

	for i := high; i < low+high; i++ {
		go patient(lwait, dent, i)
	}

	//Take a break and then introduce more patients at random
	//Take a break and then introduce batches of patients at semi-random intervals
	time.Sleep(10 * time.Second)

	//variable keeps in check the customer number
	index := high + low

	//Total number of customers served should be 107 (3+5+10*10)
	for x := 0; x < 10; x++ {
		time.Sleep(time.Duration((rand.Intn(10000))) * time.Millisecond)
		fmt.Print("Program: Adding batch of customers with random priorities \n")
		for i := 0; i < 10; i++ {
			time.Sleep(time.Duration((rand.Intn(1000))) * time.Millisecond)
			//Randomly assigns priority, with 1/5 being in high, and 4/5 being in low
			if rand.Intn(5) < 1 {
				go patient(hwait, dent, index)
			} else {
				go patient(lwait, dent, index)
			}
			index++
		}
	}

	time.Sleep(200 * time.Second)
}

func assistant(hwait chan chan int, lwait <-chan chan int, wait chan<- chan int) {
	//communicates with patients with hwait, lwait
	//communicates with dentist with wait
	var priorChangeCount int
	var m time.Duration = 2000 * time.Millisecond
	var timer = time.NewTimer(m)
	for {
		select {
		case <-timer.C:
			select {
			case patient, ok := <-lwait:
				if ok {
					select {
					case hwait <- patient:
						fmt.Printf("Assistant: I'm moving one patient to high priority, I moved (%d) so far\n", priorChangeCount)
						priorChangeCount++
						timer.Reset(m)
					default:
						fmt.Printf("Assistant: high-priority queue is full, I cannot move anyone \n")
						timer.Reset(m)
					}
				}
			default:
				{
					//lwait is empty so just reset timer
					timer.Reset(m)
				}
			}
		default:
			{
				select {
				case patient := <-hwait:
					wait <- patient
					//fmt.Printf("DEBUG: Wait <- Patient {Hwait}\n")
				default:
					select {
					case patient := <-lwait:
						wait <- patient
						//fmt.Printf("DEBUG: Wait <- Patient {Lwait}\n")
						timer.Reset(m)
					default:
						//Nothing because channels are both empty
					}
				}
			}
		}
	}
}

func dentist(wait chan chan int, dent <-chan chan int) {
	var clientNumber = 0
	for {
		select {
		case patient := <-dent:
			time.Sleep(time.Duration((500 + rand.Intn(1000))) * time.Millisecond)
			fmt.Printf("Dentist: I was awoken and had to serve a patient, I treated (%d) so far\n", clientNumber)
			clientNumber++
			patient <- 100
		default:
			fmt.Println("Dentist: Sleeping")
			time.Sleep(2 * time.Second)
		}
		for len(wait) > 0 {
			select {
			case patient := <-wait:
				fmt.Printf("Dentist: I'm treating a patient from the queue, I treated (%d) so far\n", clientNumber)
				time.Sleep(time.Duration((1000 + rand.Intn(1000))) * time.Millisecond)
				clientNumber++
				patient <- 100
			}
		}
	}
}

func patient(wait chan<- chan int, dent chan<- chan int, id int) {
	patient := make(chan int)
	fmt.Printf("Patient %d: I want a treatment\n", id)
	select {
	case dent <- patient:
		fmt.Printf("Patient %d: I'm waking up the dentist\n", id)
	case <-time.After(1 * time.Second):
		fmt.Printf("Patient %d: I'm waiting in the queue\n", id)
		wait <- patient
	}

	<-patient
	fmt.Printf("Patient %d: My teeth are shiny!\n", id)
}
