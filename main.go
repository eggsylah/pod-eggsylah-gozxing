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
	"math/big"
	"os"
	"strings"

	"github.com/babashka/transit-go"
	"github.com/eggsylah/pod-eggsylah-gozxing/babashka"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/datamatrix"
	"github.com/makiuchi-d/gozxing/oned"
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

func decode(barcodeFormat string, arg interface{}) (string, error) {
	img, err := readImage(arg)
	if err != nil {
		return "", err
	}
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", err
	}
	var result *gozxing.Result

	switch barcodeFormat {
	case ":Code128":
		result, err = oned.NewCode128Reader().Decode(bmp, nil)
	case ":Code39":
		result, err = oned.NewCode39Reader().Decode(bmp, nil)
	case ":QR":
		result, err = qrcode.NewQRCodeReader().Decode(bmp, nil)
	case ":DataMatrix":
		result, err = datamatrix.NewDataMatrixReader().Decode(bmp, nil)
	case ":UPC-A":
		result, err = oned.NewUPCAReader().Decode(bmp, nil)
	default:
		err = errors.New(fmt.Sprintf("Unsupported barcode format for decode: %s", barcodeFormat))
	}
	if err != nil {
		return "", err
	}
	return result.GetText(), nil
}

// convert options into a map using strings as keys
func getOptionsMap(options interface{}) map[string]interface{} {
	ret := make(map[string]interface{})

	if m, ok := options.(map[interface{}]interface{}); ok {
		for k, v := range m {
			n := fmt.Sprintf("%v", k)
			ret[n] = v
		}
	}
	return ret
}

func getOptionValueAsInt(val interface{}) (ret int, err error) {
	if val != nil {
		switch n := val.(type) {
		case int64:
			ret = int(n)
		case int:
			ret = int(n)
		case float64:
			ret = int(n)
		case *big.Int:
			x, _ := n.Float64()
			ret = int(x)
		case big.Rat:
			x, _ := n.Float64()
			ret = int(x)
		default:
			err = errors.New(fmt.Sprintf("value should be a number not %v, a %T", val, val))
		}
	}
	return ret, err
}

func getSize(opts map[string]interface{}) (sizeX int, sizeY int, err error) {
	v := opts[":size"]
	if v != nil {
		switch n := v.(type) {
		case []interface{}:
			nn := v.([]interface{})
			if len(nn) != 2 {
				err = errors.New(fmt.Sprintf("size must be a number or a 2-vector of numbers not %v", nn))
			}
			if err == nil {
				sizeX, err = getOptionValueAsInt(nn[0])
			}
			if err == nil {
				sizeY, err = getOptionValueAsInt(nn[1])
			}

		default:
			var size int
			size, err = getOptionValueAsInt(n)
			sizeX = size
			sizeY = size
		}
	}
	return sizeX, sizeY, err
}

func getECLevel(opts map[string]interface{}) (ecLevel string, err error) {
	ecLevel, err = "", nil

	v := opts[":ec-level"]
	if v != nil {
		val := fmt.Sprintf("%v", v)
		ecLevel = val
		if ecLevel[0] == ':' { // hack so keyword maps to string
			ecLevel = ecLevel[1:]
		}
		if len(ecLevel) > 1 || strings.Index("LMQH", ecLevel) < 0 {
			err = errors.New(fmt.Sprintf("ec-level must be one of L/M/Q/H not %s", val))
		}
	}
	return ecLevel, err
}

func encode(barcodeFormat string, text string, path string, sizeX int, sizeY int, ecLevel string) (err error) {
	var bmp image.Image
	switch barcodeFormat {
	case ":Code128":
		bmp, err = oned.NewCode128Writer().Encode(text, gozxing.BarcodeFormat_CODE_128, sizeX, sizeY, nil)
	case ":Code39":
		bmp, err = oned.NewCode39Writer().Encode(text, gozxing.BarcodeFormat_CODE_39, sizeX, sizeY, nil)
	case ":UPC-A":
		bmp, err = oned.NewUPCAWriter().Encode(text, gozxing.BarcodeFormat_UPC_A, sizeX, sizeY, nil)
	case ":QR":
		hints := map[gozxing.EncodeHintType]interface{}{gozxing.EncodeHintType_ERROR_CORRECTION: ecLevel}
		bmp, err = qrcode.NewQRCodeWriter().Encode(text, gozxing.BarcodeFormat_QR_CODE, sizeX, sizeY, hints)
	case ":DataMatrix":
		bmp, err = datamatrix.NewDataMatrixWriter().Encode(text, gozxing.BarcodeFormat_DATA_MATRIX, sizeX, sizeY, nil)
	default:
		err = errors.New(fmt.Sprintf("Unsupported barcode format for encode: %s", barcodeFormat))
	}
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

func processDecode(message *babashka.Message) {
	var text string
	args, err := decodeArgs(message.Args)
	if err == nil && len(args) < 1 {
		err = errors.New("decode expects 1 argument: a file path or image bytes")
	}

	// defaults ...
	barcodeFormat := ":QR"
	if err == nil && len(args) >= 2 {
		optMap := getOptionsMap(args[1])
		for k, _ := range optMap {
			switch k {
			case ":format":
				barcodeFormat = fmt.Sprintf("%v", optMap[":format"])
			default:
				err = errors.New(fmt.Sprintf("unsupported decode option: %s", k))
				break
			}
		}
	}
	if err == nil {
		text, err = decode(barcodeFormat, args[0])
	}

	if err == nil {
		respond(message, text)
	} else {
		babashka.WriteErrorResponse(message, err)
	}
}

func processEncode(message *babashka.Message) {
	args, err := decodeArgs(message.Args)
	var text, path string
	var ok bool

	if err == nil && len(args) < 2 {
		err = errors.New("encode expects at least 2 arguments: text and output path")
	}
	if err == nil {
		text, ok = args[0].(string)
		if !ok {
			err = errors.New("encode: text must be a string")
		}
	}
	if err == nil {
		path, ok = args[1].(string)
		if !ok {
			err = errors.New("encode: output path must be a string")
		}
	}
	// defaults ...
	barcodeFormat := ":QR"
	sizeX, sizeY := 256, 256
	ecLevel := ""
	if err == nil && len(args) >= 3 {
		optMap := getOptionsMap(args[2])
		for k, _ := range optMap {
			switch k {
			case ":format":
				barcodeFormat = fmt.Sprintf("%v", optMap[":format"])
			case ":size":
				sizeX, sizeY, err = getSize(optMap)
			case ":ec-level":
				ecLevel, err = getECLevel(optMap)
			default:
				err = errors.New(fmt.Sprintf("unsupported encode option: %s", k))
			}
			if err == nil && barcodeFormat != ":QR" && ecLevel != "" {
				err = errors.New("ec-level only supported for QR barcode encoding")
			}
		}
	}
	if ecLevel == "" {
		ecLevel = "L"
	}

	if err == nil {
		err = encode(barcodeFormat, text, path, sizeX, sizeY, ecLevel)
	}
	if err == nil {
		respond(message, path)
	} else {
		babashka.WriteErrorResponse(message, err)
	}
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
			processDecode(message)
		case podName + "/encode":
			processEncode(message)
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
