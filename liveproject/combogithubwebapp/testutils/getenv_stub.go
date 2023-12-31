package testutils

func GetenvStub(envVar string) string {
	switch envVar {
	case "CLIENT_ID":
		return "a-client-id"

	case "CLIENT_SECRET":
		return "a-client-secret"
	default:
		return ""
	}
}
