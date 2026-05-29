#!/usr/bin/env bb

(ns script
  (:require [babashka.fs :as fs]
            [babashka.pods :as pods]
            [clojure.test :as t :refer [deftest is]]))

(pods/load-pod "./pod-babashka-gozxing")

(require '[pod.babashka.gozxing :as qr])

(def tmp "test/out.png")

(deftest round-trip-test
  (let [text "https://babashka.org"]
    (qr/encode text tmp {:size 256})
    (is (fs/exists? tmp))
    (is (= text (qr/decode tmp)))))

(deftest decode-bytes-test
  (let [text "from-bytes"]
    (qr/encode text tmp)
    (is (= text (qr/decode (fs/read-all-bytes tmp))))))

(let [{:keys [fail error]} (t/run-tests)]
  (fs/delete-if-exists tmp)
  (System/exit (+ fail error)))
