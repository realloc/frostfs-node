package balance

import (
	"fmt"

	"github.com/TrueCloudLab/frostfs-node/pkg/morph/client"
	"github.com/TrueCloudLab/frostfs-sdk-go/user"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
)

// TransferPrm groups parameters of TransferX method.
type TransferPrm struct {
	Amount int64

	From, To user.ID

	Details []byte

	client.InvokePrmOptional
}

// TransferX transfers p.Amount of GASe-12 from p.From to p.To
// with details p.Details through direct smart contract call.
//
// If TryNotary is provided, calls notary contract.
func (c *Client) TransferX(p TransferPrm) error {
	from, err := address.StringToUint160(p.From.EncodeToString())
	if err != nil {
		return err
	}

	to, err := address.StringToUint160(p.To.EncodeToString())
	if err != nil {
		return err
	}

	prm := client.InvokePrm{}
	prm.SetMethod(transferXMethod)
	prm.SetArgs(from, to, p.Amount, p.Details)
	prm.InvokePrmOptional = p.InvokePrmOptional

	err = c.client.Invoke(prm)
	if err != nil {
		return fmt.Errorf("could not invoke method (%s): %w", transferXMethod, err)
	}
	return nil
}
