package rsas

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"os"
	"testing"
)

func Test_crt(t *testing.T) {
	baseinfo := CertInformation{Country: []string{"CN"}, Organization: []string{"WS"}, IsCA: true,
		OrganizationalUnit: []string{"work-stacks"}, EmailAddress: []string{"czxichen@163.com"},
		Locality: []string{"SuZhou"}, Province: []string{"JiangSu"}, CommonName: "Work-stacks", EncryptLen: 2048}

	rootca, rootkey, err := CreatePemCRT(nil, nil, baseinfo)
	if err != nil {
		t.Log("Create crt error,Error info:", err)
		return
	}
	baseinfo.CommonName = "localhost"
	baseinfo.Names = []pkix.AttributeTypeAndValue{{asn1.ObjectIdentifier{2, 1, 3}, "MAC_ADDR"}}

	os.Stdout.WriteString(string(rootca))
	os.Stdout.WriteString(string(rootkey))
	crt, err := ParseCrt(rootca)
	if err != nil {
		t.Log("Parse crt error,Error info:", err)
		return
	}

	pri, err := ParseKey(rootkey)
	if err != nil {
		t.Log("Parse key error,Error info:", err)
		return
	}
	rootca, rootkey, err = CreateCRT(crt, pri, baseinfo)
	if err != nil {
		t.Log("Create crt error,Error info:", err)
		return
	}
	Write(os.Stdout, rootca, "CERTIFICATE")
	Write(os.Stdout, rootkey, "PRIVATE KEY")
}
