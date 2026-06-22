#!/usr/bin/env bb

(ns script
  (:require [babashka.fs :as fs]
            [babashka.pods :as pods]
            [babashka.process :as process]
            [clojure.string :as string]
            [clojure.test :as t :refer [deftest is]]))

(pods/load-pod "./pod-eggsylah-gozxing")

(require '[pod.eggsylah.gozxing :as qr])

(def tmp "test/out.png")

(deftest round-trip-qr-test
  (let [text "https://babashka.org"]
    (qr/encode text tmp {:size 256})
    (is (fs/exists? tmp))
    (is (= text (qr/decode tmp)))))

(deftest decode-bytes-test
  (let [text "from-bytes"]
    (qr/encode text tmp)
    (is (= text (qr/decode (fs/read-all-bytes tmp))))))

(deftest round-trip-datamatrix-test
  (let [text "https://babashka.org"]
    (qr/encode text tmp {:format :DataMatrix :size 256})
    (is (fs/exists? tmp))
    (is (= text (qr/decode tmp {:format :DataMatrix})))))

(deftest round-trip-code128-test
  (let [text "https://babashka.org"]
    (qr/encode text tmp {:format :Code128 :size 256})
    (is (fs/exists? tmp))
    (is (= text (qr/decode tmp {:format :Code128})))))

(deftest round-trip-code39-test
  (let [text "8901234567890123456"]
    (qr/encode text tmp {:format :Code39 :size 256})
    (is (fs/exists? tmp))
    (is (= text (qr/decode tmp {:format :Code39})))))

(deftest round-trip-upca-test
  (let [text "725272730706"]
    (qr/encode text tmp {:format :UPC-A :size 256})
    (is (fs/exists? tmp))
    (is (= text (qr/decode tmp {:format :UPC-A})))))

(deftest size-option-test
  (let [text "725272730706"]
    (qr/encode text tmp {:format :UPC-A :size 512})
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)
    (qr/encode text tmp {:format :UPC-A :size 512.0})   ;float
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)
    (qr/encode text tmp {:format :UPC-A :size 512N})   ;big int
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)
    (qr/encode text tmp {:format :UPC-A :size (/ 1024 3)})   ;rational
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)
    (qr/encode text tmp {:format :UPC-A :size [368 32]})   ;vector
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)

    (is (thrown-with-msg? Exception #"value should be a number not bad, a string"
                          (qr/encode text tmp {:format :UPC-A :size "bad"})))

    (is (thrown-with-msg? Exception #"value should be a number not :T, a transit.Keyword"
                          (qr/encode text tmp {:format :UPC-A :size :T})))

    (is (thrown-with-msg? Exception #"size must be a number or a 2-vector of numbers not \[1]"
                          (qr/encode text tmp {:format :UPC-A :size [1]})))

    (is (thrown-with-msg? Exception #"size must be a number or a 2-vector of numbers not \[200 201 202]"
                          (qr/encode text tmp {:format :UPC-A :size [200 201 202]})))

    (is (thrown-with-msg? Exception #"value should be a number not ABC, a string"
                          (qr/encode text tmp {:format :UPC-A :size ["ABC" 256]})))
    ))

(deftest arg-test
  (let [text "https://babashka.org"]
    (is (thrown-with-msg? Exception #"encode expects at least 2 arguments: text and output path"
                          (qr/encode tmp))
        "one argument")
    (is (thrown-with-msg? Exception #"encode: text must be a string"
                          (qr/encode 55 tmp)))
    (is (thrown-with-msg? Exception #"encode: output path must be a string"
                          (qr/encode text 55)))
    ;(is (thrown-with-msg? Exception #"Could not resolve symbol: qr/unknown"
    ;                      (qr/unknown)))
    (is (thrown-with-msg? Exception #"expected a vector of arguments"
                          (qr/decode)))
    (is (thrown-with-msg? Exception #"decode expects a file path string or image bytes"
                          (qr/decode [])))
    (is (thrown-with-msg? Exception #"open : no such file or directory"
                          (qr/decode "")))
    (is (thrown-with-msg? Exception #"image: unknown format"
                          (qr/decode ".")))
    ))

(deftest ec-level-option-test
  (let [text "725272730706"]
    (qr/encode text tmp {:ec-level :L})
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)
    (qr/encode text tmp {:ec-level :M})
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)
    (qr/encode text tmp {:ec-level :Q})
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)
    (qr/encode text tmp {:ec-level :H})
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)
    (is (thrown-with-msg? Exception #"ec-level must be one of L/M/Q/H not :Z"
                          (qr/encode text tmp {:ec-level :Z}))
        "check error if invalid ec-level used")
    (is (thrown-with-msg? Exception #"ec-level only supported for QR barcode encoding"
                          (qr/encode text tmp {:format :DataMatrix :ec-level :H}))
        "ec-level option only supported for QR encoding")
    (is (thrown-with-msg? Exception #"unsupported decode option: :ec-level"
                          (qr/decode tmp {:ec-level :H}))
        "ec-level option not supported for QR decoding")
    ))


(deftest option-test
  (let [text "725272730706"]
    (is (thrown-with-msg? Exception #"unsupported encode option: :bad-option"
                          (qr/encode text tmp {:bad-option 5})))
    (is (thrown-with-msg? Exception #"Unsupported barcode format for encode: :PDF417"
                          (qr/encode text tmp {:format :PDF417}))
        "unsupported barcode PDF417 (sadly)")
    (is (thrown-with-msg? Exception #"unsupported decode option: :bad-option"
                          (qr/decode tmp {:bad-option 5})))
    ))

(defn files-equal [file1 file2]
  (let [d (process/shell {:continue true :out :string} "cmp" file1 file2)]
  (zero? (:exit d))))

(defn file-compare-tst 
  ([barcode size text] (file-compare-tst barcode size nil text))
  ([barcode size ec-level text] 
  (let [ssize (if (vector? size) (string/join "x" size) size)
        file1 (as-> [(name barcode) ssize] x
                  (if ec-level (conj x (name ec-level)) x)
                  (string/join "-" x)
                  (str x ".png"))
        file2 (str "test/references/" file1)
        opts {:format barcode :size size}]
    (qr/encode text file1 (if (some? ec-level) (assoc opts :ec-level ec-level) opts))
    (is (files-equal file1 file2) (str "Compare " file1 " and " file2))
    (fs/delete-if-exists file1))))

(deftest reference-test
    (file-compare-tst :QR  256 :H "https://babashka.org")
    (file-compare-tst :QR  512 :Q "https://babashka.org")
    (file-compare-tst :QR  768 :M "https://babashka.org")
    (file-compare-tst :QR 1024 :L "https://babashka.org")
    (file-compare-tst :DataMatrix 256 "https://babashka.org")
    (file-compare-tst :DataMatrix 512 "https://babashka.org")
    (file-compare-tst :UPC-A [112 32] "725272730706")
    (file-compare-tst :Code128 [192 32] "8901234567890123456")
    (file-compare-tst :Code39 [320 32] "8901234567890123456")
  )

(let [{:keys [fail error]} (t/run-tests)]
  (fs/delete-if-exists tmp)
  (System/exit (+ fail error)))
