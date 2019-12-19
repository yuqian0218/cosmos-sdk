package keys

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/multisig"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/cosmos-sdk/types"
)

// Keybase exposes operations on a generic keystore
type Keybase interface {
	// CRUD on the keystore
	List() ([]Info, error)
	// Get returns the public information about one key.
	Get(name string) (Info, error)
	// Get performs a by-address lookup and returns the public
	// information about one key if there's any.
	GetByAddress(address types.AccAddress) (Info, error)
	// Delete removes a key.
	Delete(name, passphrase string, skipPass bool) error
	// Sign bytes, looking up the private key to use.
	Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error)

	// CreateMnemonic generates a new mnemonic, derives a hierarchical deterministic
	// key from that. and persists it to storage, encrypted using the provided password.
	// It returns the generated mnemonic and the key Info. It returns an error if it fails to
	// generate a key for the given algo type, or if another key is already stored under the
	// same name.
	CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo) (info Info, seed string, err error)

	// CreateAccount converts a mnemonic to a private key using a BIP44 path 44'/118'/{account}'/0/{index}
	// and persists it, encrypted with the given password.
	CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error)

	// Derive computes a BIP39 seed from th mnemonic and bip39Passwd.
	// Derive private key from the seed using the BIP44 params.
	// Encrypt the key to disk using encryptPasswd.
	// See https://github.com/cosmos/cosmos-sdk/issues/2095
	Derive(name, mnemonic, bip39Passwd, encryptPasswd string, params hd.BIP44Params) (Info, error)

	// CreateLedger creates, stores, and returns a new Ledger key reference
	CreateLedger(name string, algo SigningAlgo, hrp string, account, index uint32) (info Info, err error)

	// CreateOffline creates, stores, and returns a new offline key reference
	CreateOffline(name string, pubkey crypto.PubKey) (info Info, err error)

	// CreateMulti creates, stores, and returns a new multsig (offline) key reference
	CreateMulti(name string, pubkey crypto.PubKey) (info Info, err error)

	// The following operations will *only* work on locally-stored keys
	Update(name, oldpass string, getNewpass func() (string, error)) error

	// Import imports ASCII armored Info objects.
	Import(name string, armor string) (err error)

	// ImportPrivKey imports a private key in ASCII armor format.
	// It returns an error if a key with the same name exists or a wrong encryption passphrase is
	// supplied.
	ImportPrivKey(name, armor, passphrase string) error

	// ImportPubKey imports ASCII-armored public keys.
	// Store a new Info object holding a public key only, i.e. it will
	// not be possible to sign with it as it lacks the secret key.
	ImportPubKey(name string, armor string) (err error)

	// Export exports an Info object in ASCII armored format.
	Export(name string) (armor string, err error)

	// ExportPubKey returns public keys in ASCII armored format.
	// Retrieve a Info object by its name and return the public key in
	// a portable format.
	ExportPubKey(name string) (armor string, err error)

	// ExportPrivKey returns a private key in ASCII armored format.
	// It returns an error if the key does not exist or a wrong encryption passphrase is supplied.
	ExportPrivKey(name, decryptPassphrase, encryptPassphrase string) (armor string, err error)

	// ExportPrivateKeyObject *only* works on locally-stored keys. Temporary method until we redo the exporting API
	ExportPrivateKeyObject(name string, passphrase string) (crypto.PrivKey, error)

	// CloseDB closes the database.
	CloseDB()
}

// KeyType reflects a human-readable type for key listing.
type KeyType uint

// Info KeyTypes
const (
	TypeLocal   KeyType = 0
	TypeLedger  KeyType = 1
	TypeOffline KeyType = 2
	TypeMulti   KeyType = 3
)

var keyTypes = map[KeyType]string{
	TypeLocal:   "local",
	TypeLedger:  "ledger",
	TypeOffline: "offline",
	TypeMulti:   "multi",
}

// String implements the stringer interface for KeyType.
func (kt KeyType) String() string {
	return keyTypes[kt]
}

// Info is the publicly exposed information about a keypair
type Info interface {
	// Human-readable type for key listing
	GetType() KeyType
	// Name of the key
	GetName() string
	// Public key
	GetPubKey() crypto.PubKey
	// Address
	GetAddress() types.AccAddress
	// Bip44 Path
	GetPath() (*hd.BIP44Params, error)
}

var (
	_ Info = &localInfo{}
	_ Info = &ledgerInfo{}
	_ Info = &offlineInfo{}
	_ Info = &multiInfo{}
)

// localInfo is the public information about a locally stored key
type localInfo struct {
	Name         string        `json:"name"`
	PubKey       crypto.PubKey `json:"pubkey"`
	PrivKeyArmor string        `json:"privkey.armor"`
}

func newLocalInfo(name string, pub crypto.PubKey, privArmor string) Info {
	return &localInfo{
		Name:         name,
		PubKey:       pub,
		PrivKeyArmor: privArmor,
	}
}

// GetType implements Info interface
func (i localInfo) GetType() KeyType {
	return TypeLocal
}

// GetType implements Info interface
func (i localInfo) GetName() string {
	return i.Name
}

// GetType implements Info interface
func (i localInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

// GetType implements Info interface
func (i localInfo) GetAddress() types.AccAddress {
	return i.PubKey.Address().Bytes()
}

// GetType implements Info interface
func (i localInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// ledgerInfo is the public information about a Ledger key
type ledgerInfo struct {
	Name   string         `json:"name"`
	PubKey crypto.PubKey  `json:"pubkey"`
	Path   hd.BIP44Params `json:"path"`
}

func newLedgerInfo(name string, pub crypto.PubKey, path hd.BIP44Params) Info {
	return &ledgerInfo{
		Name:   name,
		PubKey: pub,
		Path:   path,
	}
}

// GetType implements Info interface
func (i ledgerInfo) GetType() KeyType {
	return TypeLedger
}

// GetName implements Info interface
func (i ledgerInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i ledgerInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

// GetAddress implements Info interface
func (i ledgerInfo) GetAddress() types.AccAddress {
	return i.PubKey.Address().Bytes()
}

// GetPath implements Info interface
func (i ledgerInfo) GetPath() (*hd.BIP44Params, error) {
	tmp := i.Path
	return &tmp, nil
}

// offlineInfo is the public information about an offline key
type offlineInfo struct {
	Name   string        `json:"name"`
	PubKey crypto.PubKey `json:"pubkey"`
}

func newOfflineInfo(name string, pub crypto.PubKey) Info {
	return &offlineInfo{
		Name:   name,
		PubKey: pub,
	}
}

// GetType implements Info interface
func (i offlineInfo) GetType() KeyType {
	return TypeOffline
}

// GetName implements Info interface
func (i offlineInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i offlineInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

// GetAddress implements Info interface
func (i offlineInfo) GetAddress() types.AccAddress {
	return i.PubKey.Address().Bytes()
}

// GetPath implements Info interface
func (i offlineInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

type multisigPubKeyInfo struct {
	PubKey crypto.PubKey `json:"pubkey"`
	Weight uint          `json:"weight"`
}

// multiInfo is the public information about a multisig key
type multiInfo struct {
	Name      string               `json:"name"`
	PubKey    crypto.PubKey        `json:"pubkey"`
	Threshold uint                 `json:"threshold"`
	PubKeys   []multisigPubKeyInfo `json:"pubkeys"`
}

// NewMultiInfo creates a new multiInfo instance
func NewMultiInfo(name string, pub crypto.PubKey) Info {
	multiPK := pub.(multisig.PubKeyMultisigThreshold)

	pubKeys := make([]multisigPubKeyInfo, len(multiPK.PubKeys))
	for i, pk := range multiPK.PubKeys {
		// TODO: Recursively check pk for total weight?
		pubKeys[i] = multisigPubKeyInfo{pk, 1}
	}

	return &multiInfo{
		Name:      name,
		PubKey:    pub,
		Threshold: multiPK.K,
		PubKeys:   pubKeys,
	}
}

// GetType implements Info interface
func (i multiInfo) GetType() KeyType {
	return TypeMulti
}

// GetName implements Info interface
func (i multiInfo) GetName() string {
	return i.Name
}

// GetPubKey implements Info interface
func (i multiInfo) GetPubKey() crypto.PubKey {
	return i.PubKey
}

// GetAddress implements Info interface
func (i multiInfo) GetAddress() types.AccAddress {
	return i.PubKey.Address().Bytes()
}

// GetPath implements Info interface
func (i multiInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, fmt.Errorf("BIP44 Paths are not available for this type")
}

// encoding info
func marshalInfo(i Info) []byte {
	return CryptoCdc.MustMarshalBinaryLengthPrefixed(i)
}

// decoding info
func unmarshalInfo(bz []byte) (info Info, err error) {
	err = CryptoCdc.UnmarshalBinaryLengthPrefixed(bz, &info)
	return
}

type (
	// PrivKeyGenFunc defines the function to convert derived key bytes to a tendermint private key
	PrivKeyGenFunc func(bz [32]byte) crypto.PrivKey

	// KeybaseOption overrides options for the db
	KeybaseOption func(*kbOptions)
)
