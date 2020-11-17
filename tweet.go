package main

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"gocv.io/x/gocv"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const imagePath = "pics"
const uploadTweetDuration = 100
const skippedFrameOptional = 10

func main(){
	 api := anaconda.NewTwitterApiWithCredentials(
		"your acces token",
		"your acces token secret",
		"your consumer key",
		"your consumer secret",
		)

	vc, err := gocv.OpenVideoCapture("videos/video.mp4")
	defer vc.Close()
	if nil != err {
		log.Fatal("open video capture failed: ", err)
	}


	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		err := os.Mkdir(imagePath, 0764)
		if err != nil {
			log.Fatal(err)
		}
	}

	counter := 1

	images, err := getVideoImageByFrameNumber(vc, counter)
	for ; err == nil; counter++{
		file, _ := os.Create(fmt.Sprintf("pics/%d.jpg", counter))
		err = jpeg.Encode(file,images,nil)
		images, err = getVideoImageByFrameNumber(vc, counter+1)
	}



	files, _ := ioutil.ReadDir("pics")
	for index, file := range files{
		caption := fmt.Sprintf("Kopi dangdut frame %d of %d in order", index+1, len(files))
		text, err := uploadTweetImg(file.Name(), api, caption)
		if err != nil{
			fmt.Println(err)
		}

		fmt.Println(text)
		time.Sleep(uploadTweetDuration * time.Second)
	}

}

func uploadTweetImg(fileName string, api *anaconda.TwitterApi, caption string) (string, error){
	imageEncoded, err := encodedImgToBase64(fmt.Sprintf("pics/%s", fileName))
	media, err := api.UploadMedia(imageEncoded)
	if err != nil {
		return "" , err
	}
	v := url.Values{}

	v.Set("media_ids", media.MediaIDString)
	t, err := api.PostTweet(caption, v)

	if err != nil {
		return "", err
	}

	return t.FullText, nil
}

func getVideoImageByFrameNumber(vc *gocv.VideoCapture, frameNumber int) (image.Image, error) {
	num := float64(frameNumber*skippedFrameOptional)
	vc.Set(gocv.VideoCapturePosFrames, num)

	frame := gocv.NewMat()
	defer frame.Close()

	vc.Read(&frame)

	if frame.Empty() {
		return nil, errors.New("frame out of context")
	}

	return frame.ToImage()
}

func encodedImgToBase64(path string) (string, error){
	if ext := filepath.Ext(path); ext != ".jpg" {
		return "", errors.New("Not image file")
	}

	f, err := os.Open(path)
	if err != nil{
		return "", errors.New("cant open file")
	}

	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)
	encoded := base64.StdEncoding.EncodeToString(content)

	return encoded, nil
}
