package dto

type CreationRequest struct {
	Name   *string  `json:"name"`
	Author *string  `json:"author"`
	Tags   []string `json:"tags"`
}

func (r CreationRequest) IsZero() bool {
	if r.Tags == nil {
		return true
	}
	return r.Name == nil || r.Author == nil
}
