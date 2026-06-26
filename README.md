# pod-eggsylah-gozxing

A [babashka](https://github.com/babashka/babashka) pod for reading and writing
QR codes, backed by the Go library
[gozxing](https://github.com/makiuchi-d/gozxing) (a port of ZXing).

This pod exists because `javax.imageio.ImageIO` / AWT do not work in the
babashka native image (GraalVM does not ship the AWT native libraries, see
[oracle/graal#10951](https://github.com/oracle/graal/issues/10951)). The Go pod
does image decoding and QR processing out of process.

## Usage

```clojure
(require '[babashka.pods :as pods])
(pods/load-pod 'org.babashka/gozxing "0.0.2")
;; or load a local build:
;; (pods/load-pod "./pod-eggsylah-gozxing")
(require '[pod.eggsylah.gozxing :as barcode])

;; encode text into a QR png with different error-correcting levels and sizes
(barcode/encode "https://babashka.org" "out.png")
(barcode/encode "https://babashka.org" "out.png" {:format :QR})
(barcode/encode "https://babashka.org" "out.png" {:ec-level :H})
(barcode/encode "https://babashka.org" "out.png" {:size 512})

;; encode text into a DataMatrix png 
(barcode/encode "https://babashka.org" "out.png" {:format :DataMatrix})

;; encode text into a 1D barcode of different formats and sizes
(barcode/encode "89012345678901234563" "code128.png" {:format :Code128 :size [220 48]})
(barcode/encode "89012345678901234563" "code39.png" {:format :Code39 :size [256 24]})
(barcode/encode "725272730706" "upc-a.png" {:format :UPC-A})

;; decode a QR from a file path
(barcode/decode "out.png") ;;=> "https://babashka.org"

;; decode a QR code from raw image bytes
(require '[babashka.fs :as fs])
(barcode/decode (fs/read-all-bytes "out.png"))
```

### `decode`

`(decode path-or-bytes)` / `(decode path-or-bytes opts)` - reads a QR code from a PNG/JPEG/GIF file path or raw
image bytes. Returns the decoded text. Throws error if no barcode is found.

Options:
- `:format` - barcode format to use, `:DataMatrix`, `:QR`, `:Code-128`, `:Code-39`, `:EAN-13`, `:ITF`, :`UPC-A`  (default `:QR`)

### `encode`

`(encode text path)` / `(encode text path opts)` - writes `text` as specified barcode
PNG to `path`. 

Options:
- `:format` - barcode format to use, `:DataMatrix`, `:QR`, `:Code-128`, `:Code-39`, `:EAN-13`, `:ITF`, :`UPC-A`  (default `:QR`)
- `:size` - width/height in pixels (default `256`)
   For a 1D barcode the width and height of the barcode can be specified as a vector, such as [192 32]
- `:ec-level` - error-correcting level for QR codes `:H` (high, 30%), `:Q` (quartile, 25%), `:M` (medium, 15%), 
   `:L` (low, 7%) width/height in pixels (default `:L`)
   This option is only supported for QR codes.



Returns the output path.

## Build

Build for current platform
```bash
script/build
```

Cross-compile for 64-bit Windows
```bash
script/build-win
```

## Test

```bash
script/test
```

## License

Copyright © 2026 Michiel Borkent

Distributed under the MIT License. See LICENSE.
