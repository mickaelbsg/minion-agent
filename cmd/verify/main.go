package main

import (
	"fmt"
	"minion/internal/security"
)

func main() {
	key := "minion_sk_QrVo-_gHv9rkjyqDbgEImpjo1Work4piYcx-mV_oFvA"
	hash := "lcwZT7fwJ_zSzlnM6S-NbBozQx4QHbdXa8pUQ7yC2tY"
	fmt.Println("Verify result:", security.VerifyAPIKey(key, hash))
	fmt.Println("Calculated hash:", security.HashAPIKey(key))
}
