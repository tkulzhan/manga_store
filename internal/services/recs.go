package services

type RecsService struct {}

func NewRecsService() RecsService {
	return RecsService{}
}

func (s RecsService) GetRecsByPreferences() error {
	return nil
}

func (s RecsService) GetRecsBySimilarUsers() error {
	return nil
}
