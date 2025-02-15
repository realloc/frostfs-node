package compression

import (
	"bytes"
	"strings"

	objectSDK "github.com/TrueCloudLab/frostfs-sdk-go/object"
	"github.com/klauspost/compress/zstd"
)

// Config represents common compression-related configuration.
type Config struct {
	Enabled                    bool
	UncompressableContentTypes []string

	encoder *zstd.Encoder
	decoder *zstd.Decoder
}

// zstdFrameMagic contains first 4 bytes of any compressed object
// https://github.com/klauspost/compress/blob/master/zstd/framedec.go#L58 .
var zstdFrameMagic = []byte{0x28, 0xb5, 0x2f, 0xfd}

// Init initializes compression routines.
func (c *Config) Init() error {
	var err error

	if c.Enabled {
		c.encoder, err = zstd.NewWriter(nil)
		if err != nil {
			return err
		}
	}

	c.decoder, err = zstd.NewReader(nil)
	if err != nil {
		return err
	}

	return nil
}

// NeedsCompression returns true if the object should be compressed.
// For an object to be compressed 2 conditions must hold:
// 1. Compression is enabled in settings.
// 2. Object MIME Content-Type is allowed for compression.
func (c *Config) NeedsCompression(obj *objectSDK.Object) bool {
	if !c.Enabled || len(c.UncompressableContentTypes) == 0 {
		return c.Enabled
	}

	for _, attr := range obj.Attributes() {
		if attr.Key() == objectSDK.AttributeContentType {
			for _, value := range c.UncompressableContentTypes {
				match := false
				switch {
				case len(value) > 0 && value[len(value)-1] == '*':
					match = strings.HasPrefix(attr.Value(), value[:len(value)-1])
				case len(value) > 0 && value[0] == '*':
					match = strings.HasSuffix(attr.Value(), value[1:])
				default:
					match = attr.Value() == value
				}
				if match {
					return false
				}
			}
		}
	}

	return c.Enabled
}

// Decompress decompresses data if it starts with the magic
// and returns data untouched otherwise.
func (c *Config) Decompress(data []byte) ([]byte, error) {
	if len(data) < 4 || !bytes.Equal(data[:4], zstdFrameMagic) {
		return data, nil
	}
	return c.decoder.DecodeAll(data, nil)
}

// Compress compresses data if compression is enabled
// and returns data untouched otherwise.
func (c *Config) Compress(data []byte) []byte {
	if c == nil || !c.Enabled {
		return data
	}
	maxSize := c.encoder.MaxEncodedSize(len(data))
	return c.encoder.EncodeAll(data, make([]byte, 0, maxSize))
}

// Close closes encoder and decoder, returns any error occurred.
func (c *Config) Close() error {
	var err error
	if c.encoder != nil {
		err = c.encoder.Close()
	}
	if c.decoder != nil {
		c.decoder.Close()
	}
	return err
}
