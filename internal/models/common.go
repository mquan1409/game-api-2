package models

type GameID string
type UserID string
type DateID string
type MatchID string
type AttributeName string
type AttributeStat int

// Attributes represents the attributes of a player in a game.
type AttributesStatsMap map[AttributeName]AttributeStat
