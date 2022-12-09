package resource

import (
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/kong/koko/internal/crypto"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"google.golang.org/protobuf/proto"
)

const (
	TypeCertificate model.Type = "certificate"
)

func NewCertificate() Certificate {
	return Certificate{
		Certificate: &v1.Certificate{},
	}
}

type Certificate struct {
	Certificate *v1.Certificate
}

func (r Certificate) ID() string {
	if r.Certificate == nil {
		return ""
	}
	return r.Certificate.Id
}

func (r Certificate) Type() model.Type {
	return TypeCertificate
}

func (r Certificate) Resource() model.Resource {
	return r.Certificate
}

// SetResource implements the Object.SetResource interface.
func (r Certificate) SetResource(pr model.Resource) error { return model.SetResource(r, pr) }

func (r Certificate) Validate(ctx context.Context) error {
	err := validation.Validate(string(TypeCertificate), r.Certificate)
	if err != nil {
		return err
	}

	var cert *x509.Certificate
	var pubKey []byte

	if !isReference(r.Certificate.Cert) {
		cert, _ = crypto.ParsePEMCert([]byte(r.Certificate.Cert))
	}
	if !isReference(r.Certificate.Key) {
		pubKey, err = crypto.RetrievePublicFromPrivateKey([]byte(r.Certificate.Key))
		if err != nil {
			return validation.Error{
				Errs: []*v1.ErrorDetail{
					{
						Type:     v1.ErrorType_ERROR_TYPE_ENTITY,
						Messages: []string{fmt.Sprintf("failed to get public key from certificate: %v", err)},
					},
				},
			}
		}
	}
	if !isReference(r.Certificate.Cert) && !isReference(r.Certificate.Key) {
		certPubKey, _ := x509.MarshalPKIXPublicKey(cert.PublicKey)
		if !bytes.Equal(certPubKey, pubKey) {
			return validation.Error{
				Errs: []*v1.ErrorDetail{
					{
						Type:     v1.ErrorType_ERROR_TYPE_ENTITY,
						Messages: []string{"certificate does not match key"},
					},
				},
			}
		}
	}

	if r.Certificate.CertAlt == "" {
		return nil
	}

	var altCert *x509.Certificate
	var altPubKey []byte

	if !isReference(r.Certificate.CertAlt) {
		altCert, _ = crypto.ParsePEMCert([]byte(r.Certificate.CertAlt))
	}
	if !isReference(r.Certificate.Cert) && !isReference(r.Certificate.CertAlt) {
		if cert.PublicKeyAlgorithm == altCert.PublicKeyAlgorithm {
			return validation.Error{
				Errs: []*v1.ErrorDetail{
					{
						Type: v1.ErrorType_ERROR_TYPE_ENTITY,
						Messages: []string{fmt.Sprintf("certificate and alternative certificate need to have "+
							"different type (e.g. RSA and ECDSA), the provided "+
							"certificates were both of the same type '%s'", cert.PublicKeyAlgorithm.String())},
					},
				},
			}
		}
	}

	if !isReference(r.Certificate.KeyAlt) {
		altPubKey, err = crypto.RetrievePublicFromPrivateKey([]byte(r.Certificate.KeyAlt))
		if err != nil {
			return validation.Error{
				Errs: []*v1.ErrorDetail{
					{
						Type:     v1.ErrorType_ERROR_TYPE_ENTITY,
						Messages: []string{fmt.Sprintf("failed to get public key from alternate certificate: %v", err)},
					},
				},
			}
		}
	}
	if !isReference(r.Certificate.CertAlt) && !isReference(r.Certificate.KeyAlt) {
		altCertPubKey, _ := x509.MarshalPKIXPublicKey(altCert.PublicKey)
		if !bytes.Equal(altCertPubKey, altPubKey) {
			return validation.Error{
				Errs: []*v1.ErrorDetail{
					{
						Type:     v1.ErrorType_ERROR_TYPE_ENTITY,
						Messages: []string{"alternate certificate does not match key"},
					},
				},
			}
		}
	}
	return nil
}

func (r Certificate) ProcessDefaults(ctx context.Context) error {
	if r.Certificate == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&r.Certificate.Id)
	return nil
}

func (r Certificate) Indexes() []model.Index {
	return nil
}

// MarshalResourceJSON is used here to make sure that certificate
// metadata, which cannot be passed via API, are still included
// in the marshalled object.
//
// At the same time, we don't want to store the metadata object in the
// database, so here we are creating a clone of the original object and
// setting the metadata to nil. This way the original object, which
// includes the metadata, will be returned to the caller, while the
// clone object, which doesn't include the metadata, will be stored in
// the database.
func (r Certificate) MarshalResourceJSON() ([]byte, error) {
	if !isReference(r.Certificate.Cert) {
		if err := extractCertInfo(r.Certificate); err != nil {
			return nil, err
		}
	}
	newCert := proto.Clone(r.Certificate)
	newCertObj, ok := newCert.(*v1.Certificate)
	if !ok {
		return nil, errors.New("unexpected type while converting Certificate clone")
	}
	newCertObj.Metadata = nil
	return json.Marshal(r.Certificate)
}

// UnmarshalResourceJSON is used here to make sure that certificate
// metadata, which cannot be passed via API, are still included
// in the unmarshalled object.
func (r Certificate) UnmarshalResourceJSON(cert []byte) error {
	if err := json.Unmarshal(cert, &r.Certificate); err != nil {
		return err
	}
	if !isReference(r.Certificate.Cert) {
		if err := extractCertInfo(r.Certificate); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	err := model.RegisterType(TypeCertificate, &v1.Certificate{}, func() model.Object {
		return NewCertificate()
	})
	if err != nil {
		panic(err)
	}

	certificateSchema := &generator.Schema{
		Properties: map[string]*generator.Schema{
			"id": typedefs.ID,
			"cert": {
				Type: "string",
				AnyOf: []*generator.Schema{
					{Format: "pem-encoded-cert"},
					typedefs.Reference,
				},
				XKokoConfig: &extension.Config{
					DisableValidateEndpoint: true,
					Referenceable:           true,
				},
			},
			"key": {
				Type: "string",
				AnyOf: []*generator.Schema{
					{Format: "pem-encoded-private-key"},
					typedefs.Reference,
				},
				XKokoConfig: &extension.Config{
					DisableValidateEndpoint: true,
					Referenceable:           true,
				},
			},
			"cert_alt": {
				Type: "string",
				AnyOf: []*generator.Schema{
					{Format: "pem-encoded-cert"},
					typedefs.Reference,
				},
				XKokoConfig: &extension.Config{
					DisableValidateEndpoint: true,
					Referenceable:           true,
				},
			},
			"key_alt": {
				Type: "string",
				AnyOf: []*generator.Schema{
					{Format: "pem-encoded-private-key"},
					typedefs.Reference,
				},
				XKokoConfig: &extension.Config{
					DisableValidateEndpoint: true,
					Referenceable:           true,
				},
			},
			"tags":       typedefs.Tags,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
		},
		Required: []string{
			"id",
			"cert",
			"key",
		},
		Dependencies: map[string]*generator.Schema{
			"cert_alt": {
				Required: []string{
					"key_alt",
				},
			},
			"key_alt": {
				Required: []string{
					"cert_alt",
				},
			},
		},
		AdditionalProperties: &falsy,
		XKokoConfig: &extension.Config{
			ResourceAPIPath: "certificates",
		},
	}
	err = generator.DefaultRegistry.Register(string(TypeCertificate), certificateSchema)
	if err != nil {
		panic(err)
	}
}

func extractKeyUsages(cert *x509.Certificate) []v1.KeyUsageType {
	keyUsages := []v1.KeyUsageType{}
	if cert.KeyUsage&x509.KeyUsageDigitalSignature != 0 {
		keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_DIGITAL_SIGNATURE)
	}
	if cert.KeyUsage&x509.KeyUsageContentCommitment != 0 {
		keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_CONTENT_COMMITMENT)
	}
	if cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
		keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_KEY_ENCIPHERMENT)
	}
	if cert.KeyUsage&x509.KeyUsageDataEncipherment != 0 {
		keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_DATA_ENCIPHERMENT)
	}
	if cert.KeyUsage&x509.KeyUsageKeyAgreement != 0 {
		keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_KEY_AGREEMENT)
	}
	if cert.KeyUsage&x509.KeyUsageCertSign != 0 {
		keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_KEY_CERT_SIGN)
	}
	if cert.KeyUsage&x509.KeyUsageCRLSign != 0 {
		keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_CRL_SIGN)
	}
	if x509.KeyUsageEncipherOnly != 0 {
		keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_ENCIPHER_ONLY)
	}
	if cert.KeyUsage&x509.KeyUsageDecipherOnly != 0 {
		keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_DECIPHER_ONLY)
	}

	for _, ext := range cert.ExtKeyUsage {
		switch ext {
		case x509.ExtKeyUsageAny:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_ANY)
		case x509.ExtKeyUsageServerAuth:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_SERVER_AUTH)
		case x509.ExtKeyUsageClientAuth:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_CLIENT_AUTH)
		case x509.ExtKeyUsageCodeSigning:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_CODE_SIGNING)
		case x509.ExtKeyUsageEmailProtection:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_EMAIL_PROTECTION)
		case x509.ExtKeyUsageIPSECEndSystem:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_IPSEC_END_SYSTEM)
		case x509.ExtKeyUsageIPSECTunnel:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_IPSEC_TUNNEL)
		case x509.ExtKeyUsageIPSECUser:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_IPSEC_USER)
		case x509.ExtKeyUsageTimeStamping:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_TIME_STAMPING)
		case x509.ExtKeyUsageOCSPSigning:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_OSCP_SIGNING)
		case x509.ExtKeyUsageMicrosoftServerGatedCrypto:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_MICROSOFT_SERVER_GATED_CRYPTO)
		case x509.ExtKeyUsageNetscapeServerGatedCrypto:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_NETSCAPE_SERVER_GATED_CRYPTO)
		case x509.ExtKeyUsageMicrosoftCommercialCodeSigning:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_MICROSOFT_COMMERCIAL_CODE_SIGNING)
		case x509.ExtKeyUsageMicrosoftKernelCodeSigning:
			keyUsages = append(keyUsages, v1.KeyUsageType_KEY_USAGE_TYPE_MICROSOFT_KERNEL_CODE_SIGNING)
		}
	}
	return keyUsages
}

func extractCertInfo(cert *v1.Certificate) error {
	decoded, err := crypto.ParsePEMCert([]byte(cert.Cert))
	if err != nil {
		return fmt.Errorf("decoding certificate: %w", err)
	}
	cert.Metadata = &v1.CertificateMetadata{
		Issuer:    decoded.Issuer.String(),
		Subject:   decoded.Subject.String(),
		KeyUsages: extractKeyUsages(decoded),
		Expiry:    int32(decoded.NotAfter.Unix()),
		SanNames:  decoded.DNSNames,
	}
	return nil
}
