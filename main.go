package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jwwsjlm/utils/xdb"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"strings"
)

type NetWorkLogger struct{}

func main() {
	server := gin.Default()

	//加载xdb
	var dbPath = "ip2region.xdb"
	// 1、从 dbPath 加载整个 xdb 到内存
	cBuff, err := xdb.LoadContentFromFile(dbPath)
	if err != nil {
		fmt.Printf("failed to load content from `%s`: %s\n", dbPath, err)
		return
	}

	// 2、用全局的 cBuff 创建完全基于内存的查询对象。
	Searcher, err := xdb.NewWithBuffer(cBuff)
	if err != nil {
		fmt.Printf("failed to create searcher with content: %s\n", err)
		return
	}
	//创建一个服务

	server.GET("/hello", func(context *gin.Context) {
		context.JSON(200, gin.H{"msg": "hello,world"})

	})

	server.POST("/getapi", func(context *gin.Context) {
		fmt.Println("收到get-api请求")
		data, _ := context.GetRawData()
		var m map[string]interface{}
		_ = json.Unmarshal(data, &m)
		types := m["type"].(string)
		msg := m["msg"].(string)
		fmt.Println(types)

		switch types {
		case "getip":
			region, err := Searcher.SearchByStr(msg) //查询Ip位置
			if err != nil {
				context.JSON(201, gin.H{"msg": "ip未找到"})
			}
			context.JSON(200, gin.H{"msg": region})
		case "geetest":
			geetest, err := geetest(msg)
			fmt.Println(err)
			if err != nil {

				context.JSON(201, gin.H{"msg": "错误啦"})
			}
			context.JSON(200, gin.H{"msg": "data:image/jpeg;base64," + geetest})

		default:
			context.JSON(201, gin.H{"msg": "我不知道你发的什么东东"})
		}

	})

	server.POST("/getimage", func(context *gin.Context) {

		file, err := context.FormFile("image")
		if err != nil {
			context.JSON(201, gin.H{"msg": "获取图像失败"})

			return
		}
		imageFile, err := file.Open()
		if err != nil {
			context.JSON(201, gin.H{"msg": "无法打开图像文件"})
			return
		}
		imageData, err := io.ReadAll(imageFile)
		if err != nil {
			context.JSON(201, gin.H{"msg": "无法读取图像数据"})
			return
		}
		img, _, _ := image.Decode(bytes.NewReader(imageData))
		fmt.Println("处理图片请求!!!!!")
		array := []int{39, 38, 48, 49, 41, 40, 46, 47, 35, 34, 50, 51, 33, 32, 28, 29, 27, 26, 36, 37, 31, 30, 44, 45, 43, 42, 12, 13, 23, 22, 14, 15, 21, 20, 8, 9, 25, 24, 6, 7, 3, 2, 0, 1, 11, 10, 4, 5, 19, 18, 16, 17}

		convertedImage := image.NewRGBA(image.Rect(0, 0, 312, 160))
		//

		for i := 0; i < len(array); i++ {
			c := array[i]%26*12 + 1
			fmt.Println(c)
			u := 0
			db := 0

			if array[i] > 25 {
				u = 80
			} else {
				u = 0
			}

			if i > 25 {
				db = 80
			} else {
				db = 0
			}
			copyImage(convertedImage, img, i%26*10, db, c, u, 10, 80)
			// 使用 array[i] 进行操作
		}
		defer func(imageFile multipart.File) {
			err := imageFile.Close()
			if err != nil {
			}
		}(imageFile)
		buffer := new(bytes.Buffer)
		jpeg.Encode(buffer, convertedImage, nil)
		context.Data(200, "image/jpeg", buffer.Bytes())
	})
	err = server.Run(":8082")
	if err != nil {
		return
	}

}

func geetest(base64 string) (string, error) {
	array := []int{39, 38, 48, 49, 41, 40, 46, 47, 35, 34, 50, 51, 33, 32, 28, 29, 27, 26, 36, 37, 31, 30, 44, 45, 43, 42, 12, 13, 23, 22, 14, 15, 21, 20, 8, 9, 25, 24, 6, 7, 3, 2, 0, 1, 11, 10, 4, 5, 19, 18, 16, 17}

	originalImage, err := base64toimage(base64)
	fmt.Println(err)
	if err != nil {
		// 处理解码错误

		return "", errors.New("error")
	}
	convertedImage := image.NewRGBA(image.Rect(0, 0, 312, 160))
	//

	for i := 0; i < len(array); i++ {
		c := array[i]%26*12 + 1
		fmt.Println(c)
		u := 0
		db := 0

		if array[i] > 25 {
			u = 80
		} else {
			u = 0
		}

		if i > 25 {
			db = 80
		} else {
			db = 0
		}
		copyImage(convertedImage, originalImage, i%26*10, db, c, u, 10, 80)
		// 使用 array[i] 进行操作
	}

	base64Result := imagetobase64(convertedImage)
	return base64Result, nil

}
func copyImage(dest *image.RGBA, src image.Image, dx, dy, sx, sy, width, height int) {

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			convertedPixel := color.RGBAModel.Convert(src.At(x+sx, y+sy))
			dest.SetRGBA(x+dx, y+dy, convertedPixel.(color.RGBA))
		}
	}
}

func base64toimage(base64Str string) (image.Image, error) {
	splitStr := strings.Split(base64Str, ",")
	if len(splitStr) != 2 {
		return nil, fmt.Errorf("base64Str 格式不正确")
	}

	data := splitStr[1]

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("解码失败: %v", err)
	}

	img, err := jpeg.Decode(bytes.NewReader(decoded))
	if err != nil {
		return nil, fmt.Errorf("图片解码失败: %v", err)
	}

	return img, nil
}
func imagetobase64(img image.Image) string {
	// 创建字节缓冲区
	var buff bytes.Buffer

	// 将图像编码为PNG格式，并写入到字节缓冲区
	if err := png.Encode(&buff, img); err != nil {
		// 处理编码错误
		return ""
	}

	// 将字节缓冲区的内容进行Base64编码
	base64Str := base64.StdEncoding.EncodeToString(buff.Bytes())

	return base64Str
}
