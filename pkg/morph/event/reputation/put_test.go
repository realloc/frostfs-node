package reputation

import (
	"math/big"
	"testing"

	"github.com/TrueCloudLab/frostfs-node/pkg/morph/event"
	"github.com/TrueCloudLab/frostfs-sdk-go/reputation"
	reputationtest "github.com/TrueCloudLab/frostfs-sdk-go/reputation/test"
	"github.com/nspcc-dev/neo-go/pkg/core/state"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/stretchr/testify/require"
)

func TestParsePut(t *testing.T) {
	var (
		peerID = reputationtest.PeerID()

		value      reputation.GlobalTrust
		trust      reputation.Trust
		trustValue float64 = 0.64

		epoch uint64 = 42
	)

	trust.SetValue(trustValue)
	trust.SetPeer(peerID)

	value.SetTrust(trust)

	rawValue := value.Marshal()

	t.Run("wrong number of parameters", func(t *testing.T) {
		prms := []stackitem.Item{
			stackitem.NewMap(),
			stackitem.NewMap(),
		}

		_, err := ParsePut(createNotifyEventFromItems(prms))
		require.EqualError(t, err, event.WrongNumberOfParameters(3, len(prms)).Error())
	})

	t.Run("wrong epoch parameter", func(t *testing.T) {
		_, err := ParsePut(createNotifyEventFromItems([]stackitem.Item{
			stackitem.NewMap(),
		}))

		require.Error(t, err)
	})

	t.Run("wrong peerID parameter", func(t *testing.T) {
		_, err := ParsePut(createNotifyEventFromItems([]stackitem.Item{
			stackitem.NewBigInteger(new(big.Int).SetUint64(epoch)),
			stackitem.NewMap(),
		}))

		require.Error(t, err)
	})

	t.Run("wrong value parameter", func(t *testing.T) {
		_, err := ParsePut(createNotifyEventFromItems([]stackitem.Item{
			stackitem.NewBigInteger(new(big.Int).SetUint64(epoch)),
			stackitem.NewByteArray(peerID.PublicKey()),
			stackitem.NewMap(),
		}))

		require.Error(t, err)
	})

	t.Run("correct behavior", func(t *testing.T) {
		ev, err := ParsePut(createNotifyEventFromItems([]stackitem.Item{
			stackitem.NewBigInteger(new(big.Int).SetUint64(epoch)),
			stackitem.NewByteArray(peerID.PublicKey()),
			stackitem.NewByteArray(rawValue),
		}))
		require.NoError(t, err)

		require.Equal(t, Put{
			epoch:  epoch,
			peerID: peerID,
			value:  value,
		}, ev)
	})
}

func createNotifyEventFromItems(items []stackitem.Item) *state.ContainedNotificationEvent {
	return &state.ContainedNotificationEvent{
		NotificationEvent: state.NotificationEvent{
			Item: stackitem.NewArray(items),
		},
	}
}
