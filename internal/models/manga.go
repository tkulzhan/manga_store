package models

type Manga struct {
	ID            string   `json:"id" bson:"_id,omitempty"`
	Title         string   `json:"title" bson:"title"`
	Author        string   `json:"author" bson:"author"`
	Genres        []string `json:"genres" bson:"genres"`
	Price         float64  `json:"price" bson:"price"`
	Description   string   `json:"description" bson:"description"`
	Ratings       []Rating `json:"ratings" bson:"ratings"`
	AverageRating float64  `json:"averageRating" bson:"averageRating"`
	Likes         int      `json:"likes" bson:"likes"`
	Views         int      `json:"views" bson:"views"`
	IsDeleted     bool     `json:"isDeleted" bson:"isDeleted"`
}

type Rating struct {
	UserID string `json:"userId" bson:"userId"`
	Score  int    `json:"score" bson:"score"`
}
