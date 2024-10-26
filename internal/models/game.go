package models

import "errors"

type Game struct {
	GameID      GameID  `json:"GameID"`
	Description string            `json:"Description"`
	Attributes  []AttributeName `json:"Attributes"`
	RankedAttributes []AttributeName `json:"RankedAttributes"`
}

func NewGame(id GameID, description string, attributes []AttributeName, rankedAttributes []AttributeName) (*Game, error) {
	if id == "" {
		return nil, errors.New("game id cannot be empty")
	}

	if attributes == nil {
		attributes = []AttributeName{}
	}

	if rankedAttributes == nil {
		rankedAttributes = []AttributeName{}
	}

	return &Game{
		GameID:            id,
		Description:       description,
		Attributes:        attributes,
		RankedAttributes:  rankedAttributes,
	}, nil
}
