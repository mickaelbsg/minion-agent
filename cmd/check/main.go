package main
import (
    "fmt"
    "minion/internal/security"
)
func main(){
    key := "minion_sk_Qse7d49RT3XdbYD_W82lHjs9P7czZIktif0ZX_Z1KSo"
    hash := "wmg96hGHLFhJ9yxRuH-fmgmQJ-ENnHcbhGMU7ao1Zwk"
    fmt.Println("verify:", security.VerifyAPIKey(key, hash))
    fmt.Println("hash:", security.HashAPIKey(key))
}
