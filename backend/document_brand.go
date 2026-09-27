package main

import (
	_ "embed"
	"encoding/base64"
	"strings"
)

//go:embed assets/company.png
var companyLogo []byte
var companyLogoBase64 = base64.StdEncoding.EncodeToString(companyLogo)

func partnerAccount(m Mitra) string {
	if strings.TrimSpace(m.Rekening) == "" {
		return "Rekening mitra belum diisi — lengkapi sebelum pembayaran."
	}
	return m.Rekening
}
