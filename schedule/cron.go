// Package schedule 定时任务调度框架
//
// 提供 6 字段 Spring 风格 Cron 表达式解析（CronExpression）、
// 任务调度器（Scheduler）、任务接口（Task）以及 @Scheduled 注解扫描。
package schedule

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var monthNames = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

var dayNames = map[string]int{
	"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6,
}

// CronExpression 6字段 Spring 风格 Cron 表达式（秒 分 时 日 月 周）
type CronExpression struct {
	second        uint64
	minute        uint64
	hour          uint64
	dayOfMonth    uint64
	month         uint64
	dayOfWeek     uint64
	hasDayOfMonth bool
	hasDayOfWeek  bool
}

// Parse 将 6字段 Cron 表达式解析为 CronExpression
//
// 格式: 秒 分 时 日 月 周
//   - 秒: 0-59
//   - 分: 0-59
//   - 时: 0-23
//   - 日: 1-31，支持 ? 表示不指定
//   - 月: 1-12 或 JAN-DEC
//   - 周: 0-6 或 SUN-SAT，支持 ? 表示不指定
//
// 每个字段支持: 固定值, 列表(,), 范围(-), 步进(/), 通配符(*)
func Parse(expr string) (*CronExpression, error) {
	fields := strings.Fields(expr)
	if len(fields) != 6 {
		return nil, fmt.Errorf("schedule: cron expression must have 6 fields, got %d", len(fields))
	}

	ce := &CronExpression{}
	var err error

	ce.second, err = parseField(fields[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("schedule: invalid seconds field %q: %w", fields[0], err)
	}

	ce.minute, err = parseField(fields[1], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("schedule: invalid minutes field %q: %w", fields[1], err)
	}

	ce.hour, err = parseField(fields[2], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("schedule: invalid hours field %q: %w", fields[2], err)
	}

	if fields[3] == "?" {
		ce.hasDayOfMonth = false
	} else {
		ce.hasDayOfMonth = true
		ce.dayOfMonth, err = parseField(fields[3], 1, 31)
		if err != nil {
			return nil, fmt.Errorf("schedule: invalid day-of-month field %q: %w", fields[3], err)
		}
	}

	ce.month, err = parseFieldWithNames(fields[4], 1, 12, monthNames)
	if err != nil {
		return nil, fmt.Errorf("schedule: invalid month field %q: %w", fields[4], err)
	}

	if fields[5] == "?" {
		ce.hasDayOfWeek = false
	} else {
		ce.hasDayOfWeek = true
		ce.dayOfWeek, err = parseFieldWithNames(fields[5], 0, 6, dayNames)
		if err != nil {
			return nil, fmt.Errorf("schedule: invalid day-of-week field %q: %w", fields[5], err)
		}
	}

	if !ce.hasDayOfMonth {
		ce.dayOfMonth = allBits(1, 31)
	}
	if !ce.hasDayOfWeek {
		ce.dayOfWeek = allBits(0, 6)
	}

	return ce, nil
}

// Next 计算从 after 时刻起下一个满足 Cron 表达式的时间点
//
// 从 after+1秒 开始向前搜索，最大搜索范围为 4 年。
// 如果 4 年内无法找到匹配时间，返回零值 time.Time{}。
func (ce *CronExpression) Next(after time.Time) time.Time {
	t := after.UTC().Add(time.Second).Truncate(time.Second)
	end := t.AddDate(4, 0, 0)

	for t.Before(end) {
		t = ce.advanceToMonth(t)
		if !t.Before(end) {
			break
		}

		var dayAdvanced bool
		t, dayAdvanced = ce.advanceToDay(t)
		if !t.Before(end) {
			break
		}
		if dayAdvanced {
			continue
		}

		h := t.Hour()
		if !isBitSet(ce.hour, h) {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC).Add(time.Hour)
			continue
		}

		m := t.Minute()
		if !isBitSet(ce.minute, m) {
			t = t.Add(time.Minute)
			continue
		}

		s := t.Second()
		if !isBitSet(ce.second, s) {
			t = t.Add(time.Second)
			continue
		}

		return t
	}

	return time.Time{}
}

// advanceToMonth 将 t 推进到满足月份条件的月份第一天
func (ce *CronExpression) advanceToMonth(t time.Time) time.Time {
	m := int(t.Month())
	if isBitSet(ce.month, m) {
		return t
	}
	nextMonth := t.Month() + 1
	y := t.Year()
	if nextMonth > 12 {
		nextMonth = 1
		y++
	}
	return time.Date(y, nextMonth, 1, 0, 0, 0, 0, time.UTC)
}

// advanceToDay 检查日/周条件，匹配时返回 (t, false)，不匹配时推进到次日返回 (nextDay, true)
func (ce *CronExpression) advanceToDay(t time.Time) (time.Time, bool) {
	dom := t.Day()
	dow := int(t.Weekday())
	domMatch := isBitSet(ce.dayOfMonth, dom)
	dowMatch := isBitSet(ce.dayOfWeek, dow)

	matched := false
	if ce.hasDayOfMonth && ce.hasDayOfWeek {
		matched = domMatch || dowMatch
	} else if ce.hasDayOfMonth {
		matched = domMatch
	} else {
		matched = dowMatch
	}

	if matched {
		return t, false
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1), true
}

// allBits 生成 [min, max] 范围内所有位均为 1 的位图
func allBits(min, max int) uint64 {
	var bits uint64
	for i := min; i <= max; i++ {
		setBit(&bits, i)
	}
	return bits
}

// setBit 将位图中 pos 位置为 1
func setBit(bits *uint64, pos int) {
	*bits |= 1 << uint(pos)
}

// isBitSet 检查位图中 pos 位置是否为 1
func isBitSet(bits uint64, pos int) bool {
	if pos < 0 || pos > 63 {
		return false
	}
	return bits&(1<<uint(pos)) != 0
}

// parseField 解析单个 Cron 字段，返回对应的位图
//
// 支持格式:
//   - 通配符: *
//   - 固定值: 5
//   - 列表: 1,3,5
//   - 范围: 1-5
//   - 步进: */5, 1-10/2, 5/2
func parseField(field string, min, max int) (uint64, error) {
	if field == "*" || field == "?" {
		return allBits(min, max), nil
	}

	var bits uint64

	parts := strings.SplitSeq(field, ",")
	for part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var err error
		if strings.Contains(part, "/") {
			bits, err = parseStep(bits, part, min, max)
			if err != nil {
				return 0, err
			}
		} else if strings.Contains(part, "-") {
			vals, err := parseRange(part, min, max)
			if err != nil {
				return 0, err
			}
			bits |= vals
		} else {
			val, err := strconv.Atoi(part)
			if err != nil {
				return 0, fmt.Errorf("invalid value %q", part)
			}
			if val < min || val > max {
				return 0, fmt.Errorf("value %d out of range [%d,%d]", val, min, max)
			}
			setBit(&bits, val)
		}
	}

	return bits, nil
}

// parseFieldWithNames 解析支持名称映射的字段（月份、星期）
func parseFieldWithNames(field string, min, max int, names map[string]int) (uint64, error) {
	resolved := resolveNames(field, names)
	return parseField(resolved, min, max)
}

// resolveNames 将字段中的名称替换为数字值（如 JAN→1, MON→1）
func resolveNames(field string, names map[string]int) string {
	lower := strings.ToLower(field)
	for name, val := range names {
		lower = strings.ReplaceAll(lower, strings.ToLower(name), strconv.Itoa(val))
	}
	return lower
}

// parseStep 解析步进表达式（如 */5、1-10/2、5/2）
func parseStep(bits uint64, part string, min, max int) (uint64, error) {
	subparts := strings.SplitN(part, "/", 2)
	if len(subparts) != 2 {
		return 0, fmt.Errorf("invalid step expression %q", part)
	}

	step, err := strconv.Atoi(subparts[1])
	if err != nil || step <= 0 {
		return 0, fmt.Errorf("invalid step value %q", subparts[1])
	}

	var rangeMin, rangeMax int
	if subparts[0] == "*" {
		rangeMin = min
		rangeMax = max
	} else if strings.Contains(subparts[0], "-") {
		rangeParts := strings.SplitN(subparts[0], "-", 2)
		rangeMin, err = strconv.Atoi(rangeParts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid range start %q", rangeParts[0])
		}
		rangeMax, err = strconv.Atoi(rangeParts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid range end %q", rangeParts[1])
		}
	} else {
		rangeMin, err = strconv.Atoi(subparts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid step start %q", subparts[0])
		}
		rangeMax = max
	}

	if rangeMin > rangeMax {
		return 0, fmt.Errorf("range start %d > end %d", rangeMin, rangeMax)
	}
	if rangeMin < min || rangeMax > max {
		return 0, fmt.Errorf("step range [%d,%d] out of bounds [%d,%d]", rangeMin, rangeMax, min, max)
	}

	// 步进值不能超过范围大小，防止整数溢出
	if step > rangeMax-rangeMin {
		setBit(&bits, rangeMin)
	} else {
		for i := rangeMin; i <= rangeMax; i += step {
			setBit(&bits, i)
		}
	}
	return bits, nil
}

// parseRange 解析范围表达式（如 1-5），返回范围中所有值的位图
func parseRange(part string, min, max int) (uint64, error) {
	rangeParts := strings.SplitN(part, "-", 2)
	rangeMin, err := strconv.Atoi(rangeParts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid range start %q", rangeParts[0])
	}
	rangeMax, err := strconv.Atoi(rangeParts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid range end %q", rangeParts[1])
	}

	if rangeMin > rangeMax {
		return 0, fmt.Errorf("range start %d > end %d", rangeMin, rangeMax)
	}
	if rangeMin < min || rangeMax > max {
		return 0, fmt.Errorf("range [%d,%d] out of bounds [%d,%d]", rangeMin, rangeMax, min, max)
	}

	var bits uint64
	for i := rangeMin; i <= rangeMax; i++ {
		setBit(&bits, i)
	}
	return bits, nil
}
