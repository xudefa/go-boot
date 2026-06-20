package aop

import (
	"cmp"
	"slices"
)

// AspectMeta 切面元数据
//
// 用于存储切面的实例、切点、通知和执行顺序.
// 字段说明:
//   - Instance: 切面实例
//   - PointCut: 切点定义,匹配目标方法
//   - Advice: 通知,包含增强逻辑
//   - Order: 执行顺序,数字越小越先执行
type AspectMeta struct {
	Instance any
	PointCut PointCut
	Advice   Advice
	Order    int
}

// SortAspectsByOrder 按Order升序排序切面列表
//
// 使用标准库slices.SortFunc按Order升序排列切面，Order小的在前。
func SortAspectsByOrder(aspects []*AspectMeta) {
	if len(aspects) < 2 {
		return
	}
	slices.SortFunc(aspects, func(a, b *AspectMeta) int {
		if a == nil && b == nil {
			return 0
		}
		if a == nil {
			return 1
		}
		if b == nil {
			return -1
		}
		return cmp.Compare(a.Order, b.Order)
	})
}
