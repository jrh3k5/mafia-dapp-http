package game

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
)

type InMemoryEngine struct {
	gameStatesMutex sync.RWMutex
	gameStates      map[string]*gameState
}

func NewInMemoryGameEngine() *InMemoryEngine {
	return &InMemoryEngine{
		gameStates: make(map[string]*gameState),
	}
}

func (i *InMemoryEngine) AccuseAsMafia(ctx context.Context, hostAddress string, accuserAddress string, accuseeAddress string) error {
	gameState, hasGameState := i.getGameState(hostAddress)
	if !hasGameState {
		return fmt.Errorf("failed to find game state for host address '%s'", hostAddress)
	}

	return gameState.accuseAsMafia(accuserAddress, accuseeAddress)
}

func (i *InMemoryEngine) CancelGame(ctx context.Context, hostAddress string) error {
	i.gameStatesMutex.Lock()
	defer i.gameStatesMutex.Unlock()

	delete(i.gameStates, hostAddress)

	return nil
}

func (i *InMemoryEngine) ExecutePhase(ctx context.Context, hostAddress string) error {
	gameState, hasGameState := i.getGameState(hostAddress)
	if !hasGameState {
		return fmt.Errorf("failed to find game state for host address '%s'", hostAddress)
	}

	currentPhase := gameState.getCurrentPhase()

	phaseExecution := &PhaseExecution{
		HostAddress:  hostAddress,
		CurrentPhase: currentPhase,
	}
	switch currentPhase {
	case TimeOfDayDay:
		accusedAddress, hasAccused := gameState.tallyMafiaVotes()
		if hasAccused {
			phaseExecution.ConvictedPlayers = append(phaseExecution.ConvictedPlayers, accusedAddress)
		}
	case TimeOfDayNight:
		killedAddress, hasKilled := gameState.tallyKillVotes()
		if hasKilled {
			phaseExecution.KilledPlayers = append(phaseExecution.KilledPlayers, killedAddress)
		}
	default:
		return fmt.Errorf("unhandled phase: %v", currentPhase)
	}

	phaseExecution.PhaseOutcome = gameState.calculatePhaseOutcome()

	gameState.notifyOfPhaseExecution(phaseExecution)

	return nil
}

func (i *InMemoryEngine) FinishGame(ctx context.Context, hostAddress string) error {
	i.gameStatesMutex.Lock()
	defer i.gameStatesMutex.Unlock()

	delete(i.gameStates, hostAddress)

	return nil
}

func (i *InMemoryEngine) GetPlayer(ctx context.Context, hostAddress string, playerAddress string) (*Player, error) {
	gameState, hasGameState := i.getGameState(hostAddress)
	if !hasGameState {
		return nil, fmt.Errorf("no game found for host address '%s'", hostAddress)
	}

	return gameState.getPlayer(playerAddress), nil
}

func (i *InMemoryEngine) GetPlayers(ctx context.Context, hostAddress string) ([]*Player, error) {
	gameState, hasGameState := i.getGameState(hostAddress)
	if !hasGameState {
		return nil, fmt.Errorf("no game found for host address '%s'", hostAddress)
	}

	return gameState.getPlayers(), nil
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
	game, hasGame := i.getGameState(hostAddress)
	if !hasGame {
		return errors.New("no game found")
	} else if game.started {
		return errors.New("cannot join a game already in progress")
	}

	if player := game.getPlayer(playerAddress); player != nil {
		return errors.New("cannot join a game multiple times")
	}

	game.addPlayer(newPlayer(playerAddress, playerNickname))

	return nil
}

func (i *InMemoryEngine) StartGame(_ context.Context, hostAddress string) error {
	game, hasGame := i.getGameState(hostAddress)
	if !hasGame {
		return errors.New("a game cannot be started without initialization")
	}

	if game.started {
		return errors.New("a game in progress cannot be started again")
	}

	// assign roles - one mafia for every five players, rounded up
	players := game.getPlayers()
	mafiaCount := int(math.Ceil(float64(len(players)) / 5))
	playersCopy := make([]*Player, len(players))
	copy(playersCopy, players)
	rand.Shuffle(len(playersCopy), func(i, j int) {
		old := playersCopy[i]
		playersCopy[i] = playersCopy[j]
		playersCopy[j] = old
	})
	for playerIndex, player := range playersCopy {
		if playerIndex < mafiaCount {
			player.PlayerRole = PlayerRoleMafia
		} else {
			player.PlayerRole = PlayerRoleCivilian
		}
	}

	if startErr := game.announceStart(); startErr != nil {
		return fmt.Errorf("failed to start game: %w", startErr)
	}

	return nil
}

func (i *InMemoryEngine) VoteToKill(ctx context.Context, hostAddress string, killerAddress string, killeeAddress string) error {
	gameState, hasGameState := i.getGameState(hostAddress)
	if !hasGameState {
		return fmt.Errorf("no game found for host address '%s'", hostAddress)
	}

	return gameState.voteToKill(killerAddress, killeeAddress)
}

func (i *InMemoryEngine) WaitForGameStart(ctx context.Context, hostAddress string) error {
	game, hasGame := i.getGameState(hostAddress)
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

func (i *InMemoryEngine) WaitForPhaseExecution(ctx context.Context, hostAddress string) (*PhaseExecution, error) {
	game, hasGame := i.getGameState(hostAddress)
	if !hasGame {
		return nil, fmt.Errorf("no game state found for host address '%s'", hostAddress)
	}

	subChan, err := game.subscribeToPhaseExecution()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to phase execution: %w", err)
	}

	select {
	case phaseExecution := <-subChan:
		return phaseExecution, nil
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	}
}

type gameState struct {
	started bool

	players      map[string]*Player
	playersMutex sync.RWMutex

	currentPhase      TimeOfDay
	currentPhaseMutex sync.RWMutex

	gameStartSubs  []chan any
	gameStartMutex sync.Mutex

	phaseExecutionSubs  []chan *PhaseExecution
	phaseExecutionMutex sync.RWMutex

	mafiaAccusations      map[string]string
	mafiaAccusationsMutex sync.RWMutex

	killVotes      map[string]string
	killVotesMutex sync.RWMutex
}

func newGameState() *gameState {
	return &gameState{
		players:          make(map[string]*Player),
		mafiaAccusations: make(map[string]string),
		killVotes:        make(map[string]string),
	}
}

func (g *gameState) announceStart() error {
	g.gameStartMutex.Lock()
	defer g.gameStartMutex.Unlock()

	if g.started {
		return errors.New("game cannot be started multiple times")
	}

	fmt.Printf("Notifying %d subscribers of game start\n", len(g.gameStartSubs))

	for _, subChan := range g.gameStartSubs {
		subChan <- nil
		close(subChan)
	}

	g.gameStartSubs = nil

	return nil
}

func (g *gameState) accuseAsMafia(accuserAddress string, accuseeAddress string) error {
	if g.getCurrentPhase() != TimeOfDayDay {
		return errors.New("Mafia accusations can only be made during the day")
	}

	if accuserPlayer := g.getPlayer(accuserAddress); accuserPlayer == nil {
		return fmt.Errorf("accuser '%s' must be a member of game", accuserAddress)
	} else if !accuserPlayer.CanAct() {
		return fmt.Errorf("accuser '%s' must be able to take actions in the game", accuserAddress)
	}

	if accuseePlayer := g.getPlayer(accuseeAddress); accuseePlayer == nil {
		return fmt.Errorf("the accused '%s' must be a member of the game", accuseeAddress)
	} else if !accuseePlayer.CanAct() {
		return fmt.Errorf("the accused '%s' must be able to take actions in the game", accuseeAddress)
	}

	g.mafiaAccusationsMutex.Lock()
	defer g.mafiaAccusationsMutex.Unlock()

	if _, hasAccusation := g.mafiaAccusations[accuserAddress]; hasAccusation {
		return errors.New("a Mafia vote accusation cannot be made twice")
	}

	g.mafiaAccusations[accuserAddress] = accuseeAddress

	return nil
}

func (g *gameState) addPlayer(player *Player) {
	g.playersMutex.Lock()
	defer g.playersMutex.Unlock()

	g.players[player.PlayerAddress] = player
}

func (g *gameState) calculatePhaseOutcome() PhaseOutcome {
	players := g.getPlayers()

	var mafiaPlayers []*Player
	var civvies []*Player
	for _, player := range players {
		if !player.CanAct() {
			continue
		}

		switch player.PlayerRole {
		case PlayerRoleCivilian:
			civvies = append(civvies, player)
		case PlayerRoleMafia:
			mafiaPlayers = append(mafiaPlayers, player)
		}
	}

	if len(civvies) <= len(mafiaPlayers) {
		return PhaseOutcomeMafiaVictory
	} else if len(mafiaPlayers) == 0 {
		return PhaseOutcomeCivilianVictory
	}

	return PhaseOutcomeContinuation
}

// findHighestVote finds the address in the given map that has singly accrued the highest number
// of votes. If one address has a majority of the votes, this returns the address and a bool value of 'true'.
// If there is no clear winner of the votes, this returns false for the bool value.
func (g *gameState) findHighestVote(votes map[string]string) (string, bool) {
	voteCounts := make(map[string]int)
	var highestVoteCount int
	for _, vote := range votes {
		voteCount := voteCounts[vote] + 1
		voteCounts[vote] = voteCount
		if voteCount > highestVoteCount {
			highestVoteCount = voteCount
		}
	}

	var winners []string
	for candidate, count := range voteCounts {
		if count == highestVoteCount {
			winners = append(winners, candidate)
		}
	}

	if len(winners) == 1 {
		return winners[0], true
	}

	return "", false
}

func (g *gameState) getCurrentPhase() TimeOfDay {
	g.currentPhaseMutex.RLock()
	defer g.currentPhaseMutex.RUnlock()

	return g.currentPhase
}

func (i *InMemoryEngine) getGameState(hostAddress string) (*gameState, bool) {
	i.gameStatesMutex.RLock()
	defer i.gameStatesMutex.RUnlock()

	gameState, hasGameState := i.gameStates[hostAddress]
	return gameState, hasGameState
}

func (g *gameState) getPlayer(playerAddress string) *Player {
	g.playersMutex.RLock()
	defer g.playersMutex.RUnlock()

	if player, hasPlayer := g.players[playerAddress]; hasPlayer {
		return player
	}

	return nil
}

func (g *gameState) getPlayers() []*Player {
	g.playersMutex.RLock()
	defer g.playersMutex.RUnlock()

	players := make([]*Player, 0, len(g.players))
	for _, player := range g.players {
		players = append(players, player)
	}
	return players
}

func (g *gameState) notifyOfPhaseExecution(phaseExecution *PhaseExecution) {
	switch phaseExecution.CurrentPhase {
	case TimeOfDayDay:
		defer func() {
			g.mafiaAccusationsMutex.Lock()
			defer g.mafiaAccusationsMutex.Unlock()

			g.mafiaAccusations = make(map[string]string)
			g.currentPhase = TimeOfDayNight
		}()
	case TimeOfDayNight:
		defer func() {
			g.killVotesMutex.Lock()
			defer g.killVotesMutex.Unlock()

			g.killVotes = make(map[string]string)
			g.currentPhase = TimeOfDayDay
		}()
	}

	g.phaseExecutionMutex.Lock()
	defer g.phaseExecutionMutex.Unlock()

	fmt.Printf("Notifying %d users of phase execution\n", len(g.phaseExecutionSubs))

	for _, phaseSub := range g.phaseExecutionSubs {
		phaseSub <- phaseExecution
		close(phaseSub)
	}

	g.phaseExecutionSubs = nil
}

func (g *gameState) subscribeToPhaseExecution() (<-chan *PhaseExecution, error) {
	g.phaseExecutionMutex.Lock()
	defer g.phaseExecutionMutex.Unlock()

	newSub := make(chan *PhaseExecution)
	g.phaseExecutionSubs = append(g.phaseExecutionSubs, newSub)

	fmt.Printf("%d users have subscribed for phase execution\n", len(g.phaseExecutionSubs))

	return newSub, nil
}

func (g *gameState) subscribeToStart() (<-chan any, error) {
	g.gameStartMutex.Lock()
	defer g.gameStartMutex.Unlock()

	if g.started {
		return nil, errors.New("cannot subscribe to game start when it is already started")
	}

	newSub := make(chan any)
	g.gameStartSubs = append(g.gameStartSubs, newSub)

	fmt.Printf("%d users have subscribed for game start\n", len(g.gameStartSubs))

	return newSub, nil
}

func (g *gameState) tallyMafiaVotes() (string, bool) {
	g.mafiaAccusationsMutex.RLock()
	defer g.mafiaAccusationsMutex.RUnlock()

	convictedAddress, isConvicted := g.findHighestVote(g.mafiaAccusations)
	if isConvicted {
		g.getPlayer(convictedAddress).Convicted = true
	}

	return convictedAddress, isConvicted
}

func (g *gameState) tallyKillVotes() (string, bool) {
	g.killVotesMutex.RLock()
	defer g.killVotesMutex.RUnlock()

	killedAddress, isKilled := g.findHighestVote(g.killVotes)
	if isKilled {
		g.getPlayer(killedAddress).Dead = true
	}

	return killedAddress, isKilled
}

func (g *gameState) voteToKill(voterAddress string, victimAddress string) error {
	if g.getCurrentPhase() != TimeOfDayNight {
		return errors.New("Votes to kill can only be made during the night")
	}

	if voterPlayer := g.getPlayer(voterAddress); voterPlayer == nil {
		return errors.New("voter must be a member of game")
	} else if !voterPlayer.CanAct() {
		return errors.New("voter must be able to take actions in the game")
	} else if voterPlayer.PlayerRole != PlayerRoleMafia {
		return errors.New("only members of the Mafia can take actions in the game")
	}

	if victimPlayer := g.getPlayer(victimAddress); victimPlayer == nil {
		return errors.New("the victim must be a member of the game")
	} else if !victimPlayer.CanAct() {
		return errors.New("the victim player must be able to take actions in the game")
	}

	g.killVotesMutex.Lock()
	defer g.killVotesMutex.Unlock()

	if _, hasKillVote := g.killVotes[voterAddress]; hasKillVote {
		return errors.New("a vote to kill cannot be made twice")
	}

	g.killVotes[voterAddress] = victimAddress

	return nil
}

func newPlayer(playerAddress string, playerNickname string) *Player {
	return &Player{
		PlayerAddress:  playerAddress,
		PlayerNickname: playerNickname,
	}
}
