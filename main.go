package main

import (
	"context"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/alecthomas/kong"
	"golang.org/x/exp/maps"
	"os"
	"sort"
	"storj.io/common/base58"
	"storj.io/common/identity"
	"storj.io/common/peertls/tlsopts"
	"storj.io/common/rpc"
	"storj.io/common/storj"
	"storj.io/drpc"
	"strings"
	"time"
	"unicode/utf8"
)

func main() {
	ktx := kong.Parse(&Main{})
	err := ktx.Run()
	if err != nil {
		panic(err)
	}
}

var base32Encoding = base32.StdEncoding.WithPadding(base32.NoPadding)
var pathEncoding = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

type Main struct {
	Source string `arg:"" help:"the encoded bytes to be decoded"`
	From   string `help:"define the format of the source bytea"`
	To     string `help:"pick the used destination format"`
	NL     bool   `help:"by default if --from and --to are set, new line is not added (to make it easier to use in scripts). This flag adds the new line back"`
}

func (m Main) Run() error {
	source := m.Source

	decodings := map[string]func(string) ([]byte, error){
		"remote-id": getRemoteID,
		"base58": func(s string) ([]byte, error) {
			result, _, err := base58.CheckDecode(s)
			return result, err
		},
		"base32": base32Encoding.DecodeString,
		"auth-base32": func(s string) ([]byte, error) {
			return base32Encoding.DecodeString(strings.ToUpper(s))
		},
		"base64":  base64.URLEncoding.DecodeString,
		"base64s": base64.StdEncoding.DecodeString,
		"hex":     hex.DecodeString,
		"path":    pathEncoding.DecodeString,
		"file":    readFromFile,
	}

	encodings := map[string]func([]byte) string{
		"hex":    hex.EncodeToString,
		"base58": base58.Encode,
		"nodeid": func(bytes []byte) string {
			fromBytes, err := storj.NodeIDFromBytes(bytes)
			if err != nil {
				return ""
			}
			return fromBytes.String()
		},
		"nodeurl": func(bytes []byte) string {
			if !strings.Contains(source, ":") {
				return ""
			}
			fromBytes, err := storj.NodeIDFromBytes(bytes)
			if err != nil {
				return ""
			}
			return fromBytes.String() + "@" + source
		},
		"base32": base32Encoding.EncodeToString,
		"base64": base64.URLEncoding.EncodeToString,
		"path":   pathEncoding.EncodeToString,
		"string": func(bytes []byte) string {
			raw := string(bytes)
			if utf8.ValidString(raw) {
				return raw
			}
			return ""
		},
		"binary": func(bytes []byte) string {
			return string(bytes)
		},
	}

	keys := maps.Keys(decodings)
	sort.Strings(keys)
	for _, id := range keys {
		if m.From != "" && m.From != id {
			continue
		}
		decoded, err := decodings[id](source)
		if err == nil {
			if m.From == "" {
				fmt.Printf("Using %s as %s\n", source, id)
			}
			ekeys := maps.Keys(encodings)
			sort.Strings(ekeys)
			for _, ei := range ekeys {
				if m.To != "" && m.To != ei {
					continue
				}
				decoded := encodings[ei](decoded)
				if decoded != "" {
					if m.To != "" {
						fmt.Print(decoded)
						if m.NL {
							fmt.Println()
						}
					} else {
						fmt.Println("  ", ei, decoded)
					}

				}
			}

		}
	}

	return nil
}

func readFromFile(s string) ([]byte, error) {
	_, err := os.Stat(s)
	if err != nil {
		return nil, err
	}
	id, err := identity.NodeIDFromCertPath(s)
	if err != nil {
		return nil, err
	}
	return id.Bytes(), nil
}

func getRemoteID(hostPort string) ([]byte, error) {
	if !strings.Contains(hostPort, ":") {
		return nil, errors.New("doesn't look like a host:port")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	tlsOptions, err := createEphemeralTLSIdentity(ctx)
	if err != nil {
		return nil, err
	}

	dialer := rpc.NewDefaultDialer(tlsOptions)
	dialer.Pool = rpc.NewDefaultConnectionPool()

	dialer.DialTimeout = 10 * time.Second
	// TODO
	//dialContext := socket.BackgroundDialer().DialContext
	//
	////lint:ignore SA1019 it's safe to use TCP here instead of QUIC + TCP
	//dialer.Connector = rpc.NewDefaultTCPConnector(&rpc.ConnectorAdapter{DialContext: dialContext}) //nolint:staticcheck

	conn, err := dialer.DialAddressInsecure(ctx, hostPort)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	in := struct{}{}
	out := struct{}{}
	_ = conn.Invoke(ctx, "asd", &NullEncoding{}, in, out)
	peerIdentity, err := conn.PeerIdentity()
	if err != nil {
		return nil, err
	}

	return peerIdentity.ID.Bytes(), nil
}

func createEphemeralTLSIdentity(ctx context.Context) (*tlsopts.Options, error) {
	ident, err := identity.NewFullIdentity(ctx, identity.NewCAOptions{
		Difficulty:  0,
		Concurrency: 1,
	})
	if err != nil {
		return nil, err
	}

	tlsConfig := tlsopts.Config{
		UsePeerCAWhitelist: false,
		PeerIDVersions:     "0",
	}

	tlsOptions, err := tlsopts.NewOptions(ident, tlsConfig, nil)
	if err != nil {
		return nil, err
	}

	return tlsOptions, nil
}

type NullEncoding struct {
}

func (n NullEncoding) Marshal(msg drpc.Message) ([]byte, error) {
	return []byte{1}, nil
}

func (n NullEncoding) Unmarshal(buf []byte, msg drpc.Message) error {
	return nil
}

var _ drpc.Encoding = &NullEncoding{}
