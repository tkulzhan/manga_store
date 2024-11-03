package models

type Manga struct {
	ID          string   `json:"id" bson:"_id,omitempty"`
	Title       string   `json:"title" bson:"title"`
	Author      string   `json:"author" bson:"author"`
	Genres      []string `json:"genres" bson:"genres"`
	Price       float64  `json:"price" bson:"price"`
	Description string   `json:"description" bson:"description"`
	RatedTimes  int      `json:"ratedTimes" bson:"ratedTimes"`
	Rating      float64  `json:"rating" bson:"rating"`
	Views       int      `json:"views" bson:"views"`
	IsDeleted   bool     `json:"isDeleted" bson:"isDeleted"`
	CreatedAt   int      `json:"createdAt" bson:"createdAt"`
	Quantity    int      `json:"quantity" bson:"quantity"`
	Sold        int      `json:"sold" bson:"sold"`
}

type SearchMangaRequest struct {
	Query  string   `json:"query"`
	Genres []string `json:"genres"`
	Author string   `json:"author"`
	Limit  int      `json:"limit"`
}
