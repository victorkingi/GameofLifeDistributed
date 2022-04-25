package gol

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/util"
)

type controllerChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioInput    <-chan byte
	ioOutput   chan<- byte
	keyPresses <-chan rune
}

const isAlive = byte(255)

var globClient *rpc.Client

/*  the exitCode will determine if all turns where completed on quiting or
the user interrupted execution i.e. exitCode = 1
*/
var exitCode = 0

type aliveCellsInfo struct {
	turns int
	alive int
}
type status struct {
	val string
	m   sync.Mutex
}

//prevents a race condition when changing pause and play status
func (s *status) Get() string {
	s.m.Lock()
	defer s.m.Unlock()
	return s.val
}

func (s *status) Set(val string) {
	s.m.Lock()
	defer s.m.Unlock()
	s.val = val
}

// sendByteWorldValues sends byte values of world to c.ioOutput channel for saving to folder /out
func sendByteWorldValues(world [][]byte, p Params, c controllerChannels, doneOutputWorld chan<- bool) {
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.ioOutput <- world[y][x]
		}
	}
	doneOutputWorld <- true
}

// saveOutput checks if image output is complete and if so, sends ImageOutputComplete event
func saveOutput(p Params, c controllerChannels, world [][]byte, turn int, doneSaving chan<- bool) {
	doneOutputWorld := make(chan bool)
	c.ioCommand <- ioOutput
	c.ioFilename <- strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(turn)
	go sendByteWorldValues(world, p, c, doneOutputWorld)
	<-doneOutputWorld
	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- ImageOutputComplete{
		CompletedTurns: turn,
		Filename:       strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(turn),
	}
	doneSaving <- true
}

// handleStates invokes a server methods that handle ticker, key presses and pause status
func handleStates(p Params, c controllerChannels, world [][]byte, currentState *aliveCellsInfo,
	ticker *time.Ticker, status *status) {
	doneSaving := make(chan bool)

	for {
		select {
		// if 2 seconds have past, send a request to server calling stubs.CellsHandler function
		// that returns the current number of completed turns and alive cells
		case <-ticker.C:
			if status.Get() == "Play" {
				//stubs.CellHandler doesn't require any request value hence the empty string
				makeTickerRequest(currentState)
				c.events <- AliveCellsCount{
					CompletedTurns: currentState.turns,
					CellsCount:     currentState.alive,
				}
			}
		//with each key press, call stubs.KeyHandler method with request being the event
		case key := <-c.keyPresses:
			switch key {
			//e.g. if event == "save", KeyResponse will be current world and completed turns
			case 's':
				response := makeKeyRequest("save")
				currentState.turns = response.CurrentTurn
				world = response.World
				go saveOutput(p, c, world, currentState.turns, doneSaving)
				<-doneSaving

			case 'q':
				exitCode = 1
				makeKeyRequest("quit")
				return

			case 'p':
				go changePauseState(c, currentState.turns, status)
				makeKeyRequest("pause")

			}
		default:
			if status.Get() == "Pause" {
				//user has paused
				//NB :- after pressing s or q a number of times while paused,
				//the execution will go into deadlock, hence a need to unpause after
				//every 2 or 3 's' presses
				unpause := <-c.keyPresses
				if unpause == 'p' {
					go changePauseState(c, currentState.turns, status)
					makeKeyRequest("unpause")

				} else if unpause == 's' {
					response := makeKeyRequest("save")
					currentState.turns = response.CurrentTurn
					world = response.World
					go saveOutput(p, c, world, currentState.turns, doneSaving)
					<-doneSaving

				} else if unpause == 'q' {
					exitCode = 1
					makeKeyRequest("quit")
					return
				}
			}
		}
	}
}

/*
	makeTickerRequest calls CellHandler method that returns
	number of alive cells and completed turns every 2 seconds
*/
func makeTickerRequest(currentState *aliveCellsInfo) {
	request := StateRequest{Start: ""}
	response := new(AliveResponse)

	err := globClient.Call(CellsHandler, request, response)
	if err != nil {
		log.Fatal(err)
	}
	currentState.alive = response.Cells
	currentState.turns = response.Turns
}

/*
	makeKeyRequest sends a request to server according to
	the key press event and receives a respective response
*/
func makeKeyRequest(event string) *KeyResponse {
	request := StateRequest{Start: event}
	response := new(KeyResponse)

	err := globClient.Call(KeyHandler, request, response)
	if err != nil {
		log.Fatal(err)
	}
	return response
}

/*
	changePauseState sends pause and unpause events and sets the status value from
	"Pause" to "Play" and vice versa
*/
func changePauseState(c controllerChannels, turn int, status *status) {
	if status.Get() == "Play" {
		c.events <- StateChange{turn, Paused}
		status.Set("Pause")

	} else if status.Get() == "Pause" {
		fmt.Println("Continuing")
		c.events <- StateChange{turn, Executing}
		status.Set("Play")
	}
}

func calculateAliveCells(p Params, world [][]byte) []util.Cell {
	var aliveCells []util.Cell
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == isAlive {
				aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
			}
		}
	}

	return aliveCells
}

func quitExecution(p Params, c controllerChannels, world [][]byte, turns int) {
	doneSaving := make(chan bool)
	go saveOutput(p, c, world, turns, doneSaving)
	//make sure saving is complete before exiting
	<-doneSaving
	quitCells := calculateAliveCells(p, world)
	c.events <- FinalTurnComplete{CompletedTurns: turns, Alive: quitCells}

	c.events <- StateChange{turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

	if c.keyPresses != nil {
		//give time for quitting event to be printed
		time.Sleep(2 * time.Second)
		os.Exit(exitCode)
	}
}

func serverGameOfLife(p Params, c controllerChannels) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	initialWorld, currentState, status, request := initialiseVariables(p, c)
	response := new(Response)

	sendInitialEvents(p, c, initialWorld)
	go handleStates(p, c, initialWorld, currentState, ticker, status)
	err := globClient.Call(GameHandler, request, response)
	if err != nil {
		log.Fatal(err)
	}
	if response.Status == "OK" {
		fmt.Println("Responded: " + response.Status)
		quitExecution(p, c, response.World, response.CurrentTurn)
	}
}

func sendInitialEvents(p Params, c controllerChannels, initialWorld [][]byte) {
	c.events <- AliveCellsCount{CompletedTurns: 0, CellsCount: 0}
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if initialWorld[y][x] == isAlive {
				c.events <- CellFlipped{CompletedTurns: 0, Cell: util.Cell{X: x, Y: y}}
			}
		}
	}
}

/*
	initialiseVariables initialises the world, current state,
	status to play, and Request to world and params values
*/
func initialiseVariables(p Params, c controllerChannels) ([][]byte, *aliveCellsInfo, *status, Request) {
	currentState := &aliveCellsInfo{
		turns: 0,
		alive: 0,
	}
	status := &status{}

	go func() {
		status.Set("Play")
	}()
	initialWorld := initialiseWorld(p, c)

	params := GolParams_{
		Turns:       p.Turns,
		ImageWidth:  p.ImageWidth,
		ImageHeight: p.ImageHeight,
	}
	request := Request{
		World:  initialWorld,
		Params: params,
	}
	return initialWorld, currentState, status, request
}

func initialiseWorld(p Params, c controllerChannels) [][]byte {
	initialWorld := make([][]byte, p.ImageHeight)
	for i := 0; i < p.ImageWidth; i++ {
		initialWorld[i] = make([]byte, p.ImageWidth)
	}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			initialWorld[y][x] = <-c.ioInput
		}
	}
	return initialWorld
}

// distributor divides the work between workers and interacts with other goroutines.
func controller(p Params, c controllerChannels) {
	if globClient == nil {
		addrPtr := flag.String("ip", "localhost:8030", "IP:port string to connect to")
		flag.Parse()
		fmt.Println("Server: ", *addrPtr)
		client, _ := rpc.Dial("tcp", *addrPtr)
		globClient = client
	}
	c.ioCommand <- ioInput
	c.ioFilename <- strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	serverGameOfLife(p, c)
}
