//stm: #integration
package itests

import (
	"context"
	"testing"
	"time"

	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/itests/kit"
	"github.com/stretchr/testify/require"
)

func TestNonceIncremental(t *testing.T) {
	//stm: @BLOCKCHAIN_SYNCER_LOAD_GENESIS_001, @BLOCKCHAIN_SYNCER_FETCH_TIPSET_001,
	//stm: @BLOCKCHAIN_SYNCER_START_001, @BLOCKCHAIN_SYNCER_SYNC_001, @BLOCKCHAIN_BEACON_VALIDATE_BLOCK_01
	//stm: @BLOCKCHAIN_SYNCER_COLLECT_CHAIN_001, @BLOCKCHAIN_SYNCER_COLLECT_HEADERS_001, @BLOCKCHAIN_SYNCER_VALIDATE_TIPSET_001
	//stm: @BLOCKCHAIN_SYNCER_NEW_PEER_HEAD_001, @BLOCKCHAIN_SYNCER_VALIDATE_MESSAGE_META_001, @BLOCKCHAIN_SYNCER_STOP_001
	ctx := context.Background()

	kit.QuietMiningLogs()

	client, _, ens := kit.EnsembleMinimal(t, kit.MockProofs())
	ens.InterconnectAll().BeginMining(10 * time.Millisecond)

	// create a new address where to send funds.
	addr, err := client.WalletNew(ctx, types.KTBLS)
	require.NoError(t, err)

	// get the existing balance from the default wallet to then split it.
	bal, err := client.WalletBalance(ctx, client.DefaultKey.Address)
	require.NoError(t, err)

	const iterations = 100

	// we'll send half our balance (saving the other half for gas),
	// in `iterations` increments.
	toSend := big.Div(bal, big.NewInt(2))
	each := big.Div(toSend, big.NewInt(iterations))

	var sms []*types.SignedMessage
	for i := 0; i < iterations; i++ {
		msg := &types.Message{
			From:  client.DefaultKey.Address,
			To:    addr,
			Value: each,
		}

		sm, err := client.MpoolPushMessage(ctx, msg, nil)
		require.NoError(t, err)
		require.EqualValues(t, i, sm.Message.Nonce)

		sms = append(sms, sm)
	}

	for _, sm := range sms {
		_, err := client.StateWaitMsg(ctx, sm.Cid(), 3, api.LookbackNoLimit, true)
		require.NoError(t, err)
	}
}
