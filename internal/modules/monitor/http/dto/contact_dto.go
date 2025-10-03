package dto

type CreateContactRequest struct {
	Name        string `json:"name"`
	ContactType string `json:"contact_type"`
	ContactData string `json:"contact_data"`
}

type CreateContactResponse struct {
	ContactID   uint64 `json:"contact_id"`
	Name        string `json:"name"`
	ContactType string `json:"contact_type"`
	ContactData string `json:"contact_data"`
}

type UpdateContactRequest struct {
	Name        string `json:"name"`
	ContactType string `json:"contact_type"`
	ContactData string `json:"contact_data"`
	IsEnabled   bool   `json:"is_enabled"`
}

type ContactResponse struct {
	ContactID   uint64 `json:"contact_id"`
	Name        string `json:"name"`
	ContactType string `json:"contact_type"`
	ContactData string `json:"contact_data"`
	IsEnabled   bool   `json:"is_enabled"`
}
