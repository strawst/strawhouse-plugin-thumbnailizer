package process

import (
	"bytes"
	"github.com/bsthun/gut"
	"github.com/strawst/strawhouse-go/pb"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"strawhouse-plugin-thumbnailizer/service/resize"
	"strawhouse-plugin-thumbnailizer/type/iface"
	"strings"
)

func UploadProcessor(c iface.Callbacker, res *pb.UploadFeedResponse) {
	directories := strings.Split(res.Directory, "/")
	if directories[4] != "upload" {
		return
	}

	gut.Debug("[thumbnailizer] processing", res.Name)
	writer := new(bytes.Buffer)
	if err := c.Callback().Get(filepath.Join(res.Directory, res.Name), writer); err != nil {
		gut.Debug("error getting file", err, err.Error())
		return
	}
	content := writer.Bytes()

	// Decode the JPEG image
	img, _, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		gut.Debug("[thumbnailizer] error decoding image", err)
		return
	}

	resized03, er := resize.ResizeImage(img, 300000, 70)
	if er != nil {
		gut.Debug("[thumbnailizer] error resizing image", er)
		return
	}

	resized20, er := resize.ResizeImage(img, 2000000, 80)
	if er != nil {
		gut.Debug("[thumbnailizer] error resizing image", er)
		return
	}

	resized80, er := resize.ResizeImage(img, 8000000, 90)
	if er != nil {
		gut.Debug("[thumbnailizer] error resizing image", er)
		return
	}

	go func() {
		res.Name = strings.TrimSuffix(res.Name, filepath.Ext(res.Name)) + ".webp"

		reader := bytes.NewReader(resized03)
		_, _, _, er = c.Callback().Upload(res.Name, filepath.Join(strings.Join(directories[0:4], "/"), "tmb03"), reader)
		if er != nil {
			gut.Debug("[thumbnailizer] error uploading thumbnail", er)
			return
		}

		reader = bytes.NewReader(resized20)
		_, _, _, er = c.Callback().Upload(res.Name, filepath.Join(strings.Join(directories[0:4], "/"), "tmb20"), reader)
		if er != nil {
			gut.Debug("[thumbnailizer] error uploading thumbnail", er)
			return
		}

		reader = bytes.NewReader(resized80)
		_, _, _, er = c.Callback().Upload(res.Name, filepath.Join(strings.Join(directories[0:4], "/"), "tmb80"), reader)
		if er != nil {
			gut.Debug("[thumbnailizer] error uploading thumbnail", er)
			return
		}
	}()
}
