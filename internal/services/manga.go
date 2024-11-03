package services

type MangaService struct {}

func NewMangaService() MangaService {
	return MangaService{}
}

func (s MangaService) GetNewestManga() error {
	return nil
}

func (s MangaService) SearchManga() error {
	return nil
}

func (s MangaService) GetMangaByID() error {
	return nil
}

func (s MangaService) PurchaseManga() error {
	return nil
}

func (s MangaService) GetPopularManga() error {
	return nil
}

func (s MangaService) RateManga() error {
	return nil
}

func (s MangaService) UpdateMangaRating() error {
	return nil
}

func (s MangaService) RemoveMangaRating() error {
	return nil
}
