package searchsvc

import (
	"github.com/TrueCloudLab/frostfs-node/pkg/core/client"
	"github.com/TrueCloudLab/frostfs-node/pkg/core/netmap"
	"github.com/TrueCloudLab/frostfs-node/pkg/local_object_storage/engine"
	"github.com/TrueCloudLab/frostfs-node/pkg/services/object/util"
	"github.com/TrueCloudLab/frostfs-node/pkg/services/object_manager/placement"
	"github.com/TrueCloudLab/frostfs-node/pkg/util/logger"
	cid "github.com/TrueCloudLab/frostfs-sdk-go/container/id"
	oid "github.com/TrueCloudLab/frostfs-sdk-go/object/id"
	"go.uber.org/zap"
)

// Service is an utility serving requests
// of Object.Search service.
type Service struct {
	*cfg
}

// Option is a Service's constructor option.
type Option func(*cfg)

type searchClient interface {
	// searchObjects searches objects on the specified node.
	// MUST NOT modify execCtx as it can be accessed concurrently.
	searchObjects(*execCtx, client.NodeInfo) ([]oid.ID, error)
}

type ClientConstructor interface {
	Get(client.NodeInfo) (client.MultiAddressClient, error)
}

type cfg struct {
	log *logger.Logger

	localStorage interface {
		search(*execCtx) ([]oid.ID, error)
	}

	clientConstructor interface {
		get(client.NodeInfo) (searchClient, error)
	}

	traverserGenerator interface {
		generateTraverser(cid.ID, uint64) (*placement.Traverser, error)
	}

	currentEpochReceiver interface {
		currentEpoch() (uint64, error)
	}

	keyStore *util.KeyStorage
}

func defaultCfg() *cfg {
	return &cfg{
		log:               &logger.Logger{Logger: zap.L()},
		clientConstructor: new(clientConstructorWrapper),
	}
}

// New creates, initializes and returns utility serving
// Object.Get service requests.
func New(opts ...Option) *Service {
	c := defaultCfg()

	for i := range opts {
		opts[i](c)
	}

	return &Service{
		cfg: c,
	}
}

// WithLogger returns option to specify Get service's logger.
func WithLogger(l *logger.Logger) Option {
	return func(c *cfg) {
		c.log = &logger.Logger{Logger: l.With(zap.String("component", "Object.Search service"))}
	}
}

// WithLocalStorageEngine returns option to set local storage
// instance.
func WithLocalStorageEngine(e *engine.StorageEngine) Option {
	return func(c *cfg) {
		c.localStorage = &storageEngineWrapper{
			storage: e,
		}
	}
}

// WithClientConstructor returns option to set constructor of remote node clients.
func WithClientConstructor(v ClientConstructor) Option {
	return func(c *cfg) {
		c.clientConstructor.(*clientConstructorWrapper).constructor = v
	}
}

// WithTraverserGenerator returns option to set generator of
// placement traverser to get the objects from containers.
func WithTraverserGenerator(t *util.TraverserGenerator) Option {
	return func(c *cfg) {
		c.traverserGenerator = (*traverseGeneratorWrapper)(t)
	}
}

// WithNetMapSource returns option to set network
// map storage to receive current network state.
func WithNetMapSource(nmSrc netmap.Source) Option {
	return func(c *cfg) {
		c.currentEpochReceiver = &nmSrcWrapper{
			nmSrc: nmSrc,
		}
	}
}

// WithKeyStorage returns option to set private
// key storage for session tokens and node key.
func WithKeyStorage(store *util.KeyStorage) Option {
	return func(c *cfg) {
		c.keyStore = store
	}
}
