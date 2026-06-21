package main

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/eggsylah/pod-eggsylah-gozxing/babashka"
	"github.com/babashka/transit-go"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

const podName = "pod.eggsylah.gozxing"
func listToSlice(l *list.List) []interface{} {
	slice := make([]interface{}, l.Len())
	i := 0
	for e := l.Front(); e != nil; e = e.Next() {
		slice[i] = e.Value
		i++
	}
	return slice
}

func decodeArgs(args string) ([]interface{}, error) {
	decoder := transit.NewDecoder(strings.NewReader(args))
	value, err := decoder.Decode()
	if err != nil {
		return nil, err
	}
	l, ok := value.(*list.List)
	if !ok {
		return nil, errors.New("expected a vector of arguments")
	}
	return listToSlice(l), nil
}

func respond(message *babashka.Message, response interface{}) {
	buf := bytes.NewBufferString("")
	encoder := transit.NewEncoder(buf, false)
	if err := encoder.Encode(response); err != nil {
		babashka.WriteErrorResponse(message, err)
	} else {
		babashka.WriteInvokeResponse(message, buf.String())
	}
}

// readImage accepts a file path (string) or raw image bytes ([]byte).
func readImage(arg interface{}) (image.Image, error) {
	switch v := arg.(type) {
	case string:
		f, err := os.Open(v)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		return img, err
	case []byte:
		img, _, err := image.Decode(bytes.NewReader(v))
		return img, err
	default:
		return nil, errors.New("decode expects a file path string or image bytes")
	}
}

func qrDecode(arg interface{}) (string, error) {
	img, err := readImage(arg)
	if err != nil {
		return "", err
	}
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", err
	}
	result, err := qrcode.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		return "", err
	}
	return result.GetText(), nil
}

func optSize(opts interface{}) int {
	size := 256
	if m, ok := opts.(map[interface{}]interface{}); ok {
		for k, v := range m {
			if fmt.Sprintf("%v", k) == "size" {
				switch n := v.(type) {
				case int64:
					size = int(n)
				case int:
					size = n
				case float64:
					size = int(n)
				}
			}
		}
	}
	return size
}

func qrEncode(text string, path string, size int) error {
	bmp, err := qrcode.NewQRCodeWriter().Encode(text, gozxing.BarcodeFormat_QR_CODE, size, size, nil)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, bmp)
}

func processMessage(message *babashka.Message) {
	switch message.Op {
	case "describe":
		babashka.WriteDescribeResponse(
			&babashka.DescribeResponse{
				Format: "transit+json",
				Namespaces: []babashka.Namespace{
					{
						Name: podName,
						Vars: []babashka.Var{
							{Name: "decode"},
							{Name: "encode"},
						},
					},
				},
			})
	case "invoke":
		switch message.Var {
		case podName + "/decode":
			args, err := decodeArgs(message.Args)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}
			if len(args) < 1 {
				babashka.WriteErrorResponse(message, errors.New("decode expects 1 argument: a file path or image bytes"))
				return
			}
			text, err := qrDecode(args[0])
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}
			respond(message, text)
		case podName + "/encode":
			args, err := decodeArgs(message.Args)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}
			if len(args) < 2 {
				babashka.WriteErrorResponse(message, errors.New("encode expects at least 2 arguments: text and output path"))
				return
			}
			text, ok := args[0].(string)
			if !ok {
				babashka.WriteErrorResponse(message, errors.New("encode: text must be a string"))
				return
			}
			path, ok := args[1].(string)
			if !ok {
				babashka.WriteErrorResponse(message, errors.New("encode: output path must be a string"))
				return
			}
			size := 256
			if len(args) >= 3 {
				size = optSize(args[2])
			}
			if err := qrEncode(text, path, size); err != nil {
				babashka.WriteErrorResponse(message, err)
				return
			}
			respond(message, path)
		default:
			babashka.WriteErrorResponse(message, fmt.Errorf("Unknown var %s", message.Var))
		}
	default:
		babashka.WriteErrorResponse(message, fmt.Errorf("Unknown op %s", message.Op))
	}
}

func main() {
	for {
		message, err := babashka.ReadMessage()
		if err != nil {
			if err.Error() == "EOF" {
				log.Fatal("Unrecoverable error: EOF")
			}
			fmt.Fprintln(os.Stderr, "Error reading message:", err)
			continue
		}
		processMessage(message)
	}
}
