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


(deftest size-option-test
  (let [text "https://babashka.org"]
    (qr/encode text tmp {:size 512})
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)
    (qr/encode text tmp {:size 512.0})   ;float
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)
    (qr/encode text tmp {:size 512N})   ;big int
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)
    (qr/encode text tmp {:size (/ 1024 3)})   ;rational
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)
    (qr/encode text tmp {:size [512 256]})   ;vector
    (is (fs/exists? tmp))
    (fs/delete-if-exists tmp)

    (is (thrown-with-msg? Exception #"value should be a number not bad, a string"
                          (qr/encode text tmp {:size "bad"})))

    (is (thrown-with-msg? Exception #"value should be a number not :T, a transit.Keyword"
                          (qr/encode text tmp {:size :T})))

    (is (thrown-with-msg? Exception #"size must be a number or a 2-vector of numbers not \[1]"
                          (qr/encode text tmp {:size [1]})))

    (is (thrown-with-msg? Exception #"size must be a number or a 2-vector of numbers not \[200 201 202]"
                          (qr/encode text tmp {:size [200 201 202]})))

    (is (thrown-with-msg? Exception #"value should be a number not ABC, a string"
                          (qr/encode text tmp {:size ["ABC" 256]})))
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

(deftest option-test
  (let [text "725272730706"]
    (is (thrown-with-msg? Exception #"unsupported encode option: :bad-option"
                          (qr/encode text tmp {:bad-option 5})))
    (is (thrown-with-msg? Exception #"unsupported decode option: :bad-option"
                          (qr/decode tmp {:bad-option 5})))
    ))

(defn files-equal [file1 file2]
  (let [d (process/shell {:continue true :out :string} "cmp" file1 file2)]
  (zero? (:exit d))))

(defn file-compare-tst 
  [barcode size text] 
  (let [ssize (if (vector? size) (string/join "x" size) size)
        file1 (as-> [(name barcode) ssize] x
                  (conj x "L")
                  (string/join "-" x)
                  (str x ".png"))
        file2 (str "test/references/" file1)
        opts {:size size}]
    (qr/encode text file1 opts)
    (is (files-equal file1 file2) (str "Compare " file1 " and " file2))
    (fs/delete-if-exists file1)))

(deftest reference-test
    (file-compare-tst :QR 1024 "https://babashka.org")
  )

(let [{:keys [fail error]} (t/run-tests)]
  (fs/delete-if-exists tmp)
  (System/exit (+ fail error)))
