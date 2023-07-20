package game

import "context"

type Engine interface {
	AccuseAsMafia(ctx context.Context, hostAddress string, accuserAddress string, accuseeAddress string) error
	CancelGame(ctx context.Context, hostAddress string) error
	ExecutePhase(ctx context.Context, hostAddress string) error
	FinishGame(ctx context.Context, hostAddress string) error
	GetPlayer(ctx context.Context, hostAddress string, playerAddress string) (*Player, error)
	GetPlayers(ctx context.Context, hostAddress string) ([]*Player, error)
	InitializeGame(ctx context.Context, hostAddress string) error
	JoinGame(ctx context.Context, hostAddress string, playerNickname string) error
	StartGame(ctx context.Context, hostAddress string) error
	VoteToKill(ctx context.Context, hostAddress string, killerAddress string, killeeAddress string) error
	WaitForGameStart(ctx context.Context, hostAddress string) error
	WaitForPhaseExecution(ctx context.Context, hostAddress string) (*PhaseExecution, error)
}

type PhaseOutcome int

const PhaseOutcomeContinuation PhaseOutcome = 0
const PhaseOutcomeCivilianVictory PhaseOutcome = 1
const PhaseOutcomeMafiaVictory PhaseOutcome = 2

type TimeOfDay int

const TimeOfDayDay TimeOfDay = 0
const TimeOfDayNight TimeOfDay = 1

type PlayerRole int

const PlayerRoleCivilian PlayerRole = 0
const PlayerRoleMafia PlayerRole = 1

type PhaseExecution struct {
	HostAddress      string
	PhaseOutcome     PhaseOutcome
	CurrentPhase     TimeOfDay
	KilledPlayers    []string
	ConvictedPlayers []string
}

type Player struct {
	PlayerAddress  string
	PlayerNickname string
	PlayerRole     PlayerRole
}
