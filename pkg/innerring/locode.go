package innerring

import (
	"github.com/TrueCloudLab/frostfs-node/pkg/innerring/processors/netmap"
	irlocode "github.com/TrueCloudLab/frostfs-node/pkg/innerring/processors/netmap/nodevalidation/locode"
	"github.com/TrueCloudLab/frostfs-node/pkg/util/locode"
	locodedb "github.com/TrueCloudLab/frostfs-node/pkg/util/locode/db"
	locodebolt "github.com/TrueCloudLab/frostfs-node/pkg/util/locode/db/boltdb"
	"github.com/spf13/viper"
)

func (s *Server) newLocodeValidator(cfg *viper.Viper) (netmap.NodeValidator, error) {
	locodeDB := locodebolt.New(locodebolt.Prm{
		Path: cfg.GetString("locode.db.path"),
	},
		locodebolt.ReadOnly(),
	)

	s.registerStarter(locodeDB.Open)
	s.registerIOCloser(locodeDB)

	return irlocode.New(irlocode.Prm{
		DB: (*locodeBoltDBWrapper)(locodeDB),
	}), nil
}

type locodeBoltEntryWrapper struct {
	*locodedb.Key
	*locodedb.Record
}

func (l *locodeBoltEntryWrapper) LocationName() string {
	return l.Record.LocationName()
}

type locodeBoltDBWrapper locodebolt.DB

func (l *locodeBoltDBWrapper) Get(lc *locode.LOCODE) (irlocode.Record, error) {
	key, err := locodedb.NewKey(*lc)
	if err != nil {
		return nil, err
	}

	rec, err := (*locodebolt.DB)(l).Get(*key)
	if err != nil {
		return nil, err
	}

	return &locodeBoltEntryWrapper{
		Key:    key,
		Record: rec,
	}, nil
}
