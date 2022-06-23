package structs

type Order struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accural    *int   `json:"accural,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}
