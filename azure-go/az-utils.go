package main

import(
	"math/rand"
	"encoding/hex"

	"os"
	"fmt"
)

func RandomString(len int) string {
	b := make([]byte, len)
	rand.Read(b) 
	return hex.EncodeToString(b)
}

func errorHandle(err error, crit bool) {
  if err != nil {
    fmt.Println(err)

    if crit {
      os.Exit(1)
    }
  }
}