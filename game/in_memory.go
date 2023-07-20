package game

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// TODO: implement
// - AccuseAsMafia
// - ExecutePhase
// - GetPlayer
// - GetPlayers
// - VoteToKill
// - WaitForPhaseExecution

type InMemoryEngine struct {
	gameStatesMutex sync.RWMutex
	gameStates      map[string]*gameState
}

func NewInMemoryGameEngine() *InMemoryEngine {
	return &InMemoryEngine{
		gameStates: make(map[string]*gameState),
	}
}

func (i *InMemoryEngine) CancelGame(ctx context.Context, hostAddress string) error {
	i.gameStatesMutex.Lock()
	defer i.gameStatesMutex.Unlock()

	delete(i.gameStates, hostAddress)

	return nil
}

func (i *InMemoryEngine) FinishGame(ctx context.Context, hostAddress string) error {
	i.gameStatesMutex.Lock()
	defer i.gameStatesMutex.Unlock()

	delete(i.gameStates, hostAddress)

	return nil
}

func (i *InMemoryEngine) InitializeGame(_ context.Context, hostAddress string) error {
	i.gameStatesMutex.Lock()
	defer i.gameStatesMutex.Unlock()

	if _, hasGame := i.gameStates[hostAddress]; hasGame {
		return errors.New("a game cannot be initialized twice")
	}

	i.gameStates[hostAddress] = newGameState()

	return nil
}

func (i *InMemoryEngine) JoinGame(_ context.Context, hostAddress string, playerAddress string, playerNickname string) error {
	i.gameStatesMutex.Lock()
	defer i.gameStatesMutex.Unlock()

	game, hasGame := i.gameStates[hostAddress]
	if !hasGame {
		return errors.New("no game found")
	} else if game.started {
		return errors.New("cannot join a game already in progress")
	}

	if _, hasPlayer := game.players[playerAddress]; hasPlayer {
		return errors.New("cannot join a game multiple times")
	}

	game.players[playerAddress] = newPlayer(playerAddress)

	return nil
}

func (i *InMemoryEngine) StartGame(_ context.Context, hostAddress string) error {
	i.gameStatesMutex.Lock()
	defer i.gameStatesMutex.Unlock()

	game, hasGame := i.gameStates[hostAddress]
	if !hasGame {
		return errors.New("a game cannot be started without initialization")
	}

	if game.started {
		return errors.New("a game in progress cannot be started again")
	}

	if startErr := game.announceStart(); startErr != nil {
		return fmt.Errorf("failed to start game: %w", startErr)
	}

	return nil
}

func (i *InMemoryEngine) WaitForGameStart(ctx context.Context, hostAddress string) error {
	i.gameStatesMutex.Lock()
	defer i.gameStatesMutex.Unlock()

	game, hasGame := i.gameStates[hostAddress]
	if !hasGame {
		return errors.New("a game cannot be started without initialization")
	}

	subChan, err := game.subscribeToStart()
	if err != nil {
		return fmt.Errorf("failed to subscribe to game start: %w", err)
	}

	select {
	case <-subChan:
		return nil
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

type gameState struct {
	started bool
	players map[string]*Player

	gameStartSubs []chan any

	gameStartMutex sync.Mutex
}

func newGameState() *gameState {
	return &gameState{
		players: make(map[string]*Player),
	}
}

func (g *gameState) announceStart() error {
	g.gameStartMutex.Lock()
	defer g.gameStartMutex.Unlock()

	if g.started {
		return errors.New("game cannot be started multiple times")
	}

	for _, subChan := range g.gameStartSubs {
		subChan <- nil
		close(subChan)
	}

	g.gameStartSubs = nil

	return nil
}

func (g *gameState) subscribeToStart() (<-chan any, error) {
	g.gameStartMutex.Lock()
	defer g.gameStartMutex.Unlock()

	if g.started {
		return nil, errors.New("cannot subscribe to game start when it is already started")
	}

	newSub := make(chan any)
	g.gameStartSubs = append(g.gameStartSubs, newSub)
	return newSub, nil
}

func newPlayer(playerAddress string) *Player {
	return &Player{
		PlayerAddress: playerAddress,
	}
}
