package models

import "errors"

type Match struct {
	MatchID MatchID `json:"MatchID"`
	DateID DateID `json:"DateID"`	
	GameID GameID `json:"GameID"`
	TeamNames []string `json:"TeamNames"`
	TeamScores []int `json:"TeamScores"`
	TeamMembers [][]string `json:"TeamMembers"`
	PlayerAttributesMap map[UserID]AttributesStatsMap `json:"PlayerAttributesMap"`
}

func NewMatch(matchID MatchID, dateID DateID, gameID GameID, teamNames []string, teamScores []int, teamMembers [][]string, playerAttributesMap map[UserID]AttributesStatsMap) (*Match, error) {
	if matchID == "" {
		return nil, errors.New("match id cannot be empty")
	}
	if dateID == "" {
		return nil, errors.New("date id cannot be empty")
	}
	if gameID == "" {
		return nil, errors.New("game id cannot be empty")
	}
	if len(teamNames) == 0 {
		return nil, errors.New("team names list cannot be empty")
	}
	if len(teamScores) == 0 {
		return nil, errors.New("team scores list cannot be empty")
	}
	if len(teamMembers) == 0 {
		return nil, errors.New("team members list cannot be empty")
	}
	if len(teamMembers) != len(teamScores) {
		return nil, errors.New("team members list and team scores list must have the same length")
	}
	if playerAttributesMap == nil {
		playerAttributesMap = make(map[UserID]AttributesStatsMap)
	}
	// Check if any attribute stat value is negative
	for _, attrs := range playerAttributesMap {
		for _, value := range attrs {
			if value < 0 {
				return nil, errors.New("attribute stat values cannot be negative")
			}
		}
	}
	return &Match{
		MatchID: matchID,
		DateID: dateID,
		GameID: gameID,
		TeamNames: teamNames,
		TeamScores: teamScores,
		TeamMembers: teamMembers,
		PlayerAttributesMap: playerAttributesMap,
	}, nil
}

func (m *Match) GetPlayerAttributes(userID UserID) (AttributesStatsMap, bool) {
	if m.PlayerAttributesMap == nil {
		return nil, false
	}
	attrs, ok := m.PlayerAttributesMap[userID]
	return attrs, ok
}

func (m *Match) SetPlayerAttributes(userID UserID, attrs AttributesStatsMap) {
	if m.PlayerAttributesMap == nil {
		m.PlayerAttributesMap = make(map[UserID]AttributesStatsMap)
	}
	m.PlayerAttributesMap[userID] = attrs
}