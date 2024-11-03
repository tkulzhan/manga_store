package models

type Rating struct {
	UserID  string  `json:"userId" bson:"userId"`
	MangaID string  `json:"mangaId" bson:"mangaId"`
	Score   float64 `json:"score" bson:"score"`
}
