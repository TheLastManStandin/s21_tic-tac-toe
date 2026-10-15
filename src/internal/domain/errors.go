package domain

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrCannotConnectToGame = errors.New("cannot connect to game")
var ErrGameAccessDenied = errors.New("access denied to game")
var ErrNotYourTurn = errors.New("it is not your turn")
var ErrGameNotFound = errors.New("game not found")
var ErrWaitingForPlayers = errors.New("waiting for players")
var ErrCellOccupied = errors.New("cell is already occupied")
var ErrCantConnectGameWithRobot = errors.New("game against minimax, you cant connect to it")
var ErrGameAlreadyStarted = errors.New("game already started, you cant connect to it")
var ErrAlreadyInGame = errors.New("bro you already in game, chill")
