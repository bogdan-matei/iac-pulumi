package main

import(
	"math/rand"
	"encoding/hex"	
)

func RandomString(len int) string {
	b := make([]byte, len)
	rand.Read(b) 
	return hex.EncodeToString(b)
}
