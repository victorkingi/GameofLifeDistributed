package gol

var GameHandler = "SecretGoLOperation.GameOfLife"
var CellsHandler = "SecretCellOperation.ReportAliveCells"
var KeyHandler = "SecretKeyPressOperation.ReportKeyPress"

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
