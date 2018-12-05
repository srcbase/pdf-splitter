/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

// Test the PDF crypt support.

package core

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/unidoc/unidoc/common"
)

func init() {
	common.SetLogger(common.ConsoleLogger{})
}

func TestPadding(t *testing.T) {
	crypter := PdfCrypt{}

	// Case 1 empty pass, should match padded string.
	key := crypter.paddedPass([]byte(""))
	if len(key) != 32 {
		t.Errorf("Fail, expected padded pass length = 32 (%d)", len(key))
	}
	if key[0] != 0x28 {
		t.Errorf("key[0] != 0x28 (%q in %q)", key[0], key)
	}
	if key[31] != 0x7A {
		t.Errorf("key[31] != 0x7A (%q in %q)", key[31], key)
	}

	// Case 2, non empty pass.
	key = crypter.paddedPass([]byte("bla"))
	if len(key) != 32 {
		t.Errorf("Fail, expected padded pass length = 32 (%d)", len(key))
	}
	if string(key[0:3]) != "bla" {
		t.Errorf("Expecting start with bla (%s)", key)
	}
	if key[3] != 0x28 {
		t.Errorf("key[3] != 0x28 (%q in %q)", key[3], key)
	}
	if key[31] != 0x64 {
		t.Errorf("key[31] != 0x64 (%q in %q)", key[31], key)
	}
}

// Test algorithm 2.
func TestAlg2(t *testing.T) {
	crypter := PdfCrypt{}
	crypter.V = 2
	crypter.R = 3
	crypter.P = -3904
	crypter.Id0 = string([]byte{0x4e, 0x00, 0x99, 0xe5, 0x36, 0x78, 0x93, 0x24,
		0xff, 0xd5, 0x82, 0xe4, 0xec, 0x0e, 0xa3, 0xb4})
	crypter.O = []byte{0xE6, 0x00, 0xEC, 0xC2, 0x02, 0x88, 0xAD, 0x8B,
		0x5C, 0x72, 0x64, 0xA9, 0x5C, 0x29, 0xC6, 0xA8, 0x3E, 0xE2, 0x51,
		0x76, 0x79, 0xAA, 0x02, 0x18, 0xBE, 0xCE, 0xEA, 0x8B, 0x79, 0x86,
		0x72, 0x6A, 0x8C, 0xDB}
	crypter.Length = 128
	crypter.EncryptMetadata = true

	key := crypter.Alg2([]byte(""))

	keyExp := []byte{0xf8, 0x94, 0x9c, 0x5a, 0xf5, 0xa0, 0xc0, 0xca,
		0x30, 0xb8, 0x91, 0xc1, 0xbb, 0x2c, 0x4f, 0xf5}

	if string(key) != string(keyExp) {
		common.Log.Debug("   Key (%d): % x", len(key), key)
		common.Log.Debug("KeyExp (%d): % x", len(keyExp), keyExp)
		t.Errorf("alg2 -> key != expected\n")
	}

}

// Test algorithm 3.
func TestAlg3(t *testing.T) {
	crypter := PdfCrypt{}
	crypter.V = 2
	crypter.R = 3
	crypter.P = -3904
	crypter.Id0 = string([]byte{0x4e, 0x00, 0x99, 0xe5, 0x36, 0x78, 0x93, 0x24,
		0xff, 0xd5, 0x82, 0xe4, 0xec, 0x0e, 0xa3, 0xb4})
	Oexp := []byte{0xE6, 0x00, 0xEC, 0xC2, 0x02, 0x88, 0xAD, 0x8B,
		0x0d, 0x64, 0xA9, 0x29, 0xC6, 0xA8, 0x3E, 0xE2, 0x51,
		0x76, 0x79, 0xAA, 0x02, 0x18, 0xBE, 0xCE, 0xEA, 0x8B, 0x79, 0x86,
		0x72, 0x6A, 0x8C, 0xDB}
	crypter.Length = 128
	crypter.EncryptMetadata = true

	O, err := crypter.Alg3([]byte(""), []byte("test"))
	if err != nil {
		t.Errorf("crypt alg3 error %s", err)
		return
	}

	if string(O) != string(Oexp) {
		common.Log.Debug("   O (%d): % x", len(O), O)
		common.Log.Debug("Oexp (%d): % x", len(Oexp), Oexp)
		t.Errorf("alg3 -> key != expected")
	}
}

// Test algorithm 5 for computing dictionary's U (user password) value
// valid for R >= 3.
func TestAlg5(t *testing.T) {
	crypter := PdfCrypt{}
	crypter.V = 2
	crypter.R = 3
	crypter.P = -3904
	crypter.Id0 = string([]byte{0x4e, 0x00, 0x99, 0xe5, 0x36, 0x78, 0x93, 0x24,
		0xff, 0xd5, 0x82, 0xe4, 0xec, 0x0e, 0xa3, 0xb4})
	crypter.O = []byte{0xE6, 0x00, 0xEC, 0xC2, 0x02, 0x88, 0xAD, 0x8B,
		0x5C, 0x72, 0x64, 0xA9, 0x5C, 0x29, 0xC6, 0xA8, 0x3E, 0xE2, 0x51,
		0x76, 0x79, 0xAA, 0x02, 0x18, 0xBE, 0xCE, 0xEA, 0x8B, 0x79, 0x86,
		0x72, 0x6A, 0x8C, 0xDB}
	crypter.Length = 128
	crypter.EncryptMetadata = true

	U, _, err := crypter.Alg5([]byte(""))
	if err != nil {
		t.Errorf("Error %s", err)
		return
	}

	Uexp := []byte{0x59, 0x66, 0x38, 0x6c, 0x76, 0xfe, 0x95, 0x7d, 0x3d,
		0x0d, 0x14, 0x3d, 0x36, 0xfd, 0x01, 0x3d, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	if string(U[0:16]) != string(Uexp[0:16]) {
		common.Log.Info("   U (%d): % x", len(U), U)
		common.Log.Info("Uexp (%d): % x", len(Uexp), Uexp)
		t.Errorf("U != expected\n")
	}
}

// Test decrypting. Example with V=2, R=3, using standard algorithm.
func TestDecryption1(t *testing.T) {
	crypter := PdfCrypt{}
	crypter.DecryptedObjects = map[PdfObject]bool{}
	// Default algorithm is V2 (RC4).
	crypter.CryptFilters = newCryptFiltersV2(crypter.Length)
	crypter.V = 2
	crypter.R = 3
	crypter.P = -3904
	crypter.Id0 = string([]byte{0x5f, 0x91, 0xff, 0xf2, 0x00, 0x88, 0x13,
		0x5f, 0x30, 0x24, 0xd1, 0x0f, 0x28, 0x31, 0xc6, 0xfa})
	crypter.O = []byte{0xE6, 0x00, 0xEC, 0xC2, 0x02, 0x88, 0xAD, 0x8B,
		0x0d, 0x64, 0xA9, 0x29, 0xC6, 0xA8, 0x3E, 0xE2, 0x51,
		0x76, 0x79, 0xAA, 0x02, 0x18, 0xBE, 0xCE, 0xEA, 0x8B, 0x79, 0x86,
		0x72, 0x6A, 0x8C, 0xDB}
	crypter.U = []byte{0xED, 0x5B, 0xA7, 0x76, 0xFD, 0xD8, 0xE3, 0x89,
		0x4F, 0x54, 0x05, 0xC1, 0x3B, 0xFD, 0x86, 0xCF, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00}
	crypter.Length = 128
	crypter.EncryptMetadata = true

	streamData := []byte{0xBC, 0x89, 0x86, 0x8B, 0x3E, 0xCF, 0x24, 0x1C,
		0xC4, 0x88, 0xF3, 0x60, 0x74, 0x8A, 0x22, 0xE3, 0xAD, 0xF4, 0x48,
		0x8E, 0x20, 0x94, 0x06, 0x4B, 0x4B, 0xB5, 0x3E, 0x93, 0x89, 0x4E,
		0x32, 0x38, 0xB4, 0xF6, 0x05, 0x3C, 0x5D, 0x0C, 0x12, 0xE4, 0xEB,
		0x9B, 0x8D, 0x26, 0x32, 0x7B, 0x09, 0x97, 0xA1, 0xC5, 0x98, 0xF6,
		0xE7, 0x1C, 0x3B}

	// Plain text stream (hello world).
	exp := []byte{0x20, 0x20, 0x42, 0x54, 0x0A, 0x20, 0x20, 0x20, 0x20,
		0x2F, 0x46, 0x31, 0x20, 0x31, 0x38, 0x20, 0x54, 0x66, 0x0A, 0x20,
		0x20, 0x20, 0x20, 0x30, 0x20, 0x30, 0x20, 0x54, 0x64, 0x0A, 0x20,
		0x20, 0x20, 0x20, 0x28, 0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57,
		0x6F, 0x72, 0x6C, 0x64, 0x29, 0x20, 0x54, 0x6A, 0x0A, 0x20, 0x20,
		0x45, 0x54}
	rawText := "2 0 obj\n<< /Length 55 >>\nstream\n" + string(streamData) + "\nendstream\n"

	parser := PdfParser{}
	parser.xrefs = make(XrefTable)
	parser.objstms = make(ObjectStreams)
	parser.rs, parser.reader, parser.fileSize = makeReaderForText(rawText)
	parser.crypter = &crypter

	obj, err := parser.ParseIndirectObject()
	if err != nil {
		t.Errorf("Error parsing object")
		return
	}

	so, ok := obj.(*PdfObjectStream)
	if !ok {
		t.Errorf("Should be stream (is %q)", obj)
		return
	}

	authenticated, err := parser.Decrypt([]byte(""))
	if err != nil {
		t.Errorf("Error authenticating")
		return
	}
	if !authenticated {
		t.Errorf("Failed to authenticate")
		return
	}

	parser.crypter.Decrypt(so, 0, 0)
	if string(so.Stream) != string(exp) {
		t.Errorf("Stream content wrong")
		return
	}
}

func BenchmarkAlg2b(b *testing.B) {
	// hash runs a variable number of rounds, so we need to have a
	// deterministic random source to make benchmark results comparable
	r := rand.New(rand.NewSource(1234567))
	const n = 20
	pass := make([]byte, n)
	r.Read(pass)
	data := make([]byte, n+8+48)
	r.Read(data)
	user := make([]byte, 48)
	r.Read(user)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = alg2b(data, pass, user)
	}
}

func TestAESv3(t *testing.T) {
	const keySize = 32

	seed := time.Now().UnixNano()
	rand := rand.New(rand.NewSource(seed))

	var cases = []struct {
		Name      string
		EncMeta   bool
		UserPass  string
		OwnerPass string
	}{
		{
			Name: "simple", EncMeta: true,
			UserPass: "user", OwnerPass: "owner",
		},
		{
			Name: "utf8", EncMeta: false,
			UserPass: "æøå-u", OwnerPass: "æøå-o",
		},
		{
			Name: "long", EncMeta: true,
			UserPass:  strings.Repeat("user", 80),
			OwnerPass: strings.Repeat("owner", 80),
		},
	}

	const (
		perms = 0x12345678
	)

	for _, R := range []int{5, 6} {
		R := R
		t.Run(fmt.Sprintf("R=%d", R), func(t *testing.T) {
			for _, c := range cases {
				c := c
				t.Run(c.Name, func(t *testing.T) {
					fkey := make([]byte, keySize)
					rand.Read(fkey)

					crypt := &PdfCrypt{
						V: 5, R: R,
						P:               perms,
						EncryptionKey:   append([]byte{}, fkey...),
						EncryptMetadata: c.EncMeta,
					}

					// generate encryption parameters
					err := crypt.generateR6([]byte(c.UserPass), []byte(c.OwnerPass))
					if err != nil {
						t.Fatal("Failed to encrypt:", err)
					}

					// Perms and EncryptMetadata are checked as a part of alg2a

					// decrypt using user password
					crypt.EncryptionKey = nil
					ok, err := crypt.alg2a([]byte(c.UserPass))
					if err != nil || !ok {
						t.Error("Failed to authenticate user pass:", err)
					} else if !bytes.Equal(crypt.EncryptionKey, fkey) {
						t.Error("wrong encryption key")
					}

					// decrypt using owner password
					crypt.EncryptionKey = nil
					ok, err = crypt.alg2a([]byte(c.OwnerPass))
					if err != nil || !ok {
						t.Error("Failed to authenticate owner pass:", err)
					} else if !bytes.Equal(crypt.EncryptionKey, fkey) {
						t.Error("wrong encryption key")
					}

					// try to elevate user permissions
					crypt.P = math.MaxUint32

					crypt.EncryptionKey = nil
					ok, err = crypt.alg2a([]byte(c.UserPass))
					if R == 5 {
						// it's actually possible with R=5, since Perms is not generated
						if err != nil || !ok {
							t.Error("Failed to authenticate user pass:", err)
						}
					} else {
						// not possible in R=6, should return an error
						if err == nil || ok {
							t.Error("was able to elevate permissions with R=6")
						}
					}
				})
			}
		})
	}
}