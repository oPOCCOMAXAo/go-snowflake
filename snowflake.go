package snowflake

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	DefaultMachineBits  = 10
	DefaultSequenceBits = 12
	MaxBits             = 63
)

// Default Snowflake ID layout:
//
//  0 -  0 bits: unused
//  1 - 41 bits: milliseconds since epoch
// 42 - 51 bits: machine ID
// 52 - 63 bits: sequence number
//
// The time is in milliseconds since a custom epoch.
//
// The machine ID is a unique identifier for the machine generating the ID.
//
// The sequence number is used to generate unique IDs in the same millisecond.

// Configurable Snowflake ID generator.
//
// It is safe for concurrent use.
//
// Zero value is not a valid generator.
//
// See more: https://en.wikipedia.org/wiki/Snowflake_ID
type Generator struct {
	mu               sync.Mutex
	time             uint64 // current time.
	seq              uint64 // current sequence number.
	machineIDShifted uint64 // machine ID shifted to the proper position.
	epochStartNano   uint64 // epoch start time in nanoseconds.
	machineBits      uint64 // number of bits for the machine ID.
	sequenceBits     uint64 // number of bits for the sequence number.
	timeBits         uint64 // number of bits for the time.
	machineShift     uint64 // number of bits to shift the machine ID.
	timeShift        uint64 // number of bits to shift the time.
	sequenceMask     uint64 // mask for the sequence number.
	timeMask         uint64 // mask for the time.
	nanoSecondsShift uint64 // number of bits to shift the nanoseconds.
}

type Config struct {
	// MachineID is the unique ID of the machine running the generator.
	// Required. Must be unique per generator instance.
	MachineID uint64

	// Epoch is the time in seconds since the Unix epoch.
	// Optional. Default is 1970-01-01 00:00:00 UTC.
	EpochStartUnixSeconds uint64

	// MachineBits is the number of bits to use for the machine ID.
	// Optional. Default is 10.
	// Bits sum must be 63.
	MachineBits uint64

	// SequenceBits is the number of bits to use for the sequence number.
	// Optional. Default is 12.
	// Bits sum must be 63.
	SequenceBits uint64

	// TimeBits is the number of bits to use for the time.
	// Optional. Default is padded to 63 - MachineBits - SequenceBits.
	// Bits sum must be 63.
	TimeBits uint64
}

func New(
	config Config,
) (*Generator, error) {
	res := &Generator{
		mu:             sync.Mutex{},
		epochStartNano: config.EpochStartUnixSeconds * 1e9, //nolint:mnd
		machineBits:    config.MachineBits,
		sequenceBits:   config.SequenceBits,
		timeBits:       config.TimeBits,
	}

	if res.machineBits == 0 {
		res.machineBits = DefaultMachineBits
	}

	if res.sequenceBits == 0 {
		res.sequenceBits = DefaultSequenceBits
	}

	if res.timeBits == 0 {
		res.timeBits = MaxBits - res.machineBits - res.sequenceBits
	}

	if res.sequenceBits+res.machineBits+res.timeBits != MaxBits {
		return nil, errors.New("invalid config: SequenceBits + MachineBits + TimeBits must equal 63")
	}

	res.nanoSecondsShift = 64 - res.timeBits
	res.machineShift = res.sequenceBits
	res.timeShift = res.sequenceBits + res.machineBits

	machineMax := uint64(^(int64(-1) << res.machineBits))
	if config.MachineID > machineMax {
		return nil, errors.Errorf("invalid machine id; must be 0 ≤ id < %d", machineMax)
	}

	res.machineIDShifted = config.MachineID << res.machineShift
	res.sequenceMask = uint64(^(int64(-1) << res.sequenceBits))
	res.timeMask = uint64(^(int64(-1) << res.timeBits))

	return res, nil
}

func (g *Generator) MachineID() uint64 {
	return g.machineIDShifted >> g.machineShift
}

func (g *Generator) generateTime() uint64 {
	return ((uint64(time.Now().UnixNano()) - g.epochStartNano) >> g.nanoSecondsShift) & g.timeMask
}

// Next generates a new ID.
//
// It is safe for concurrent use.
func (g *Generator) Next() uint64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	newTime := g.generateTime()

	switch {
	// if our time is in the future, use that with a zero sequence number.
	case newTime > g.time:
		g.time = newTime
		g.seq = 0

	// we now know that our time is at or before the current time.
	// if we're at the maximum sequence, bump to the next millisecond
	case g.seq == g.sequenceMask:
		g.time++
		g.seq = 0

	// otherwise, increment the sequence.
	default:
		g.seq++
	}

	return (g.time << g.timeShift) | g.machineIDShifted | g.seq
}
