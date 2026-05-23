package sdk

import "encoding/binary"

// assignBucket returns a stable bucket in [0, 9999] for a user+key pair.
// Must produce identical output to the backend's assignment logic.
//
// Uses an inlined MurmurHash3-32 (seed=0) to avoid the unsafe pointer
// arithmetic in github.com/spaolacci/murmur3, which triggers the Go runtime's
// checkptr sanitizer under -race on Go 1.21+.
// The algorithm is identical to the library's Sum32 — existing known-value
// tests pin the output and will catch any drift.
func assignBucket(userID, key string) uint32 {
	return murmur3_32([]byte(userID+":"+key)) % 10000
}

// murmur3_32 is MurmurHash3-32 with seed=0, inlined without unsafe.
// Reference: https://github.com/aappleby/smhasher/blob/master/src/MurmurHash3.cpp
func murmur3_32(data []byte) uint32 {
	const (
		c1 = uint32(0xcc9e2d51)
		c2 = uint32(0x1b873593)
	)

	h1 := uint32(0) // seed = 0
	nblocks := len(data) / 4

	for i := 0; i < nblocks; i++ {
		k1 := binary.LittleEndian.Uint32(data[i*4:])
		k1 *= c1
		k1 = (k1 << 15) | (k1 >> 17)
		k1 *= c2
		h1 ^= k1
		h1 = (h1 << 13) | (h1 >> 19)
		h1 = h1*4 + h1 + 0xe6546b64
	}

	tail := data[nblocks*4:]
	var k1 uint32
	switch len(tail) & 3 {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1
		k1 = (k1 << 15) | (k1 >> 17)
		k1 *= c2
		h1 ^= k1
	}

	h1 ^= uint32(len(data))
	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16
	return h1
}

func evaluateFlag(flag Flag, userID string) bool {
	if !flag.Enabled || userID == "" {
		return false
	}
	for _, rule := range flag.Rules {
		if rule.Type == "percentage" {
			// Value is 0–1; bucket is 0–9999. Scale: value*10000.
			bucket := assignBucket(userID, flag.Key)
			return bucket < uint32(rule.Value*10000)
		}
	}
	return true
}

// The backend only sends running experiments in the config snapshot,
// so there is no status check here.
func evaluateExperiment(exp Experiment, userID string) string {
	if userID == "" {
		return ""
	}

	bucket := assignBucket(userID, exp.Key)

	// Traffic gate — is the user allocated to this experiment?
	if bucket >= uint32(exp.TrafficPercent*100) {
		return ""
	}

	// Which variant? Use bucket mod totalWeight for stable assignment.
	totalWeight := 0
	for _, v := range exp.Variants {
		totalWeight += v.Weight
	}
	if totalWeight == 0 {
		return ""
	}

	variantBucket := int(bucket) % totalWeight
	cumulative := 0
	for _, v := range exp.Variants {
		cumulative += v.Weight
		if variantBucket < cumulative {
			return v.Key
		}
	}
	return ""
}
