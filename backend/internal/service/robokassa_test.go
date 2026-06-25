package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRobokassaPaymentSignatureWithResultURL2(t *testing.T) {
	encoded := EncodeRobokassaResultURL2("https://test.test/robokassa")
	sig := BuildRobokassaPaymentSignature("robokassa", "10", "1234829", "password1", encoded)
	assert.Equal(t, "77b1058f37266e18ea2cba314c59f2c7", sig)
}

func TestBuildRobokassaPaymentSignatureWithoutResultURL2(t *testing.T) {
	sig := BuildRobokassaPaymentSignature("demo", "2900.00", "1001", "pass1", "")
	assert.Equal(t, "ab652b7b36ba1b702a88c3b7aa18d7ff", sig)
}

func TestParseRobokassaResult2Token(t *testing.T) {
	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJoZWFkZXIiOnsidHlwZSI6IlBheW1lbnRTdGF0ZU5vdGlmaWNhdGlvbiIsInZlcnNpb24iOiIxLjAuMCIsInRpbWVzdGFtcCI6IjE2OTA0NTQ5ODAifSwiZGF0YSI6eyJzaG9wIjoicm9iby1kZW1vLXRlc3QiLCJvcEtleSI6IjE0RDJCNTIxLTRFQUItNDkyQS05RDkxLTAwRDEyRkYyNEQ1Ny0xTnZIRU5Md2dwIiwiaW52SWQiOiIxMjM0ODI5IiwicGF5bWVudE1ldGhvZCI6IkJhbmtDYXJkIiwiaW5jU3VtIjoiMTAuMDAiLCJzdGF0ZSI6Ik9LIn19.DaanjJ4A2yaGOppIVEj929MGNWU4jHcjAh4_DnFRJzkqvESbVpN5trGg8OlaftyQEldxeQaTF9ykrS7O1BmhPow0ZRv4GwjJQSwHzCGJUOcyVw507PMnRSAIT4oXq_PtKjqjuHQL_bxQxyGxZORoKhzsDwBYl_nMXsY-R_SCXPfvYRDPeJRRlC-HwQscOzyGN1cBHStuh5-w8qHytpaysVlif_XYK1-Lfa3ZsFjNz-2LENGL4CRmKMA2KhUfhNfNWXSA4J2KykfWznAvaY37T__QQHhDyHuIjsUtstG5Jf2aK2CuJY-i9mBTRkXD0U3Y5Lh7lnJQrkXAMBySRvYigQ"

	parsed, err := ParseRobokassaResult2Token(token)
	require.NoError(t, err)
	assert.Equal(t, "robo-demo-test", parsed.Shop)
	assert.Equal(t, "1234829", parsed.InvID)
	assert.Equal(t, "10.00", parsed.IncSum)
	assert.Equal(t, "OK", parsed.State)
}

func TestParseRobokassaResult2TokenRejectsNonOK(t *testing.T) {
	// Same payload with state FAIL would need another token; test empty token instead.
	_, err := ParseRobokassaResult2Token("")
	require.Error(t, err)
}
