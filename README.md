# pod-babashka-gozxing

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
(pods/load-pod 'org.babashka/gozxing "0.0.1")
;; or load a local build:
;; (pods/load-pod "./pod-babashka-gozxing")
(require '[pod.babashka.gozxing :as qr])

;; encode text into a QR png
(qr/encode "https://babashka.org" "out.png")
(qr/encode "https://babashka.org" "out.png" {:size 512})

;; decode a QR from a file path
(qr/decode "out.png") ;;=> "https://babashka.org"

;; decode from raw image bytes
(require '[babashka.fs :as fs])
(qr/decode (fs/read-all-bytes "out.png"))
```

### `decode`

`(decode path-or-bytes)` - reads a QR code from a PNG/JPEG/GIF file path or raw
image bytes. Returns the decoded text. Throws if no QR code is found.

### `encode`

`(encode text path)` / `(encode text path opts)` - writes `text` as a QR code
PNG to `path`. Options:

- `:size` - width/height in pixels (default `256`)

Returns the output path.

## Build

```bash
script/build
```

## Test

```bash
script/test
```

## License

Copyright © 2026 Michiel Borkent

Distributed under the MIT License. See LICENSE.
