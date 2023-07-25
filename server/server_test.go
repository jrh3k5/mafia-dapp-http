package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jrh3k5/mafia-dapp-http/server"
)

var _ = Describe("Server", func() {
	var ctx context.Context
	var client resty.Client
	var baseURL string

	BeforeEach(func() {
		var cancelFn context.CancelFunc
		ctx, cancelFn = context.WithTimeout(context.Background(), 10*time.Second)
		DeferCleanup(cancelFn)

		gameHandler := server.NewServer()
		httpServer := &http.Server{
			Addr:    "0.0.0.0:3000",
			Handler: gameHandler,
		}
		baseURL = "http://localhost:3000"
		go func() {
			_ = httpServer.ListenAndServe()
		}()
		DeferCleanup(func() {
			httpServer.Shutdown(ctx)
		})

		client = *resty.New()
	})

	It("successfully plays an eight-person game", func() {
		hostAddress := "gamehost"
		executePhase := func() {
			phaseExecutionURL := fmt.Sprintf("%s/game/%s/phase/execute", baseURL, hostAddress)
			executeResponse, err := client.R().SetContext(ctx).Post(phaseExecutionURL)
			Expect(err).ToNot(HaveOccurred(), "executing the phase should not fail")
			Expect(executeResponse.StatusCode()).To(Equal(http.StatusOK), "unexpected status code executing phase")
		}

		initializeResponse, err := client.R().SetContext(ctx).Post(fmt.Sprintf("%s/game/%s", baseURL, hostAddress))
		Expect(err).ToNot(HaveOccurred(), "initializing the game should not fail")
		Expect(initializeResponse.StatusCode()).To(Equal(http.StatusOK), "the game initialization response should signal success")

		playerAddresses := []string{hostAddress,
			"player0001", "player0002",
			"player0003", "player0004",
			"player0005", "player0006",
			"player0007"}

		joinWaitGroup := &sync.WaitGroup{}
		joinWaitGroup.Add(len(playerAddresses))

		startWaitGroup := &sync.WaitGroup{}
		startWaitGroup.Add(len(playerAddresses))

		for _, playerAddress := range playerAddresses {
			joiningAddress := playerAddress
			go func() {
				defer GinkgoRecover()

				func() {
					defer joinWaitGroup.Done()

					fmt.Printf("Joining as '%s'\n", joiningAddress)

					joinResponse, err := client.R().SetContext(ctx).Post(fmt.Sprintf("%s/game/%s/join?playerAddress=%s&playerNickname=%s", baseURL, hostAddress, joiningAddress, fmt.Sprintf("%sNick", joiningAddress)))
					Expect(err).ToNot(HaveOccurred(), "%s joining game should not fail", joiningAddress)
					Expect(joinResponse.StatusCode()).To(Equal(http.StatusOK), "unexpected status code when player '%s' joined game", joiningAddress)
				}()

				func() {
					defer startWaitGroup.Done()

					waitResponse, err := client.R().SetContext(ctx).Get(fmt.Sprintf("%s/game/%s/start/wait", baseURL, hostAddress))
					Expect(err).ToNot(HaveOccurred(), "waiting for game to start should not have failed")
					Expect(waitResponse.StatusCode()).To(Equal(http.StatusOK), "unexpected status code while waiting for game to start")
				}()
			}()
		}

		joinWaitGroup.Wait()

		fmt.Println("All players joined; starting game")

		startResponse, err := client.R().SetContext(ctx).Post(fmt.Sprintf("%s/game/%s/start", baseURL, hostAddress))
		Expect(err).ToNot(HaveOccurred(), "starting the game should not fail")
		Expect(startResponse.StatusCode()).To(Equal(http.StatusOK), "unexpected response to starting game")

		startWaitGroup.Wait()

		fmt.Println("Waiting for all players to be notified of game start")

		var civilianAddresses []string
		var mafiaPlayers []string
		for _, playerAddress := range playerAddresses {
			playerInfoResponse, err := client.R().SetContext(ctx).Get(fmt.Sprintf("%s/game/%s/players/%s", baseURL, hostAddress, playerAddress))
			Expect(err).ToNot(HaveOccurred(), "getting info for player '%s' should not have failed", playerAddress)
			var infoResponse map[string]any
			Expect(json.Unmarshal(playerInfoResponse.Body(), &infoResponse)).ToNot(HaveOccurred(), "unmarshalling the player '%s' info response from JSON should not fail", playerAddress)
			Expect(infoResponse).To(And(HaveLen(3), HaveKey("playerAddress"), HaveKey("playerNickname"), HaveKey("playerRole")), "the expected player for player '%s' information must be present", playerAddress)
			playerRole := infoResponse["playerRole"]
			Expect(playerRole).To(Or(Equal(float64(0)), Equal(float64(1))), "the player role must be of an expected value")
			switch infoResponse["playerRole"] {
			case float64(0):
				civilianAddresses = append(civilianAddresses, playerAddress)
			case float64(1):
				mafiaPlayers = append(mafiaPlayers, playerAddress)
			}
		}

		Expect(civilianAddresses).To(HaveLen(6), "there should be six civilians")
		Expect(mafiaPlayers).To(HaveLen(2), "there should have been two Mafia members assigned")

		// Game will progress as:
		// - round 0: civ[5] gets convicted of being Mafia
		// - round 1: civ[4] is killed (oh, no, three civvies versus two Mafia - on the verge of failure!)
		// - round 2: mafia[1] is convicted as being mafia
		// - round 3: civ[3] is killed
		// - round 4: mafia[0] is convicted of being mafia - civilian victory!

		// round 0
		round0OutcomeChannel := make(chan *phaseExecutionResponse)
		accuseAsMafia(ctx, client, baseURL, hostAddress, civilianAddresses[0:5], civilianAddresses[5], round0OutcomeChannel)
		accuseAsMafia(ctx, client, baseURL, hostAddress, mafiaPlayers, civilianAddresses[5], round0OutcomeChannel)
		accuseAsMafia(ctx, client, baseURL, hostAddress, []string{civilianAddresses[5]}, mafiaPlayers[0], round0OutcomeChannel)

		// Wait for all of the reuqests to subscribe and wait to be waiting
		time.Sleep(250 * time.Millisecond)

		executePhase()

		waitForOutcomes(ctx, round0OutcomeChannel, len(playerAddresses), 0, 0, nil, []string{civilianAddresses[5]})

		// round 1
		fmt.Println("Executing round 1")
		round1OutcomeChannel := make(chan *phaseExecutionResponse)
		voteToKill(ctx, client, baseURL, hostAddress, mafiaPlayers, civilianAddresses[4], round1OutcomeChannel)

		// Wait for all of the reuqests to subscribe and wait to be waiting
		time.Sleep(250 * time.Millisecond)

		executePhase()

		waitForOutcomes(ctx, round1OutcomeChannel, len(mafiaPlayers), 0, 1, []string{civilianAddresses[4]}, nil)

		// round 2
		fmt.Println("Executing round 2")
		round2OutcomeChannel := make(chan *phaseExecutionResponse)
		accuseAsMafia(ctx, client, baseURL, hostAddress, civilianAddresses[0:4], mafiaPlayers[1], round2OutcomeChannel)
		accuseAsMafia(ctx, client, baseURL, hostAddress, mafiaPlayers, civilianAddresses[3], round2OutcomeChannel)

		// Wait for all of the reuqests to subscribe and wait to be waiting
		time.Sleep(250 * time.Millisecond)

		executePhase()

		waitForOutcomes(ctx, round2OutcomeChannel, 4+2, 0, 0, nil, []string{mafiaPlayers[1]})

		// round 3
		fmt.Println("Executing round 3")
		round3OutcomeChannel := make(chan *phaseExecutionResponse)
		voteToKill(ctx, client, baseURL, hostAddress, []string{mafiaPlayers[0]}, civilianAddresses[3], round3OutcomeChannel)

		// Wait for all of the reuqests to subscribe and wait to be waiting
		time.Sleep(250 * time.Millisecond)

		executePhase()

		waitForOutcomes(ctx, round3OutcomeChannel, 1, 0, 1, []string{civilianAddresses[3]}, nil)

		// round 4
		fmt.Println("Executing round 4")
		round4OutcomeChannel := make(chan *phaseExecutionResponse)
		accuseAsMafia(ctx, client, baseURL, hostAddress, civilianAddresses[0:2], mafiaPlayers[0], round4OutcomeChannel)
		accuseAsMafia(ctx, client, baseURL, hostAddress, []string{mafiaPlayers[0]}, civilianAddresses[0], round4OutcomeChannel)

		// Wait for all of the reuqests to subscribe and wait to be waiting
		time.Sleep(250 * time.Millisecond)

		executePhase()

		waitForOutcomes(ctx, round4OutcomeChannel, 2+1, 1, 0, nil, []string{mafiaPlayers[0]})
	})
})

func accuseAsMafia(ctx context.Context, client resty.Client, baseURL string, hostAddress string, accuserAddresses []string, accusedAddress string, phaseExecutionChan chan<- *phaseExecutionResponse) {
	for _, accuserAddress := range accuserAddresses {
		voteURL := fmt.Sprintf("%s/game/%s/players/%s/vote/accuse?playerAddress=%s", baseURL, hostAddress, accuserAddress, accusedAddress)
		voteResponse, err := client.R().SetContext(ctx).Post(voteURL)
		Expect(err).ToNot(HaveOccurred(), "failed to accuse '%s' of being mafia on behalf of user '%s'", accusedAddress, accuserAddress)
		Expect(voteResponse.StatusCode()).To(Equal(http.StatusOK), "unexpected response status code when accusing '%s' of being mafia on behalf of user '%s'; response body was '%s'", accusedAddress, accuserAddress, string(voteResponse.Body()))

		sourceAddress := accuserAddress

		// Now wait for the phase to finish executing
		go func() {
			defer GinkgoRecover()

			waitURL := fmt.Sprintf("%s/game/%s/phase/wait", baseURL, hostAddress)
			waitResponse, err := client.R().SetContext(ctx).Get(waitURL)
			Expect(err).ToNot(HaveOccurred(), "waiting for phase execution on behalf of '%s' failed", sourceAddress)
			Expect(waitResponse.StatusCode()).To(Equal(http.StatusOK), "unexpected status code waiting for phase execution on behalf of '%s'; response body is: %s", sourceAddress, string(waitResponse.Body()))
			var phaseExecution *phaseExecutionResponse
			Expect(json.Unmarshal(waitResponse.Body(), &phaseExecution)).ToNot(HaveOccurred(), "failed to unmarshal response for phase execution waiting")
			phaseExecutionChan <- phaseExecution
		}()
	}
}

func voteToKill(ctx context.Context, client resty.Client, baseURL string, hostAddress string, voterAddresses []string, victimAddress string, phaseExecutionChan chan<- *phaseExecutionResponse) {
	for _, voterAddress := range voterAddresses {
		voteURL := fmt.Sprintf("%s/game/%s/players/%s/vote/kill?playerAddress=%s", baseURL, hostAddress, voterAddress, victimAddress)
		voteResponse, err := client.R().SetContext(ctx).Post(voteURL)
		Expect(err).ToNot(HaveOccurred(), "failed to vote to kill '%s' on behalf of user '%s'", voterAddress, victimAddress)
		Expect(voteResponse.StatusCode()).To(Equal(http.StatusOK), "unexpected response status code when voting to kill '%s' on behalf of user '%s'; response body is '%s'", voterAddress, victimAddress, string(voteResponse.Body()))

		sourceAddress := voterAddress

		// Now wait for the phase to finish executing
		go func() {
			defer GinkgoRecover()

			waitURL := fmt.Sprintf("%s/game/%s/phase/wait", baseURL, hostAddress)
			waitResponse, err := client.R().SetContext(ctx).Get(waitURL)
			Expect(err).ToNot(HaveOccurred(), "waiting for phase execution on behalf of '%s' failed", sourceAddress)
			Expect(waitResponse.StatusCode()).To(Equal(http.StatusOK), "unexpected status code waiting for phase execution on behalf of '%s'; response body is: '%s'", sourceAddress, string(waitResponse.Body()))
			var phaseExecution *phaseExecutionResponse
			Expect(json.Unmarshal(waitResponse.Body(), &phaseExecution)).ToNot(HaveOccurred(), "failed to unmarshal response for phase execution waiting")
			phaseExecutionChan <- phaseExecution
		}()
	}
}

func waitForOutcomes(ctx context.Context, outcomeChannel <-chan *phaseExecutionResponse, expectedCount int, expectedOutcome int, expectedPhase int, expectedKilledPlayers []string, expectedConvictedPlayers []string) {
	var outcomeCount int
	for {
		if outcomeCount == expectedCount {
			return
		}

		select {
		case <-ctx.Done():
			Expect(ctx.Err()).ToNot(HaveOccurred(), "test should not have timed out (having seen %d outcomes)", outcomeCount)
		case phaseOutcome := <-outcomeChannel:
			outcomeCount++
			Expect(phaseOutcome.PhaseOutcome).To(Equal(expectedOutcome), "unexpected phase outcome")
			Expect(phaseOutcome.CurrentPhase).To(Equal(expectedPhase), "unexpected current phase")
			Expect(phaseOutcome.ConvictedPlayers).To(Equal(expectedConvictedPlayers), "unexpected convicted players")
			Expect(phaseOutcome.KilledPlayers).To(Equal(expectedKilledPlayers), "unexpected killed players")
		}
	}
}

type phaseExecutionResponse struct {
	PhaseOutcome     int      `json:"phaseOutcome"`
	CurrentPhase     int      `json:"currentPhase"`
	KilledPlayers    []string `json:"killedPlayers"`
	ConvictedPlayers []string `json:"convictedPlayers"`
}
