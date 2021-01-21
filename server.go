package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	_ "github.com/lib/pq"
)

func main() {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}

	http.HandleFunc("/", requestHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	go srv.ListenAndServe()
	go gracefulShutdown(srv)
	forever := make(chan int)
	<-forever
}

const (
	host     = "localhost"
	port     = 5436
	user     = "postgres"
	password = "secret"
	dbname   = "EthereumData"
)

var (
	blockNumberGlobal int64
	dataBaseGlobal    *sql.DB
)

//Transaction is a final structure that is used to create a response in JSON format {"transactions":___,"amount":___}
type Transaction struct {
	Transactions int `json:"transactions"`
	//the number can be huge, so it should be passed as string to save the accuracy
	Amount *big.Float `json:"amount"`
}

func handleHTTPStatus(w http.ResponseWriter, errorCode int) {
	w.WriteHeader(errorCode)
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//Function converts Wei to Ether and return Ether amount in big.Float type
func weiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
}

func getAPIKey() string {
	APIKeyString, err := ioutil.ReadFile("config.ini")
	handleError(err)
	APIKey := string(APIKeyString[11 : len(APIKeyString)-1])
	return APIKey
}

func getBlockNumberInHex(w http.ResponseWriter, r *http.Request) string {
	blockNumberInt, err := strconv.ParseInt(r.URL.Path[strings.Index(r.URL.Path, "block")+6:strings.Index(r.URL.Path, "/total")], 10, 64)
	if err != nil {
		handleHTTPStatus(w, 400)
	}
	blockNumberGlobal = blockNumberInt
	return fmt.Sprintf("%x", blockNumberInt)
}

func getRequestLink(w http.ResponseWriter, r *http.Request) string {
	blockNumber := getBlockNumberInHex(w, r)
	APIKey := getAPIKey()
	stringForRequest, err := url.Parse("https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=&boolean=true&apikey=")
	handleError(err)
	rawLink := stringForRequest.Query()
	rawLink.Set("tag", blockNumber)
	rawLink.Set("apikey", APIKey)
	stringForRequest.RawQuery = rawLink.Encode()
	return stringForRequest.String()
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "images/favicon.ico")
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	requestLink := getRequestLink(w, r)
	receivedData, err := http.Get(requestLink)
	handleError(err)
	receivedDataByte, err := ioutil.ReadAll(receivedData.Body)
	handleError(err)

	writeDataToJSON(w, receivedDataByte)
}

func ethereumAnalyzer(text string) (int, *big.Float) {
	//As blockHash appears once per block - counting it will give us the number of transactions
	amountOfTransactions := strings.Count(text, `"blockHash"`)
	splitedString := strings.Split(text, ",")
	//Sometimes Wei amount is bigger than uint64, so we have to use math/big to save accuracy of numbers
	var bigIntFinal = new(big.Int)

	for _, word := range splitedString {
		if word[1:6] == "value" {
			hexString := word[9 : len(word)-1]
			var bigInt, _ = new(big.Int).SetString(hexString, 0)
			bigIntFinal.Add(bigIntFinal, bigInt)
		}

	}

	value := weiToEther(bigIntFinal)
	return amountOfTransactions, value
}

func writeDataToJSON(w http.ResponseWriter, receivedDataByte []byte) {
	amountOfTransactions, value := ethereumAnalyzer(string(receivedDataByte))
	if amountOfTransactions == 0 {
		handleHTTPStatus(w, 204)
	}

	transactionsForJSON := Transaction{Transactions: amountOfTransactions, Amount: value}
	JSONFile, err := json.Marshal(transactionsForJSON)
	handleError(err)
	ioutil.WriteFile("data.json", JSONFile, os.ModePerm)

	insertToDatabase(amountOfTransactions, value.Text('g', 30))

	fmt.Fprintf(w, string(JSONFile))
}

func insertToDatabase(transactions int, amount string) {
	psqlInfo := fmt.Sprintf("host=%s port=%d password=%s user=%s dbname=%s sslmode=disable", host, port, password, user, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	handleError(err)
	dataBaseGlobal = db
	defer dataBaseGlobal.Close()
	err = dataBaseGlobal.Ping()
	handleError(err)

	query := fmt.Sprintf("INSERT INTO blocks (blockId, transactions, amount) VALUES (%d, %d, %s)", blockNumberGlobal, transactions, amount)
	dataBaseGlobal.QueryRow(query)
}

func gracefulShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-quit
		fmt.Println("Graceful shutdown...")
		srv.Shutdown(context.Background())
		fmt.Println("Server is stopped.")
		dataBaseGlobal.Close()
		fmt.Println("Database is closed.")
		os.Exit(0)
	}()
}
