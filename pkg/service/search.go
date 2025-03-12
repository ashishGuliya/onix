package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"onix/shared/log"
	"onix/shared/model"
	"onix/shared/protocol"

	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
)

type Search struct {
	target   string
	reciever string
}

func NewSearch(target, reciever string) *Search {

	return &Search{
		target:   target,
		reciever: reciever,
	}
}

func (s *Search) Search(msg *model.Msg) error {
	req, err := s.searchRequest(msg)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(req)
	host := s.target
	if len(req.Context.BppURI) != 0 {
		host = req.Context.BppURI
	}
	host += "/search"
	c, err := idtoken.NewClient(context.Background(), host)
	if err != nil {
		return err
	}
	log.Debugf(context.Background(), "Calling BPP at: %s", host)
	resp, err := c.Post(host, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	var respBody protocol.Response
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return err
	}
	if respBody.Error != nil || respBody.Message.Ack.Status == "NACK" {
		return fmt.Errorf("search call failed: %v", respBody)
	}
	return nil
}

func (s *Search) searchRequest(msg *model.Msg) (*protocol.SearchRequest, error) {

	txnID := msg.Request.TxnID
	if len(txnID) == 0 {
		txnID = uuid.NewString()
	}
	return &protocol.SearchRequest{
		Context: protocol.Context{
			Domain:        msg.Request.Criteria.Domain,
			Country:       "IND",
			City:          "BLR",
			CoreVersion:   "0.9.2",
			Action:        "search",
			MessageID:     msg.ID,
			TransactionID: txnID,
			Timestamp:     time.Now(),
			BppURI:        msg.Request.BapUrl,
			BapURI:        s.reciever,
		},
		Message: protocol.MessageForSearch{
			Intent: protocol.Intent{
				Item: protocol.Item{
					Descriptor: protocol.Descriptor{
						Name: msg.Request.Criteria.Query,
					},
				},
			},
		},
	}, nil
}
