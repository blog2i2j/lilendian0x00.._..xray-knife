package xray

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/lilendian0x00/xray-knife/v2/pkg/protocol"
	"github.com/xtls/xray-core/infra/conf"
	"net/url"
	"reflect"
	"strings"
)

func NewWireguard(link string) Protocol {
	return &Wireguard{OrigLink: link}
}

func (w *Wireguard) Name() string {
	return "wireguard"
}

func (w *Wireguard) Parse() error {
	if !strings.HasPrefix(w.OrigLink, protocol.WireguardIdentifier) {
		return fmt.Errorf("wireguard unreconized: %s", w.OrigLink)
	}

	uri, err := url.Parse(w.OrigLink)
	if err != nil {
		return err
	}

	unescapedSecretKey, err0 := url.PathUnescape(uri.User.String())
	if err0 != nil {
		return err0
	}

	w.SecretKey = unescapedSecretKey

	w.Endpoint = uri.Host

	// Get the type of the struct
	t := reflect.TypeOf(*w)

	// Get the number of fields in the struct
	numFields := t.NumField()

	// Iterate over each field of the struct
	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")

		// If the query value exists for the field, set it
		if values, ok := uri.Query()[tag]; ok {
			value := values[0]
			v := reflect.ValueOf(w).Elem().FieldByName(field.Name)

			switch v.Type().String() {
			case "string":
				v.SetString(value)
			case "int32":
				var intValue int
				fmt.Sscanf(value, "%d", &intValue)
				v.SetInt(int64(intValue))

			}
		}
	}

	w.Remark, err = url.PathUnescape(uri.Fragment)
	if err != nil {
		w.Remark = uri.Fragment
	}

	return nil
}

func (w *Wireguard) DetailsStr() string {
	info := fmt.Sprintf("%s: %s\n%s: %s\n%s: %s\n%s: %d\n%s: %s\n%s: %v\n%s: %s\n", w.Name(),
		color.RedString("Protocol"),
		color.RedString("Remark"), w.Remark,
		color.RedString("Endpoint"), w.Endpoint,
		color.RedString("MTU"), w.Mtu,
		color.RedString("Local Addresses"), w.LocalAddress,
		color.RedString("Public Key"), w.PublicKey,
		color.RedString("Secret Key"), w.SecretKey,
	)

	return info
}

func (w *Wireguard) ConvertToGeneralConfig() (g protocol.GeneralConfig) {
	g.Protocol = w.Name()
	g.Address = w.Endpoint

	return g
}

func (w *Wireguard) BuildOutboundDetourConfig(allowInsecure bool) (*conf.OutboundDetourConfig, error) {
	out := &conf.OutboundDetourConfig{}
	out.Tag = "proxy"
	out.Protocol = w.Name()

	//c := conf.WireGuardConfig{
	//	IsClient:   true,
	//	KernelMode: nil,
	//	SecretKey:  w.SecretKey,
	//	Address:    strings.Split(w.LocalAddress, ","),
	//	Peers: []*conf.WireGuardPeerConfig{
	//		{
	//			PublicKey:    w.PublicKey,
	//			PreSharedKey: "",
	//			Endpoint:     w.Endpoint,
	//			KeepAlive:    0,
	//			AllowedIPs:   nil,
	//		},
	//	},
	//	MTU:            w.Mtu,
	//	DomainStrategy: "ForceIPv6v4",
	//}

	oset := json.RawMessage(fmt.Sprintf(`{
  "secretKey": "%s",
  "address": ["%s", "%s"],
  "peers": [
    {
      "endpoint": "%s",
      "publicKey": "%s"
    }
  ],
  "mtu": %d
}
`, w.SecretKey, strings.Split(w.LocalAddress, ",")[0], strings.Split(w.LocalAddress, ",")[1], w.Endpoint, w.PublicKey, w.Mtu,
	))
	out.Settings = &oset

	return out, nil
}

func (w *Wireguard) BuildInboundDetourConfig() (*conf.InboundDetourConfig, error) {
	return nil, nil
}
