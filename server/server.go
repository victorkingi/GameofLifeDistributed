package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"time"
)

type GolParams_ struct {
	Turns       int
	ImageWidth  int
	ImageHeight int
}

type AliveResponse struct {
	Cells int
	Turns int
}

type StateRequest struct {
	Start string
}

type KeyResponse struct {
	World       [][]byte
	CurrentTurn int
}

type Response struct {
	World       [][]byte
	Status      string
	CurrentTurn int
}

type Request struct {
	World  [][]byte
	Params GolParams_
}

const alive = byte(255)
const dead = byte(0)

type currentState struct {
	alive       int
	currentTurn int
	world       [][]byte
}
type SecretGoLOperation struct{}
type SecretCellOperation struct{}
type SecretKeyPressOperation struct{}

var current currentState
var quit bool
var pause bool
var pauseBlock chan bool

func mod(x, m int) int {
	return (x + m) % m
}

func calculateNeighbours(p GolParams_, x, y int, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i != 0 || j != 0 {
				if world[mod(y+i, p.ImageHeight)][mod(x+j, p.ImageWidth)] == alive {
					neighbours++
				}
			}
		}
	}
	return neighbours
}

func calculateNextState(p GolParams_, world [][]byte) [][]byte {
	newWorld := make([][]byte, p.ImageHeight)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			neighbours := calculateNeighbours(p, x, y, world)
			if world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			} else {
				if neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			}
		}
	}
	return newWorld
}

func calculateAliveCells(p GolParams_, world [][]byte) int {
	var aliveCells = 0
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == alive {
				aliveCells++
			}
		}
	}

	return aliveCells
}

// gameOfLife is the function called by the testing framework.
// It returns an array of alive cells.
func (s *SecretGoLOperation) GameOfLife(req Request, res *Response) (err error) {
	world := req.World
	current.world = world
	current.currentTurn = 0
	current.alive = calculateAliveCells(req.Params, world)

	for turn := 0; turn < req.Params.Turns; turn++ {
		world = calculateNextState(req.Params, world)
		current.currentTurn = turn + 1
		current.alive = calculateAliveCells(req.Params, world)
		current.world = world
		if pause {
			go func() {
				<-pauseBlock
			}()
		}
		if quit {
			break
		}
	}

	res.World = world
	res.Status = "OK"
	res.CurrentTurn = current.currentTurn

	//reset values to prevent wrong output when new client connects
	current = currentState{
		alive:       0,
		currentTurn: 0,
		world:       nil,
	}
	quit = false
	pause = false
	pauseBlock = nil

	return
}

func (s *SecretCellOperation) ReportAliveCells(req StateRequest, res *AliveResponse) (err error) {
	res.Cells = current.alive
	res.Turns = current.currentTurn
	return
}

func (s *SecretKeyPressOperation) ReportKeyPress(req StateRequest, res *KeyResponse) (err error) {
	if req.Start == "save" {
		res.World = current.world
		res.CurrentTurn = current.currentTurn
	} else if req.Start == "quit" {
		res.World = current.world
		res.CurrentTurn = current.currentTurn
		quit = true
	} else if req.Start == "pause" {
		pause = true
	} else if req.Start == "unpause" {
		pause = false
		go func() {
			pauseBlock <- true
		}()
	}
	return
}

func handleError(err error) {
	// TODO: all
	// Deal with an error event.
	if err != nil {
		fmt.Println("error")
	}
}

func main() {
	portPtr := flag.String("port", ":8030", "port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&SecretGoLOperation{})
	rpc.Register(&SecretCellOperation{})
	rpc.Register(&SecretKeyPressOperation{})

	ln, err := net.Listen("tcp", *portPtr)
	defer ln.Close()
	handleError(err)
	fmt.Println("server is running...")
	for {
		rpc.Accept(ln)
	}
}
