package messaging

import (
	"github.com/Ansiblock/Ansiblock/books"
)

// ProcessMessages processes user requests, like check balance,
// get ValidVDFValue and get Transaction count for testing.
func ProcessMessages(messages []Request, bm *books.Accounts) *Responses {
	responses := make([]Response, 0, len(messages))
	for _, message := range messages {
		switch message.Type {
		case Balance:
			balance := bm.Balance(message.PublicKey)
			response := ResponseBalance{Value: balance, Addr: message.Addr, PublicKey: message.PublicKey}
			responses = append(responses, &response)
		case ValidVDFValue:
			validVDFValue := bm.ValidVDFValue()
			response := ResponseValidVDFValue{Value: validVDFValue, Addr: message.Addr}
			responses = append(responses, &response)
		case TransactionsTotal:
			transactionsTotal := bm.TransactionsTotal()
			response := ResponseTransactionsTotal{Addr: message.Addr, Value: transactionsTotal}
			responses = append(responses, &response)
		}
	}
	return &Responses{Responses: responses}
}
