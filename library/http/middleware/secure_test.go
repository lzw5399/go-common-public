package middleware

import (
	"net"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestXForwardedForCheck(t *testing.T) {
	ips := []string{
		"127.0.0.0/8",
		"255.0.0.0/8",
		"123.1.0.0/16",
		// ipv6
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334/64",
		"1001:0DB8:85A3::/33",
	}

	ipNets := make([]*net.IPNet, 0, len(ips))
	for _, ip := range ips {
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			panic(err)
		}
		ipNets = append(ipNets, ipNet)
	}

	t.Run("empty allow", func(t *testing.T) {
		// arrange
		text := ""

		// act
		result := xForwardedForCheck(text, ipNets)

		// assert
		assert.Equal(t, result, true)
	})

	t.Run("contains illegal characters", func(t *testing.T) {
		// arrange
		text := "abcde12345@"

		// act
		result := xForwardedForCheck(text, ipNets)

		// assert
		assert.Equal(t, result, false)
	})

	t.Run("ipv4 allow", func(t *testing.T) {
		// arrange
		text := "127.1.0.1"

		// act
		result := xForwardedForCheck(text, ipNets)

		// assert
		assert.Equal(t, result, true)
	})

	t.Run("ipv4 single not allow", func(t *testing.T) {
		// arrange
		text := "123.123.0.12"

		// act
		result := xForwardedForCheck(text, ipNets)

		// assert
		assert.Equal(t, result, false)
	})

	t.Run("ipv4 multi not allow", func(t *testing.T) {
		// arrange
		text := "127.0.0.1,123.123.0.12"

		// act
		result := xForwardedForCheck(text, ipNets)

		// assert
		assert.Equal(t, result, false)
	})

	t.Run("ipv4 multi allow", func(t *testing.T) {
		// arrange
		text := "127.0.0.1,123.1.0.12"

		// act
		result := xForwardedForCheck(text, ipNets)

		// assert
		assert.Equal(t, result, true)
	})

	// ipv6 allow
	t.Run("ipv6 single allow", func(t *testing.T) {
		// arrange
		text := "2001:0db8:85a3:0000:0000:8a2e:0370:1234"

		// act
		result := xForwardedForCheck(text, ipNets)

		// assert
		assert.Equal(t, result, true)
	})

	// ipv6 single not allow
	t.Run("ipv6 single not allow", func(t *testing.T) {
		// arrange
		text := "2000:0db8:85a3:0000:0000:8a2e:0370:7335"

		// act
		result := xForwardedForCheck(text, ipNets)

		// assert
		assert.Equal(t, result, false)
	})

	// ipv6 multi allow
	t.Run("ipv6 multi allow", func(t *testing.T) {
		// arrange
		text := "2001:0db8:85a3:0000:0000:8a2e:0370:1234,2001:0db8:85a3:0000:0000:8a2e:0370:1237,2001:0db8:85a3:0000:0000:8a2e:0370:1231"

		// act
		result := xForwardedForCheck(text, ipNets)

		// assert
		assert.Equal(t, result, true)
	})

	// ipv6 multi not allow
	t.Run("ipv6 multi not allow", func(t *testing.T) {
		// arrange
		text := "2001:0db8:85a3:0000:0000:8a2e:0370:1234,2011:0db8:85a3:0000:0000:8a2e:0370:7335"

		// act
		result := xForwardedForCheck(text, ipNets)

		// assert
		assert.Equal(t, result, false)
	})

	// ipv4 ipv6 mixed allow
	t.Run("ipv4 ipv6 mixed allow", func(t *testing.T) {
		// arrange
		text := "2001:0db8:85a3:0000:0000:8a2e:0370:1234,123.1.255.255,1001:0DB8:85A3:0000:0000:8A2E:0370:7334"

		// act
		result := xForwardedForCheck(text, ipNets)

		// assert
		assert.Equal(t, result, true)
	})
}
