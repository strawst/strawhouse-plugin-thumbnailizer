package iface

import "github.com/strawst/strawhouse-go"

type Callbacker interface {
	Callback() strawhouse.PluginCallback
}
