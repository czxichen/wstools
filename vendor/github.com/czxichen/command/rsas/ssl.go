package rsas

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"io/ioutil"
	"math/big"
	rd "math/rand"
	"net"
	"os"
	"time"
)

//var Certificate = struct {
//	RootCA  *x509.Certificate
//	RootKey
//}{}
//func InitRootCA(crt, key string) (err error) {
//	Certificate.RootCA, Certificate.RootKey, err = Parse(crt, key)
//	return
//}
/*
	x := rsas.CertInformation{
		Country:            []string{"CH"},
		Organization:       []string{"www.work-stacks.com"},
		OrganizationalUnit: []string{"Paas"},
		EmailAddress:       []string{"czxichen@163.com"},
		Province:           []string{"JS"},
		Locality:           []string{"SZ"},
		CommonName:         "master.work-stacks.com",
		DNSNames:           []string{"master.work-stacks.com"},
		EncryptLen: 512,
		IsCA:       true,
		DateLen:    5,
	}
*/

func init() {
	rd.Seed(time.Now().UnixNano())
}
func GetDefaultCrtInfo() CertInformation {
	return defaultCrtInfo
}

var defaultCrtInfo = CertInformation{
	Country:            []string{"CH"},
	Organization:       []string{"www.work-stacks.com"},
	OrganizationalUnit: []string{"Paas"},
	EmailAddress:       []string{"czxichen@163.com"},
	Province:           []string{"JiangSu"},
	Locality:           []string{"SuZhou"},
	CommonName:         ".work-stacks.com",
	IPAddresses:        []net.IP{net.ParseIP("127.0.0.1")},
	EncryptLen:         1024,
	IsCA:               false,
	DateLen:            5,
}

type CertInformation struct {
	Country            []string
	Organization       []string
	OrganizationalUnit []string //使用者
	EmailAddress       []string
	Province           []string //省
	Locality           []string //市
	CommonName         string   //域名
	DNSNames           []string
	IPAddresses        []net.IP
	IsCA               bool //是否是根证书
	Names              []pkix.AttributeTypeAndValue
	EncryptLen         int //密钥长度
	DateLen            int //有效期,单位年
}

func SignerCRT(rootcrt *x509.Certificate, rootkey *rsa.PrivateKey, crt *x509.Certificate) ([]byte, error) {
	if rootcrt == nil || rootkey == nil {
		return nil, errors.New("Root crt is null")
	}

	buf, err := x509.CreateCertificate(rand.Reader, crt, rootcrt, crt.PublicKey, rootkey)
	b := bytes.NewBuffer(nil)
	err = Write(b, buf, "CERTIFICATE")
	return b.Bytes(), err
}

func SignerCRTFromFile(rc, rk, ac, oc string) error {
	crt, key, err := Parse(rc, rk)
	if err != nil {
		return err
	}
	buf, err := ioutil.ReadFile(ac)
	if err != nil {
		return err
	}
	acrt, err := ParseCrt(buf)
	if err != nil {
		return err
	}
	buf, err = SignerCRT(crt, key, acrt)
	if err != nil {
		return err
	}
	File, err := os.Create(oc)
	if err != nil {
		return err
	}
	File.Write(buf)
	return File.Close()
}

func CheckSignature(rootcrt *x509.Certificate, crt []byte) error {
	ca, err := ParseCrt(crt)
	if err != nil {
		return err
	}
	return ca.CheckSignatureFrom(rootcrt)
}

func CreatePemCRT(RootCa *x509.Certificate, RootKey *rsa.PrivateKey, info CertInformation) (pemcrt []byte, pemkey []byte, err error) {
	pemcrt, pemkey, err = CreateCRT(RootCa, RootKey, info)
	if err != nil {
		return
	}

	cFile := bytes.NewBuffer([]byte{})
	err = Write(cFile, pemcrt, "CERTIFICATE")
	if err != nil {
		return
	}
	pemcrt = cFile.Bytes()

	kFile := bytes.NewBuffer([]byte{})
	err = Write(kFile, pemkey, "PRIVATE KEY")
	pemkey = kFile.Bytes()
	return
}

func CreateCRT(RootCa *x509.Certificate, RootKey *rsa.PrivateKey, info CertInformation) (crt []byte, key []byte, err error) {
	Crt := newCertificate(info)
	if info.EncryptLen < 512 {
		info.EncryptLen = 512
	}

	Key, err := rsa.GenerateKey(rand.Reader, info.EncryptLen)
	if err != nil {
		return
	}

	key = x509.MarshalPKCS1PrivateKey(Key)
	if RootCa == nil || RootKey == nil {
		crt, err = x509.CreateCertificate(rand.Reader, Crt, Crt, &Key.PublicKey, Key)
	} else {
		crt, err = x509.CreateCertificate(rand.Reader, Crt, RootCa, &Key.PublicKey, RootKey)
	}
	return
}

func WirteFile(path string, buf []byte, typ string) error {
	//	os.MkdirAll(filepath.Dir(path), 0666)
	File, err := os.Create(path)
	defer File.Close()

	if err != nil {
		return err
	}
	return Write(File, buf, typ)
}

func Write(w io.Writer, buf []byte, typ string) error {
	b := &pem.Block{Bytes: buf, Type: typ}
	return pem.Encode(w, b)
}

func Parse(crtPath, keyPath string) (rootcertificate *x509.Certificate, rootPrivateKey *rsa.PrivateKey, err error) {
	buf, err := ioutil.ReadFile(crtPath)
	if err != nil {
		return
	}
	rootcertificate, err = ParseCrt(buf)
	if err != nil {
		return
	}

	buf, err = ioutil.ReadFile(keyPath)
	if err != nil {
		return
	}
	rootPrivateKey, err = ParseKey(buf)
	return
}

func ParseCrt(buf []byte) (*x509.Certificate, error) {
	p := &pem.Block{}
	p, _ = pem.Decode(buf)
	return x509.ParseCertificate(p.Bytes)
}

func ParseKey(buf []byte) (*rsa.PrivateKey, error) {
	p, buf := pem.Decode(buf)
	return x509.ParsePKCS1PrivateKey(p.Bytes)
}

func ReadBlock(path string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p := &pem.Block{}
	p, _ = pem.Decode(buf)
	return p.Bytes, nil
}

func newCertificate(info CertInformation) *x509.Certificate {
	if info.DateLen == 0 {
		info.DateLen = 10
	}
	return &x509.Certificate{
		SerialNumber: big.NewInt(rd.Int63()),
		Subject: pkix.Name{
			Country:            info.Country,
			Organization:       info.Organization,
			OrganizationalUnit: info.OrganizationalUnit,
			Province:           info.Province,
			CommonName:         info.CommonName,
			Locality:           info.Locality,
			ExtraNames:         info.Names,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(info.DateLen, 0, 0),
		BasicConstraintsValid: true,
		DNSNames:              info.DNSNames,
		IPAddresses:           info.IPAddresses,
		IsCA:                  info.IsCA,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		EmailAddresses:        info.EmailAddress,
	}
}
