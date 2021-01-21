package main

import (
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

/* Each test has a structure with inputs - data that we pass to our function, and
outputs - data that we expect to get. In each test we 1. Create all needed variables.
2. Iterate through them and compare expectedOutput with real output.
3. If outputs are different - test is considered as failed.*/

//This value will be used for comparing the difference between 2 values big.Float type
var comparingValidPoint, _ = new(big.Float).SetPrec(128).SetString("0.000000000000000001")

//math/big doesn't have overaloaded operators "+, -...", we need a temporary big.Float to store the result of such operations.
var tempBigFloat = new(big.Float)

var w = httptest.NewRecorder()

//Testing weiToEther function.

type TestWeiToEtherData struct {
	input          *big.Int
	expectedOutput *big.Float
}

func TestWeiToEther(t *testing.T) {
	var (
		bigInt1, _   = new(big.Int).SetString("345678912345678923123456", 10)
		bigFloat1, _ = new(big.Float).SetPrec(128).SetString("345678.912345678923123456")
		bigInt2, _   = new(big.Int).SetString("3294795325627368526387", 10)
		bigFloat2, _ = new(big.Float).SetPrec(128).SetString("3294.795325627368526387")
		bigInt3, _   = new(big.Int).SetString("132472893483477889", 10)
		bigFloat3, _ = new(big.Float).SetPrec(128).SetString("0.132472893483477889")
	)
	dataItems := []TestWeiToEtherData{
		{bigInt1, bigFloat1},
		{bigInt2, bigFloat2},
		{bigInt3, bigFloat3},
	}

	for _, item := range dataItems {
		result := weiToEther(item.input)
		/*Our function returns big.Float that is a result of converting Wei to Ether.
		As these numbers have a big tail, we cannot just compare it, so we subtract our result from expected output
		and if the difference is less than 1 Wei, we accept it -> test is passed.*/
		if tempBigFloat.Sub(item.expectedOutput, result).Cmp(comparingValidPoint) != -1 {
			t.Errorf("weiToEther failed, expected %g, got %g", item.expectedOutput, result)
		}

	}

}

//Testing getBlockNumberInHex function.

type TestGetBlockNumberInHexData struct {
	input          *http.Request
	expectedOutput string
}

func TestGetBlockNumberInHex(t *testing.T) {
	testHTTPEndpoint1, _ := http.NewRequest("POST", "/api/block/11508993/total/yguyfdgdfgdfg", nil)
	testHTTPEndpoint2, _ := http.NewRequest("POST", "/api/got/something/block/14564562/total", nil)
	testHTTPEndpoint3, _ := http.NewRequest("POST", "/api/block/980/total/yguyfdgdfgdfg", nil)

	dataItems := []TestGetBlockNumberInHexData{
		{testHTTPEndpoint1, "af9d01"},
		{testHTTPEndpoint2, "de3cd2"},
		{testHTTPEndpoint3, "3d4"},
	}

	for _, item := range dataItems {
		result := getBlockNumberInHex(w, item.input)
		if result != item.expectedOutput {
			t.Errorf("getBlockNumberInHex failed, expected %s, got %s", item.expectedOutput, result)
		}
	}
}

//Testing getRequestLink function.

type TestGetRequestLinkData struct {
	input          *http.Request
	expectedOutput string
}

func TestGetRequestLink(t *testing.T) {
	testHTTPEndpoint1, _ := http.NewRequest("POST", "/api/block/11508993/total/yguyfdgdfgdfg", nil)
	testHTTPEndpoint2, _ := http.NewRequest("POST", "/api/got/something/block/14564562/total", nil)
	testHTTPEndpoint3, _ := http.NewRequest("POST", "/block/980/total/yguyfdgdfgdfg", nil)
	correctCompletedLink1 := "https://api.etherscan.io/api?action=eth_getBlockByNumber&apikey=5HHQQC7NM61MC5AQM9DK5PT2TMB4GW59I8&boolean=true&module=proxy&tag=af9d01"
	correctCompletedLink2 := "https://api.etherscan.io/api?action=eth_getBlockByNumber&apikey=5HHQQC7NM61MC5AQM9DK5PT2TMB4GW59I8&boolean=true&module=proxy&tag=de3cd2"
	correctCompletedLink3 := "https://api.etherscan.io/api?action=eth_getBlockByNumber&apikey=5HHQQC7NM61MC5AQM9DK5PT2TMB4GW59I8&boolean=true&module=proxy&tag=3d4"

	dataItems := []TestGetBlockNumberInHexData{
		{testHTTPEndpoint1, correctCompletedLink1},
		{testHTTPEndpoint2, correctCompletedLink2},
		{testHTTPEndpoint3, correctCompletedLink3},
	}

	for _, item := range dataItems {
		result := getRequestLink(w, item.input)
		if result != item.expectedOutput {
			t.Errorf("getRequestLink failed, expected %s, got %s", item.expectedOutput, result)
		}
	}
}

//Testing ethereumAnalyzer function.

type TestEthereumAnalyzerData struct {
	input                      string
	expectedOutputTransactions int
	expectedOutputAmount       *big.Float
}

func TestEthereumAnalyzer(t *testing.T) {

	pageForTest1, _ := ioutil.ReadFile("TestPages/page1.txt")
	pageForTest2, _ := ioutil.ReadFile("TestPages/page2.txt")
	pageForTest3, _ := ioutil.ReadFile("TestPages/page3.txt")

	bigFloatExpectedAmount1, _ := new(big.Float).SetPrec(128).SetString("1130.987085446826418822")
	bigFloatExpectedAmount2, _ := new(big.Float).SetPrec(128).SetString("23.868260944825555116")
	bigFloatExpectedAmount3, _ := new(big.Float).SetPrec(128).SetString("2.170754146627130592")

	dataItems := []TestEthereumAnalyzerData{
		{string(pageForTest1), 241, bigFloatExpectedAmount1},
		{string(pageForTest2), 102, bigFloatExpectedAmount2},
		{string(pageForTest3), 27, bigFloatExpectedAmount3},
	}

	for _, item := range dataItems {
		resultTransactions, resultAmount := ethereumAnalyzer(item.input)

		if resultTransactions != item.expectedOutputTransactions {
			t.Errorf("ethereumAnalyzer failed, expected %d, got %d", item.expectedOutputTransactions, resultTransactions)
		} else if tempBigFloat.Sub(item.expectedOutputAmount, resultAmount).Cmp(comparingValidPoint) != -1 {
			t.Errorf("ethereumAnalyzer failed, expected %g, got %g", item.expectedOutputAmount, resultAmount)
		}

	}

}
