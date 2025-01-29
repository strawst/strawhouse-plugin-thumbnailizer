package main

import (
	"github.com/strawst/strawhouse-go"
	"github.com/strawst/strawhouse-go/pb"
	"strawhouse-plugin-thumbnailizer/service/process"
)

// Plugin provides plugin instance for strawhouse
func Plugin() strawhouse.Plugin {
	return new(Plug)
}

type Plug struct {
	callback strawhouse.PluginCallback
	bindId   uint64
}

func (r *Plug) Load(callback strawhouse.PluginCallback) {
	r.callback = callback
	r.bindId = r.callback.Bind(strawhouse.FeedTypeUpload, "/st/album/", func(resp any) {
		process.UploadProcessor(r, resp.(*pb.UploadFeedResponse))
	})
}

func (r *Plug) Unload() {
	r.callback.Unbind(strawhouse.FeedTypeUpload, "/st/album/", r.bindId)
}

func (r *Plug) Callback() strawhouse.PluginCallback {
	return r.callback
}
