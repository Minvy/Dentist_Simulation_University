package main

import (
	"fmt"
	"math/rand"
	"time"
)

/*
	3.b
	In part_2 If doctor attempts to move low-priority patient into high-priority when it is full,
	it will result in a deadlock, as the doctor will lock and wait for a free space become available,
	free space will not become available as he is not working.
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
	time.Sleep(20 * time.Second)
	fmt.Println("Program: Adding more patients at random")

	for i := low + high; i < low+high+high; i++ {
		if rand.Intn(2) == 1 {
			go patient(hwait, dent, i)
		} else {
			go patient(lwait, dent, i)
		}
	}

	time.Sleep(50 * time.Second)
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
					hwait <- patient
					fmt.Printf("Assistant: I'm moving one patient to high priority, I moved (%d) so far\n", priorChangeCount)
					priorChangeCount++
					timer.Reset(m)
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
