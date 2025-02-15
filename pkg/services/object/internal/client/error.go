package internal

import clientcore "github.com/TrueCloudLab/frostfs-node/pkg/core/client"

type errorReporter interface {
	ReportError(error)
}

// ReportError drops client connection if possible.
func ReportError(c clientcore.Client, err error) {
	if ce, ok := c.(errorReporter); ok {
		ce.ReportError(err)
	}
}
