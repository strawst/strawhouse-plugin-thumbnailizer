package process

import (
	"bytes"
	"github.com/bsthun/gut"
	"github.com/strawst/strawhouse-go/pb"
	"image/jpeg"
	"log"
	"path/filepath"
	"strawhouse-plugin-thumbnailizer/service/resize"
	"strawhouse-plugin-thumbnailizer/type/iface"
	"strings"
)

func UploadProcessor(c iface.Callbacker, res *pb.UploadFeedResponse) {
	log.Printf("upload processor")
	directories := strings.Split(res.Directory, "/")
	if directories[3] != "upload" {
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
	img, err := jpeg.Decode(bytes.NewReader(content))
	if err != nil {
		gut.Debug("[thumbnailizer] error decoding image", err)
		return
	}

	resized03, er := resize.ResizeImage(img, 300000)
	if er != nil {
		gut.Debug("[thumbnailizer] error resizing image", er)
		return
	}

	resized20, er := resize.ResizeImage(img, 2000000)
	if er != nil {
		gut.Debug("[thumbnailizer] error resizing image", er)
		return
	}

	go func() {
		log.Printf("resized03: %v, resized20: %v", len(resized03), len(resized20))
		reader := bytes.NewReader(resized03)
		_, _, _, er = c.Callback().Upload(res.Name, filepath.Join(strings.Join(directories[0:3], "/"), "tmb03"), reader)
		if er != nil {
			gut.Debug("[thumbnailizer] error uploading thumbnail", er)
			return
		}

		log.Printf("uploading tmb20")
		reader = bytes.NewReader(resized20)
		_, _, _, er = c.Callback().Upload(res.Name, filepath.Join(strings.Join(directories[0:3], "/"), "tmb20"), reader)
		if er != nil {
			gut.Debug("[thumbnailizer] error uploading thumbnail", er)
			return
		}
	}()
}
