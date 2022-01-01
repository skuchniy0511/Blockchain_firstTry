package main

import (
	bc "blockhain/blockchain"
	nt "blockhain/network"
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func init() {
	if len(os.Args) < 2 {
		panic("failed 1")
	}
	var (
		addrStr     = ""
		UserNewStr  = ""
		UserLoadStr = ""
	)

	var (
		addrExist     = false
		userNewExist  = false
		userLoadExist = false
	)
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case strings.HasPrefix(arg, "-loadaddr:"):
			addrStr = strings.Replace(arg, "-loadaddr:", "", 1)
			addrExist = true

		case strings.HasPrefix(arg, "-newuser:"):
			UserNewStr = strings.Replace(arg, "-newuser:", "", 1)
			userNewExist = true

		case strings.HasPrefix(arg, "-loaduser:"):
			UserLoadStr = strings.Replace(arg, "-loaduser:", "", 1)
			userLoadExist = true
		}
	}
	if !(userNewExist || userLoadExist) || !addrExist {
		panic("failed:!(userNewExist || userLoadExist) || !addrExist")
	}
	err := json.Unmarshal([]byte(readFile(addrStr)), &Addresses)
	if err != nil {
		panic("failed 3")
	}
	if len(Addresses) == 0 {
		panic("failed 4")
	}
	if userNewExist {
		User = userNew(UserNewStr)
	}
	if userLoadExist {
		User = userLoad(UserLoadStr)
	}
	if User == nil {
		panic("failed 5 ")
	}
}

func readFile(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}
	return string(data)
}

func userNew(filename string) *bc.User {
	user := bc.NewUser()
	if user == nil {
		return nil
	}

	err := writeFile(filename, user.Purse())
	if err != nil {
		return nil
	}
	return user
}

func userLoad(filename string) *bc.User {
	priv := readFile(filename)
	if priv == "" {
		return nil
	}
	user := bc.LoadUser(priv)
	if user == nil {
		return nil
	}
	return user
}

var (
	Addresses []string
	User      *bc.User
)

func writeFile(filename, data string) error {
	return ioutil.WriteFile(filename, []byte(data), 0644)
}

func handleClient() {
	var (
		message string
		splited []string
	)

	for {
		message = inputString("> ")
		splited = strings.Split(message, " ")
		switch splited[0] {
		case "/exit":
			os.Exit(0)
		case "/user":
			if len(splited) < 2 {
				fmt.Println("failed: len(user) <2/n")
				continue
			}
			switch splited[1] {
			case "address":
				userAddress()
			case "purse":
				userPurse()
			case "balance":
				userBalance()
			default:
				fmt.Println("command undefined\n")
			}
		case "/chain":
			if len(splited) < 2 {
				fmt.Println("failed: len(chain) < 2\n")
				continue
			}
			switch splited[1] {
			case "print":
				chainPrint()
			case "tx":
				cahinTx(splited[1:])
			case "balance":
				chainBalance(splited[1:])
			case "block":
				chainBlock(splited[1:])
			case "size":
				chainSize()
			default:
				fmt.Println("command undefined\n")
			}
		default:
			fmt.Println("command undefined\n")
		}
	}
}

func inputString(begin string) string {
	fmt.Print(begin)
	msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.Replace(msg, "\n", "", 1)
}

func userAddress() {
	fmt.Println("Address:", User.Address(), "\n")
}

func userPurse() {
	fmt.Println("Purse:", User.Purse(), "\n")
}

func userBalance() {
	printBalance(User.Address())
}

func chainPrint() {
	for i := 0; ; i++ {
		res := nt.Send(Addresses[0], &nt.Package{
			Option: GET_BLOCK,
			Data:   fmt.Sprintf("%d", i),
		})
		if res == nil || res.Data == "" {
			break
		}
		fmt.Printf("[%d]=>%s/n", i+1, res.Data)
	}
	fmt.Println()
}

func cahinTx(splited []string) {
	if len(splited) != 3 {
		fmt.Println("len(splited) != 3\n")
		return
	}
	num, err := strconv.Atoi(splited[2])
	if err != nil {
		fmt.Println("strconv error\n")
		return
	}
	for _, addr := range Addresses {
		res := nt.Send(addr, &nt.Package{
			Option: GET_LASTHASH,
		})
		if res == nil {
			continue
		}
		tx := bc.NewTransaction(User, bc.Base64Decode(res.Data), splited[1], uint64(num))
		if tx == nil {
			fmt.Println("tx is null\n")
			return
		}
		res = nt.Send(addr, &nt.Package{
			Option: ADD_TRASX,
			Data:   bc.SerializeTX(tx),
		})
		if res == nil {
			continue
		}
		if res.Data == "ok" {
			fmt.Printf("ok: (%s)\n", addr)
		} else {
			fmt.Printf("fail: (%s)\n", addr)
		}
	}
	fmt.Println()
}

func chainBalance(splited []string) {
	if len(splited) != 2 {
		fmt.Println("len(splited) != 2\n")
		return
	}
	printBalance(splited[1])
}

func printBalance(address string) {
	for _, addr := range Addresses {
		res := nt.Send(addr, &nt.Package{
			Option: GET_BLNCE,
			Data:   address,
		})
		if res == nil {
			continue
		}
		fmt.Println("Balance (%s): %s coins\n", addr, res.Data)
	}
	fmt.Println()
}

func chainBlock(splited []string) {
	if len(splited) != 2 {
		fmt.Println("failed: len(splited) != 2\n")
		return
	}
}

func chainSize() {
	res := nt.Send(Addresses[1], &nt.Package{
		Option: GET_CSIZE,
	})
	if res == nil || res.Data == "" {
		fmt.Println("failed: getSize\n")
		return
	}
	fmt.Printf("Size: %s blocks\n\n", res.Data)
}

const (
	ADD_BLOCK = iota + 1
	ADD_TRASX
	GET_BLOCK
	GET_LASTHASH
	GET_BLNCE
	GET_CSIZE
)

func main() {
	handleClient()
}
