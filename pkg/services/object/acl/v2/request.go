package v2

import (
	"crypto/ecdsa"
	"fmt"

	sessionV2 "github.com/TrueCloudLab/frostfs-api-go/v2/session"
	"github.com/TrueCloudLab/frostfs-sdk-go/bearer"
	"github.com/TrueCloudLab/frostfs-sdk-go/container/acl"
	cid "github.com/TrueCloudLab/frostfs-sdk-go/container/id"
	oid "github.com/TrueCloudLab/frostfs-sdk-go/object/id"
	sessionSDK "github.com/TrueCloudLab/frostfs-sdk-go/session"
	"github.com/TrueCloudLab/frostfs-sdk-go/user"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
)

// RequestInfo groups parsed version-independent (from SDK library)
// request information and raw API request.
type RequestInfo struct {
	basicACL    acl.Basic
	requestRole acl.Role
	operation   acl.Op  // put, get, head, etc.
	cnrOwner    user.ID // container owner

	idCnr cid.ID

	// optional for some request
	// e.g. Put, Search
	obj *oid.ID

	senderKey []byte

	bearer *bearer.Token // bearer token of request

	srcRequest interface{}
}

func (r *RequestInfo) SetBasicACL(basicACL acl.Basic) {
	r.basicACL = basicACL
}

func (r *RequestInfo) SetRequestRole(requestRole acl.Role) {
	r.requestRole = requestRole
}

func (r *RequestInfo) SetSenderKey(senderKey []byte) {
	r.senderKey = senderKey
}

// Request returns raw API request.
func (r RequestInfo) Request() interface{} {
	return r.srcRequest
}

// ContainerOwner returns owner if the container.
func (r RequestInfo) ContainerOwner() user.ID {
	return r.cnrOwner
}

// ObjectID return object ID.
func (r RequestInfo) ObjectID() *oid.ID {
	return r.obj
}

// ContainerID return container ID.
func (r RequestInfo) ContainerID() cid.ID {
	return r.idCnr
}

// CleanBearer forces cleaning bearer token information.
func (r *RequestInfo) CleanBearer() {
	r.bearer = nil
}

// Bearer returns bearer token of the request.
func (r RequestInfo) Bearer() *bearer.Token {
	return r.bearer
}

// BasicACL returns basic ACL of the container.
func (r RequestInfo) BasicACL() acl.Basic {
	return r.basicACL
}

// SenderKey returns public key of the request's sender.
func (r RequestInfo) SenderKey() []byte {
	return r.senderKey
}

// Operation returns request's operation.
func (r RequestInfo) Operation() acl.Op {
	return r.operation
}

// RequestRole returns request sender's role.
func (r RequestInfo) RequestRole() acl.Role {
	return r.requestRole
}

// MetaWithToken groups session and bearer tokens,
// verification header and raw API request.
type MetaWithToken struct {
	vheader *sessionV2.RequestVerificationHeader
	token   *sessionSDK.Object
	bearer  *bearer.Token
	src     interface{}
}

// RequestOwner returns ownerID and its public key
// according to internal meta information.
func (r MetaWithToken) RequestOwner() (*user.ID, *keys.PublicKey, error) {
	if r.vheader == nil {
		return nil, nil, errEmptyVerificationHeader
	}

	// if session token is presented, use it as truth source
	if r.token != nil {
		// verify signature of session token
		return ownerFromToken(r.token)
	}

	// otherwise get original body signature
	bodySignature := originalBodySignature(r.vheader)
	if bodySignature == nil {
		return nil, nil, errEmptyBodySig
	}

	key, err := unmarshalPublicKey(bodySignature.GetKey())
	if err != nil {
		return nil, nil, fmt.Errorf("invalid key in body signature: %w", err)
	}

	var idSender user.ID
	user.IDFromKey(&idSender, (ecdsa.PublicKey)(*key))

	return &idSender, key, nil
}
