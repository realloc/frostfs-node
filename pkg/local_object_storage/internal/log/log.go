package storagelog

import (
	"github.com/TrueCloudLab/frostfs-node/pkg/util/logger"
	"go.uber.org/zap"
)

// headMsg is a distinctive part of all messages.
const headMsg = "local object storage operation"

// Write writes message about storage engine's operation to logger.
func Write(logger *logger.Logger, fields ...zap.Field) {
	logger.Info(headMsg, fields...)
}

// AddressField returns logger's field for object address.
//
// Address should be type of *object.Address or string.
func AddressField(addr interface{}) zap.Field {
	return zap.Any("address", addr)
}

// OpField returns logger's field for operation type.
func OpField(op string) zap.Field {
	return zap.String("op", op)
}

// StorageTypeField returns logger's field for storage type.
func StorageTypeField(typ string) zap.Field {
	return zap.String("type", typ)
}

// StorageIDField returns logger's field for storage ID.
func StorageIDField(id []byte) zap.Field {
	return zap.String("storage_id", string(id))
}
