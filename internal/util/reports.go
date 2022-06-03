package util

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	"go.uber.org/zap"
)

var (
	reportsHost  = "kong-hf.konghq.com"
	reportsPort  = 61833
	pingInterval = 3600
	tlsConf      = tls.Config{MinVersion: tls.VersionTLS12}
	tcpTimeout   = 30 * time.Second
	dialer       = net.Dialer{Timeout: tcpTimeout}
)

const (
	product = "koko"
)

// Info holds the metadata to be sent as part of a report.
type Info struct {
	ID          string
	KokoVersion string
}

// Reporter sends anonymous reports of runtime properties and
// errors in Kong.
type Reporter struct {
	Info           Info
	serializedInfo string
	Logger         *zap.Logger
}

func (r *Reporter) once() {
	var serializedInfo string
	serializedInfo += fmt.Sprintf("v=%s;", r.Info.KokoVersion)
	serializedInfo += fmt.Sprintf("id=%s;", r.Info.ID)

	hostInfo, err := host.Info()
	if err != nil {
		r.Logger.Sugar().Debugf("failed to get host information: %v", err)
	} else {
		serializedInfo += fmt.Sprintf("hn=%s;osv=%s %s;",
			hostInfo.Hostname, hostInfo.Platform, hostInfo.PlatformVersion)
	}

	r.serializedInfo = serializedInfo
}

// Run starts the reporter. It will send reports until done is closed.
func (r Reporter) Run(ctx context.Context) {
	r.once()

	r.sendStart()
	ticker := time.NewTicker(time.Duration(pingInterval) * time.Second)
	i := 1
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.sendPing(i * pingInterval)
			i++
		}
	}
}

func (r *Reporter) sendStart() {
	signal := product + "-start"
	r.send(signal, 0)
}

func (r *Reporter) sendPing(uptime int) {
	signal := product + "-ping"
	r.send(signal, uptime)
}

func (r *Reporter) send(signal string, uptime int) {
	const base10 = 10
	message := "<14>signal=" + signal + ";uptime=" +
		strconv.Itoa(uptime) + ";" + r.serializedInfo
	conn, err := tls.DialWithDialer(&dialer, "tcp", net.JoinHostPort(reportsHost,
		strconv.FormatUint(uint64(reportsPort), base10)), &tlsConf)
	if err != nil {
		r.Logger.Sugar().Debugf("failed to connect to reporting server: %s",
			err)
		return
	}
	err = conn.SetDeadline(time.Now().Add(time.Minute))
	if err != nil {
		r.Logger.Sugar().Debugf("failed to set report connection deadline: %s", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	_, err = conn.Write([]byte(message))
	if err != nil {
		r.Logger.Sugar().Debugf("failed to send report: %s", err)
	}
}
