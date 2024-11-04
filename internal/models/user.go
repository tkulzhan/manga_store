package models

type User struct {
	ID              string     `json:"id" bson:"_id,omitempty"`
	Name            string     `json:"name" bson:"name"`
	Email           string     `json:"email" bson:"email"`
	PasswordHash    string     `json:"passwordHash" bson:"passwordHash"`
	PurchaseHistory []Purchase `json:"purchaseHistory" bson:"purchaseHistory"`
	IsDeleted       bool       `json:"isDeleted" bson:"isDeleted"`
}

type Purchase struct {
	MangaID      string  `json:"mangaId" bson:"mangaId"`
	Title        string  `json:"title" bson:"title"`
	Price        float64 `json:"price" bson:"price"`
	PurchaseDate string  `json:"purchaseDate" bson:"purchaseDate"`
}

type PurchaseRequest struct {
	MangaID string `json:"mangaId"`
}
