package id

import (
	"errors"
	"net"
	"sync"
	"time"
)

// These constants are the bit lengths of Sonyflake ID parts.
const (
	BitLenTime      = 39                               // bit length of time
	BitLenSequence  = 8                                // bit length of sequence number
	BitLenMachineID = 63 - BitLenTime - BitLenSequence // bit length of machine id
)

// Snowflake is a distributed unique ID generator.
type Snowflake struct {
	opts options

	mutex       sync.Mutex
	machineId   uint16
	startTime   int64
	elapsedTime int64
	sequence    uint16
}

type MachineId func() (uint16, error)

type options struct {
	startTime time.Time
	machineId MachineId
}

const sequenceRefreshPeriod = 1e7 // 10 mesc

var defaultOptions = options{
	startTime: time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC),
	machineId: CIDRMachineId,
}

type Option interface {
	apply(*options)
}

type funcOption struct {
	f func(*options)
}

func (fo *funcOption) apply(opts *options) {
	fo.f(opts)
}

func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

func NewSnowflake(opt ...Option) (*Snowflake, error) {
	opts := defaultOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	snowflake := &Snowflake{}

	if opts.startTime.IsZero() || opts.startTime.After(time.Now()) {
		return nil, errors.New("Invalide start time.")
	} else {
		snowflake.startTime = opts.startTime.UnixNano() / sequenceRefreshPeriod
	}

	var err error
	if opts.machineId == nil {
		snowflake.machineId, err = CIDRMachineId()
	} else {
		snowflake.machineId, err = opts.machineId()
	}

	if err != nil {
		return nil, err
	}

	return snowflake, nil
}

func (snowflake *Snowflake) NextId() (uint64, error) {
	const maskSequence = uint16(1<<BitLenSequence - 1)

	snowflake.mutex.Lock()
	defer snowflake.mutex.Unlock()

	curElapsedTime := currentElapsedTime(snowflake.startTime)
	if snowflake.elapsedTime < curElapsedTime {
		snowflake.elapsedTime = curElapsedTime
		snowflake.sequence = 0
	} else {
		snowflake.sequence = (snowflake.sequence + 1) & maskSequence
		if snowflake.sequence == 0 {
			snowflake.elapsedTime++
			overtime := snowflake.elapsedTime - curElapsedTime
			time.Sleep(sleepTime(overtime))
		}
	}

	return snowflake.toId()
}

func (snowflake *Snowflake) toId() (uint64, error) {
	if snowflake.elapsedTime >= 1<<BitLenTime {
		return 0, errors.New("Over the time limit.")
	}

	return uint64(snowflake.elapsedTime)<<(BitLenSequence+BitLenMachineID) |
		uint64(snowflake.sequence)<<BitLenMachineID |
		uint64(snowflake.machineId), nil
}

func toSequenceRefreshTime(t time.Time) int64 {
	return t.UnixNano() / sequenceRefreshPeriod
}

func currentElapsedTime(startTime int64) int64 {
	return toSequenceRefreshTime(time.Now()) - startTime
}

func sleepTime(overtime int64) time.Duration {
	return time.Duration(overtime)*10*time.Millisecond -
		time.Duration(time.Now().UnixNano()%sequenceRefreshPeriod)*time.Nanosecond
}

func MachineIdGenerator(generator MachineId) Option {
	return newFuncOption(func(opts *options) {
		opts.machineId = generator
	})
}

func StartTime(t time.Time) Option {
	return newFuncOption(func(opts *options) {
		opts.startTime = t
	})
}

// CIDRMachineID base CIDR network prefix length.
// CIDR prefix length must greater than 16.
func CIDRMachineId() (uint16, error) {
	return lower16BitPrivateIP()
}

func lower16BitPrivateIP() (uint16, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}

func privateIPv4() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}

	return nil, errors.New("No private ip address.")
}

// Private ipv4:
// 10.0.0.0~10.255.255.255, 10.0.0.0/8
// 172.16.0.0~172.31.255.255, 172.16.0.0/12
// 192.168.0.0~192.168.255.255, 192.168.0.0/16
func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 ||
			ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) ||
			ip[0] == 192 && ip[1] == 168)
}

func Decompose(id uint64) map[string]uint64 {
	const maskSequence = uint64((1<<BitLenSequence - 1) << BitLenMachineID)
	const maskMachineID = uint64(1<<BitLenMachineID - 1)

	msb := id >> 63
	time := id >> (BitLenSequence + BitLenMachineID)
	sequence := id & maskSequence >> BitLenMachineID
	machineID := id & maskMachineID
	ip2 := machineID >> 8
	ip3 := machineID & 0xff
	return map[string]uint64{
		"id":         id,
		"msb":        msb,
		"time":       time,
		"sequence":   sequence,
		"machine-id": machineID,
		"ip2":        ip2,
		"ip3":        ip3,
	}
}
