package types

type CreateSecretPayload struct {
	PlainText string `json:"plain_text"`
}

type CreateSecretResponse struct {
	Id string `json:"id"`
}

type GetSecretResponse struct {
	Data string `json:"data"`
}

type SecretData struct {
	Id     string
	Secret string
}
