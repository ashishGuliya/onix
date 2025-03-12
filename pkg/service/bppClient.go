package service

import (
	"bytes"
	"context"
	"encoding/json"
	"onix/shared/log"
	"onix/shared/model"
	"onix/shared/protocol"
	"time"

	"google.golang.org/api/idtoken"
)

type BPPConfig struct {
	HostUrl string `yaml:"hostUrl"`
	BppID   string `yaml:"bppId"`
	BppUrl  string `yaml:"bppUrl"`
	BppName string `yaml:"bppName"`
}
type BPP struct {
	cfg *BPPConfig
}

func NewBPP(cfg *BPPConfig) *BPP {
	return &BPP{cfg}
}
func (bpp *BPP) search(q string) ([]protocol.Provider, error) {
	p := bpp.cfg.HostUrl + "/search"
	c, err := idtoken.NewClient(context.Background(), p)
	if err != nil {
		return nil, err
	}
	body, _ := json.Marshal(&model.BPPSearchRequest{Query: q})

	log.Debugf(context.Background(), "Calling Seller app : %s", string(body))
	resp, err := c.Post(p, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	var respBody model.BPPSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, err
	}
	return respBody.Products, nil
}

func (bpp *BPP) Search(req *protocol.SearchRequest) (*protocol.OnSearchRequest, error) {
	resp, err := bpp.search(req.Message.Intent.Item.Descriptor.Name)
	if err != nil {
		return nil, err
	}
	return &protocol.OnSearchRequest{
		Context: protocol.Context{
			Domain:        req.Context.Domain,
			Country:       req.Context.Country,
			City:          req.Context.City,
			CoreVersion:   "0.9.2",
			Action:        "on_search",
			BapID:         req.Context.BapID,
			BapURI:        req.Context.BapURI,
			BppURI:        bpp.cfg.BppUrl,
			BppID:         bpp.cfg.BppID,
			MessageID:     req.Context.MessageID,
			TransactionID: req.Context.TransactionID,
			Timestamp:     time.Now(),
		},
		Message: protocol.MessageForOnSearch{
			Catalog: protocol.Catalog{
				Descriptor:   protocol.Descriptor{Name: bpp.cfg.BppName},
				Fulfillments: []protocol.Fulfillment{{Type: "home-delivery"}},
				Payments:     []protocol.Payment{},
				Offers:       []protocol.Offer{},
				Providers:    resp,
				Exp:          time.Time{},
				TTL:          "",
			},
		},
	}, nil
}
