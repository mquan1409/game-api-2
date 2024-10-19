package models

import "errors"

type Game struct {
	GameID      GameID  `json:"game_id"`
	Description string            `json:"description"`
	Attributes  []AttributeName `json:"attributes"`
	RankedAttributes []AttributeName `json:"ranked_attributes"`
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
