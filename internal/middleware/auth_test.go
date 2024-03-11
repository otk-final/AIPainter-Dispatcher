package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
	"testing"
)

func TestJWTCheck(t *testing.T) {

	privateKeyData, err := os.ReadFile("/Users/hxy/develops/Rust/AIPainter-Dispatcher/conf/jwt_public_key.pem")
	pem, err := jwt.ParseRSAPublicKeyFromPEM(privateKeyData)
	if err != nil {
		t.Error(err)
	}

	token := `eyJ4LXRlbmFudC1pZCI6ImFpcGFpbnRlciIsIngtYXBwLWlkIjoicGMiLCJraWQiOiI3Yjg1ZTM3OS1lODdhLTQ1NzItYTlmOS02NTZmMGUyNjlmODkiLCJ4LXVzZXItaWQiOiIxODc4NTAyMjcwMDQwMTUwMDIyIiwiYWxnIjoiUlMyNTYiLCJ4LXVzZXItdHlwZSI6InVzZXIiLCJ4LXVzZXItbmFtZSI6IueUqOaIt1p4OTNzIn0.eyJzdWIiOiIxRDIyM0Y5RTRFNUIiLCJhdWQiOiIxRDIyM0Y5RTRFNUIiLCJuYmYiOjE3MTAxNDUwNDIsInNjb3BlIjpbInBiIiwicHQiXSwicHJvZmlsZSI6eyJ2aXBFeHBpcmVkIjoiMjAyNC0xMC0xMCIsInZpcFR5cGUiOiJndWVzdCJ9LCJpc3MiOiJodHRwOi8vbG9jYWxob3N0Ojg5ODkiLCJleHAiOjE3MTAxNTIyNDIsImlhdCI6MTcxMDE0NTA0Mn0.GHviyM2QKhYVlAQagUIQPlGoqTIdWbv3fmPMiY6K9MYrBg38TUNRBwfTD51r1x9RM9LfOd-ULD71FYUsqOmkDnCFFF4huxEkfCsIIUk-SIMDGCzuLYWMzuiSKKcii_N1SuvnwzE8b3ShylPJrNl52p3rPsyajkbDHIoxQL43N1UXJ5f-IGvUeFLYgPTySsHa7fd-vvwRG-V8Pw_6v80vExpb41oDPBHtE-LXTvZ3FnI6WVj1T962mHh9kEIDBE9UF6uXQGNnI1Foh5TY5WUnaeNtGZycmWkrZhE42U7pkfT5D-ksRJodG5wjrPE6ivQZvnHJsB_XGAhzfT7ijFLeQg`
	ts, err := jwt.ParseWithClaims(token, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return pem, nil
	})

	if err != nil {
		t.Error(err)
	}
	t.Log(ts)
}
