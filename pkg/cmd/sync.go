package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/c9s/bbgo/pkg/bbgo"
)

func init() {
	SyncCmd.Flags().String("session", "", "the exchange session name for sync")
	SyncCmd.Flags().String("symbol", "BTCUSDT", "trading symbol")
	SyncCmd.Flags().String("since", "", "sync from time")
	RootCmd.AddCommand(SyncCmd)
}

var SyncCmd = &cobra.Command{
	Use:          "sync",
	Short:        "sync trades, orders",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		if len(configFile) == 0 {
			return errors.New("--config option is required")
		}

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return err
		}

		userConfig, err := bbgo.Load(configFile, false)
		if err != nil {
			return err
		}

		since, err := cmd.Flags().GetString("since")
		if err != nil {
			return err
		}

		environ := bbgo.NewEnvironment()
		if err := configureDB(ctx, environ) ; err != nil {
			return err
		}

		if err := environ.AddExchangesFromConfig(userConfig); err != nil {
			return err
		}

		var (
			// default start time
			startTime = time.Now().AddDate(0, -3, 0)
		)

		if len(since) > 0 {
			loc, err := time.LoadLocation("Asia/Taipei")
			if err != nil {
				return err
			}

			startTime, err = time.ParseInLocation("2006-01-02", since, loc)
			if err != nil {
				return err
			}
		}

		sessionName, err := cmd.Flags().GetString("session")
		if err != nil {
			return err
		}

		symbol, err := cmd.Flags().GetString("symbol")
		if err != nil {
			return err
		}

		if len(sessionName) > 0 {
			session, ok := environ.Session(sessionName)
			if !ok {
				return fmt.Errorf("session %s not found", sessionName)
			}

			return syncSession(ctx, environ, session, symbol, startTime)
		}

		for _, session := range environ.Sessions() {
			if err := syncSession(ctx, environ, session, symbol, startTime); err != nil {
				return err
			}
		}

		return nil
	},
}

func syncSession(ctx context.Context, environ *bbgo.Environment, session *bbgo.ExchangeSession, symbol string, startTime time.Time) error {
	log.Infof("starting syncing exchange session %s", session.Name)

	if session.IsolatedMargin {
		log.Infof("session is configured as isolated margin session, using isolated margin symbol %s instead of %s", session.IsolatedMarginSymbol, symbol)
		symbol = session.IsolatedMarginSymbol
	}

	log.Infof("syncing trades from exchange session %s...", session.Name)
	if err := environ.TradeSync.SyncTrades(ctx, session.Exchange, symbol, startTime); err != nil {
		return err
	}

	log.Infof("syncing orders from exchange session %s...", session.Name)
	if err := environ.TradeSync.SyncOrders(ctx, session.Exchange, symbol, startTime); err != nil {
		return err
	}

	log.Infof("exchange session %s synchronization done", session.Name)

	return nil
}
