// nsdp.go
//
// Copyright (C) 2022 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
//

package nsdp

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	nsdplib "github.com/hdecarne-github/go-nsdp"
)

type NSDP struct {
	Target      string `toml:"target"`
	DeviceLimit int    `toml:"device_limit"`
	Timeout     int    `toml:"timeout"`
	Debug       bool   `toml:"debug"`

	Log telegraf.Logger `toml:"-"`
}

func NewNSDP() *NSDP {
	return &NSDP{
		Target:  nsdplib.IPv4BroadcastTarget,
		Timeout: 2,
	}
}

func (plugin *NSDP) SampleConfig() string {
	return `
  ## The target address to use for NSDP processing
  # target = "255.255.255.255:63322"
  ## The device limit to use (0 means no limit)
  # device_limit = 0
  ## The receive timeout to use (in seconds)
  # timeout = 2
  ## Enable debug output
  # debug = false
`
}

func (plugin *NSDP) Description() string {
	return "Gather network stats from NSDP capable devices"
}

func (plugin *NSDP) Gather(a telegraf.Accumulator) error {
	conn, err := nsdplib.NewConn(plugin.Target, plugin.Debug)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.ReceiveDeviceLimit = uint(plugin.DeviceLimit)
	conn.ReceiveTimeout = time.Duration(plugin.Timeout) * time.Second
	request := nsdplib.NewMessage(nsdplib.ReadRequest)
	request.AppendTLV(nsdplib.EmptyDeviceModel())
	request.AppendTLV(nsdplib.EmptyDeviceName())
	request.AppendTLV(nsdplib.EmptyDeviceIP())
	request.AppendTLV(nsdplib.EmptyPortStatistic())
	responses, err := conn.SendReceiveMessage(request)
	if err != nil {
		return err
	}
	for device, response := range responses {
		if plugin.Debug {
			plugin.Log.Infof("Processing device: %s", device)
		}
		plugin.processResponse(a, response)
	}
	return nil
}

func (plugin *NSDP) processResponse(a telegraf.Accumulator, response *nsdplib.Message) {
	var deviceModel string
	var deviceName string
	var deviceIP net.IP
	portStatistics := make(map[uint8]*nsdplib.PortStatistic, 0)
	for _, tlv := range response.Body {
		switch tlv.Type() {
		case nsdplib.TypeDeviceModel:
			deviceModel = tlv.(*nsdplib.DeviceModel).Model
		case nsdplib.TypeDeviceName:
			deviceName = tlv.(*nsdplib.DeviceName).Name
		case nsdplib.TypeDeviceIP:
			deviceIP = tlv.(*nsdplib.DeviceIP).IP
		case nsdplib.TypePortStatistic:
			portStatistic := tlv.(*nsdplib.PortStatistic)
			portStatistics[portStatistic.Port] = portStatistic
		}
	}
	for port, statistic := range portStatistics {
		if statistic.Received != 0 || statistic.Sent != 0 {
			tags := make(map[string]string)
			tags["nsdp_device_model"] = deviceModel
			tags["nsdp_device_name"] = deviceName
			tags["nsdp_device_ip"] = deviceIP.String()
			tags["nsdp_device_port"] = strconv.FormatUint(uint64(port), 10)
			tags["nsdp_device_port_id"] = fmt.Sprintf("%s:%d", deviceName, port)
			fields := make(map[string]interface{})
			fields["received"] = statistic.Received
			fields["sent"] = statistic.Sent
			fields["packets"] = statistic.Packets
			fields["broadcasts"] = statistic.Broadcasts
			fields["multicasts"] = statistic.Multicasts
			fields["errors"] = statistic.Errors
			a.AddCounter("nsdp_port_statistic", fields, tags)
		}
	}
}

func init() {
	inputs.Add("nsdp", func() telegraf.Input {
		return NewNSDP()
	})
}
