package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetNextMonthStartMs(t *testing.T) {
	monthStartMs := GetMonthStartMs(time.Now().UnixMilli())
	fmt.Println(monthStartMs)

	monthEndMs := GetMonthEndMs(time.Now().UnixMilli())
	fmt.Println(monthEndMs)

	nextMonthStartMs := GetNextMonthStartMs(time.Now().UnixMilli())
	fmt.Println(nextMonthStartMs)

	nextMonthEndMs := GetMonthEndMs(time.Now().UnixMilli())
	fmt.Println(nextMonthEndMs)

	currentDayEndMs := GetCurDayEndMs(time.Now().UnixMilli())
	fmt.Println(currentDayEndMs)

	ms := GetCurDayEndMs(1722441600000) // 2024-08-01 00:00:00
	assert.True(t, ms == 1722527999000) // 2024-08-01 23:59:59

	ms = GetCurDayEndMs(1722527999000)  // 2024-08-01 23:59:59
	assert.True(t, ms == 1722527999000) // 2024-08-01 23:59:59

	ms = GetCurDayEndMs(1723222800000)  // 2024-08-10 01:00:00
	assert.True(t, ms == 1723305599000) // 2024-08-10 23:59:59

	ms = GetCurDayEndMs(1730394000000)  // 2024-11-01 01:00:00
	assert.True(t, ms == 1730476799000) // 2024-11-01 23:59:59

	ms = GetCurDayEndMs(1732903200000)  // 2024-11-30 02:00:00
	assert.True(t, ms == 1732989599000) // 2024-11-30 23:59:59
}
