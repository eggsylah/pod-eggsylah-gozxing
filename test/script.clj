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
    (is (files-equal file1 file2) (str "Compare " file1 " and " file2))))

(deftest reference-test
    (file-compare-tst :QR 1024 "https://babashka.org")
  )

(let [{:keys [fail error]} (t/run-tests)]
  (fs/delete-if-exists tmp)
  (System/exit (+ fail error)))
