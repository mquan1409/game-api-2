package models

type LeaderBoard struct {
	GameID        GameID         `json:"GameID"`
	AttributeName AttributeName  `json:"AttributeName"`
	UserIDs         []UserID       `json:"UserIDs"`
}

type BoundedLeaderboard struct {
	LeaderBoard
	Limit int            `json:"Limit"`
}

func NewLeaderBoard(gameID GameID, attributeName AttributeName, userIDs []UserID) LeaderBoard {
	if userIDs == nil {
		userIDs = []UserID{}
	}
	return LeaderBoard{
		GameID:        gameID,
		AttributeName: attributeName,
		UserIDs:       userIDs,
	}
}

func NewBoundedLeaderBoard(gameID GameID, attributeName AttributeName, userIDs []UserID, limit int) BoundedLeaderboard {
	return BoundedLeaderboard{
		LeaderBoard: NewLeaderBoard(gameID, attributeName, userIDs),
		Limit:       limit,
	}
}

