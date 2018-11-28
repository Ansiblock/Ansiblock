package messaging

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/Ansiblock/Ansiblock/block"
)

func TestResponseBalanceSerialize(t *testing.T) {
	responseAddr := net.UDPAddr{Port: 12345, IP: net.ParseIP("127.0.0.1")}
	keyPair := block.NewKeyPair()
	response := ResponseBalance{Value: 1024, Addr: &responseAddr, PublicKey: keyPair.Public}

	serializedResponse := response.Serialize()
	fmt.Println(serializedResponse)

	failed := (serializedResponse.Data[0] != Balance) &&
		!reflect.DeepEqual(keyPair, serializedResponse.Data[9:41])
	if failed {
		t.Errorf(`serialization of ResponseBalance failed`)
	}
}

func TestResponseBalanceDeserialize(t *testing.T) {
	responseAddr := net.UDPAddr{Port: 12345, IP: net.ParseIP("127.0.0.1")}
	keyPair := block.NewKeyPair()
	response := ResponseBalance{Value: 1024, Addr: &responseAddr, PublicKey: keyPair.Public}
	fmt.Println(response)

	serializedResponse := response.Serialize()
	fmt.Println(serializedResponse)

	var deserializedResponse ResponseBalance
	deserializedResponse.Deserialize(serializedResponse)
	fmt.Println(deserializedResponse)

	if !reflect.DeepEqual(response, deserializedResponse) {
		t.Errorf(`deserialized response does not equal original. \n
		Original: %v
		Deserialized %v`, response, deserializedResponse)
	}
}

func TestResponsResponseValidVDFValueSerialize(t *testing.T) {
	responseAddr := net.UDPAddr{Port: 12345, IP: net.ParseIP("127.0.0.1")}
	testValidVDFValue := block.VDF([]byte("TestResponsResponse"))
	response := ResponseValidVDFValue{Value: testValidVDFValue, Addr: &responseAddr}
	fmt.Println(response)

	serializedResponse := response.Serialize()
	fmt.Println(serializedResponse)

	failed := (serializedResponse.Data[0] != ValidVDFValue) &&
		!reflect.DeepEqual(testValidVDFValue, serializedResponse.Data[1:1+block.VDFSize])
	if failed {
		t.Errorf(`serialization of ResponseBalance failed`)
	}

}

func TestResponsResponseValidVDFValueDeserialize(t *testing.T) {
	responseAddr := net.UDPAddr{Port: 12345, IP: net.ParseIP("127.0.0.1")}
	testValidVDFValue := block.VDF([]byte("TestResponsResponse"))
	response := ResponseValidVDFValue{Value: testValidVDFValue, Addr: &responseAddr}
	fmt.Println(response)

	serializedResponse := response.Serialize()
	fmt.Println(serializedResponse)

	var deserializedResponse ResponseValidVDFValue
	deserializedResponse.Deserialize(serializedResponse)
	fmt.Println(deserializedResponse)

	if !reflect.DeepEqual(response, deserializedResponse) {
		t.Errorf(`deserialized response does not equal original. \n
		Original: %v
		Deserialized %v`, response, deserializedResponse)
	}
}

func TestResponseTransactionsTotalSerialize(t *testing.T) {
	responseAddr := net.UDPAddr{Port: 12345, IP: net.ParseIP("127.0.0.1")}
	testTransactionsTotal := uint64(1024)
	response := ResponseTransactionsTotal{Value: testTransactionsTotal, Addr: &responseAddr}
	fmt.Println(response)

	serializedResponse := response.Serialize()
	fmt.Println(serializedResponse)

	failed := (serializedResponse.Data[0] != TransactionsTotal) &&
		!reflect.DeepEqual(testTransactionsTotal, serializedResponse.Data[1:10])
	if failed {
		t.Errorf(`serialization of ResponseBalance failed`)
	}
}

func TestResponseTransactionsTotalDeserialize(t *testing.T) {
	responseAddr := net.UDPAddr{Port: 12345, IP: net.ParseIP("127.0.0.1")}
	testTransactionsTotal := uint64(1024)
	response := ResponseTransactionsTotal{Value: testTransactionsTotal, Addr: &responseAddr}
	fmt.Println(response)

	serializedResponse := response.Serialize()
	fmt.Println(serializedResponse)

	var deserializedResponse ResponseTransactionsTotal
	deserializedResponse.Deserialize(serializedResponse)
	fmt.Println(deserializedResponse)

	if !reflect.DeepEqual(response, deserializedResponse) {
		t.Errorf(`deserialized response does not equal original. \n
		Original: %v
		Deserialized %v`, response, deserializedResponse)
	}
}

func TestResponsesSerialize(t *testing.T) {
	responseAddr := net.UDPAddr{Port: 12345, IP: net.ParseIP("127.0.0.1")}
	keyPair := block.NewKeyPair()
	responseBalance := ResponseBalance{Value: 1024, Addr: &responseAddr, PublicKey: keyPair.Public}
	fmt.Println(responseBalance)

	testValidVDFValue := block.VDF([]byte("TestResponsResponse"))
	responseValidVDF := ResponseValidVDFValue{Value: testValidVDFValue, Addr: &responseAddr}
	fmt.Println(responseValidVDF)

	testTransactionsTotal := uint64(1024)
	responseTransactionsTotal := ResponseTransactionsTotal{Value: testTransactionsTotal, Addr: &responseAddr}
	fmt.Println(responseTransactionsTotal)

	var allResponces Responses
	allResponces.Responses = append(allResponces.Responses, &responseBalance)
	allResponces.Responses = append(allResponces.Responses, &responseValidVDF)
	allResponces.Responses = append(allResponces.Responses, &responseTransactionsTotal)

	serializedAllResponses := allResponces.Serialize()
	fmt.Println(serializedAllResponses)
}

func TestResponsesDeserialize(t *testing.T) {
	responseAddr := net.UDPAddr{Port: 12345, IP: net.ParseIP("127.0.0.1")}
	keyPair := block.NewKeyPair()
	responseBalance := ResponseBalance{Value: 1024, Addr: &responseAddr, PublicKey: keyPair.Public}
	fmt.Println(responseBalance)

	testValidVDFValue := block.VDF([]byte("TestResponsResponse"))
	responseValidVDF := ResponseValidVDFValue{Value: testValidVDFValue, Addr: &responseAddr}
	fmt.Println(responseValidVDF)

	testTransactionsTotal := uint64(1024)
	responseTransactionsTotal := ResponseTransactionsTotal{Value: testTransactionsTotal, Addr: &responseAddr}
	fmt.Println(responseTransactionsTotal)

	var allResponces Responses
	allResponces.Responses = append(allResponces.Responses, &responseBalance)
	allResponces.Responses = append(allResponces.Responses, &responseValidVDF)
	allResponces.Responses = append(allResponces.Responses, &responseTransactionsTotal)

	serializedAllResponses := allResponces.Serialize()

	var deserializedResponses Responses
	deserializedResponses.Deserialize(serializedAllResponses)

	test0 := deserializedResponses.Responses[0].(*ResponseBalance)
	test1 := deserializedResponses.Responses[1].(*ResponseValidVDFValue)
	test2 := deserializedResponses.Responses[2].(*ResponseTransactionsTotal)

	fmt.Println(*test0, *test1, *test2)

	failed := !reflect.DeepEqual(responseBalance, *test0) &&
		!reflect.DeepEqual(responseValidVDF, *test1) &&
		!reflect.DeepEqual(responseTransactionsTotal, *test2)

	if failed {
		t.Errorf(`serialization and deserialized of multiple responses has failed`)
	}

}
