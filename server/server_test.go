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
		ctx, cancelFn = context.WithTimeout(context.Background(), time.Minute)
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
		initializeResponse, err := client.R().SetContext(ctx).Post(fmt.Sprintf("%s/game/%s", baseURL, hostAddress))
		Expect(err).ToNot(HaveOccurred(), "initializing the game should not fail")
		Expect(initializeResponse.StatusCode()).To(Equal(http.StatusOK), "the game initialization response should signal success")

		playerAddresses := []string{hostAddress,
			"player0001", "player0002",
			"player0003", "player0004",
			"player0005", "player006",
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
	})
})
