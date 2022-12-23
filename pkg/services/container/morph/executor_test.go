package container_test

import (
	"context"
	"testing"

	"github.com/TrueCloudLab/frostfs-api-go/v2/container"
	"github.com/TrueCloudLab/frostfs-api-go/v2/refs"
	"github.com/TrueCloudLab/frostfs-api-go/v2/session"
	containerCore "github.com/TrueCloudLab/frostfs-node/pkg/core/container"
	containerSvc "github.com/TrueCloudLab/frostfs-node/pkg/services/container"
	containerSvcMorph "github.com/TrueCloudLab/frostfs-node/pkg/services/container/morph"
	cid "github.com/TrueCloudLab/frostfs-sdk-go/container/id"
	cidtest "github.com/TrueCloudLab/frostfs-sdk-go/container/id/test"
	containertest "github.com/TrueCloudLab/frostfs-sdk-go/container/test"
	frostfscrypto "github.com/TrueCloudLab/frostfs-sdk-go/crypto"
	frostfsecdsa "github.com/TrueCloudLab/frostfs-sdk-go/crypto/ecdsa"
	sessiontest "github.com/TrueCloudLab/frostfs-sdk-go/session/test"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/stretchr/testify/require"
)

type mock struct {
	containerSvcMorph.Reader
}

func (m mock) Put(_ containerCore.Container) (*cid.ID, error) {
	return new(cid.ID), nil
}

func (m mock) Delete(_ containerCore.RemovalWitness) error {
	return nil
}

func (m mock) PutEACL(_ containerCore.EACL) error {
	return nil
}

func TestInvalidToken(t *testing.T) {
	m := mock{}
	e := containerSvcMorph.NewExecutor(m, m)

	cnr := cidtest.ID()

	var cnrV2 refs.ContainerID
	cnr.WriteToV2(&cnrV2)

	priv, err := keys.NewPrivateKey()
	require.NoError(t, err)

	sign := func(reqBody interface {
		StableMarshal([]byte) []byte
		SetSignature(signature *refs.Signature)
	}) {
		signer := frostfsecdsa.Signer(priv.PrivateKey)
		var sig frostfscrypto.Signature
		require.NoError(t, sig.Calculate(signer, reqBody.StableMarshal(nil)))

		var sigV2 refs.Signature
		sig.WriteToV2(&sigV2)
		reqBody.SetSignature(&sigV2)
	}

	var tokV2 session.Token
	sessiontest.ContainerSigned().WriteToV2(&tokV2)

	tests := []struct {
		name string
		op   func(e containerSvc.ServiceExecutor, tokV2 *session.Token) error
	}{
		{
			name: "put",
			op: func(e containerSvc.ServiceExecutor, tokV2 *session.Token) (err error) {
				var reqBody container.PutRequestBody

				cnr := containertest.Container()

				var cnrV2 container.Container
				cnr.WriteToV2(&cnrV2)

				reqBody.SetContainer(&cnrV2)
				sign(&reqBody)

				_, err = e.Put(context.TODO(), tokV2, &reqBody)
				return
			},
		},
		{
			name: "delete",
			op: func(e containerSvc.ServiceExecutor, tokV2 *session.Token) (err error) {
				var reqBody container.DeleteRequestBody
				reqBody.SetContainerID(&cnrV2)

				_, err = e.Delete(context.TODO(), tokV2, &reqBody)
				return
			},
		},
		{
			name: "setEACL",
			op: func(e containerSvc.ServiceExecutor, tokV2 *session.Token) (err error) {
				var reqBody container.SetExtendedACLRequestBody
				reqBody.SetSignature(new(refs.Signature))
				sign(&reqBody)

				_, err = e.SetExtendedACL(context.TODO(), tokV2, &reqBody)
				return
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tok := generateToken(new(session.ObjectSessionContext))
			require.Error(t, test.op(e, tok))

			require.NoError(t, test.op(e, &tokV2))

			require.NoError(t, test.op(e, nil))
		})
	}
}

func generateToken(ctx session.TokenContext) *session.Token {
	body := new(session.TokenBody)
	body.SetContext(ctx)

	tok := new(session.Token)
	tok.SetBody(body)

	return tok
}
