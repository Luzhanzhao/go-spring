/*
 * Copyright 2012-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package SpringCore

import (
	"errors"
	"go/token"
	"go/types"
	"strings"

	"github.com/go-spring/go-spring-parent/spring-const"
	"github.com/spf13/cast"
)

// Condition 定义一个判断条件
type Condition interface {
	// Matches 成功返回 true，失败返回 false
	Matches(ctx SpringContext) bool
}

// ConditionFunc 定义 Condition 接口 Matches 方法的类型
type ConditionFunc func(ctx SpringContext) bool

// functionCondition 基于 Matches 方法的 Condition 实现
type functionCondition struct {
	fn ConditionFunc
}

// NewFunctionCondition functionCondition 的构造函数
func NewFunctionCondition(fn ConditionFunc) *functionCondition {
	return &functionCondition{fn}
}

// Matches 成功返回 true，失败返回 false
func (c *functionCondition) Matches(ctx SpringContext) bool {
	return c.fn(ctx)
}

// notCondition 对 Condition 取反的 Condition 实现
type notCondition struct {
	cond Condition
}

// NewNotCondition notCondition 的构造函数
func NewNotCondition(cond Condition) *notCondition {
	return &notCondition{cond}
}

// Matches 成功返回 true，失败返回 false
func (c *notCondition) Matches(ctx SpringContext) bool {
	return !c.cond.Matches(ctx)
}

// propertyCondition 基于属性值存在的 Condition 实现
type propertyCondition struct {
	name string
}

// NewPropertyCondition propertyCondition 的构造函数
func NewPropertyCondition(name string) *propertyCondition {
	return &propertyCondition{name}
}

// Matches 成功返回 true，失败返回 false
func (c *propertyCondition) Matches(ctx SpringContext) bool {
	return len(ctx.GetPrefixProperties(c.name)) > 0
}

// missingPropertyCondition 基于属性值不存在的 Condition 实现
type missingPropertyCondition struct {
	name string
}

// NewMissingPropertyCondition missingPropertyCondition 的构造函数
func NewMissingPropertyCondition(name string) *missingPropertyCondition {
	return &missingPropertyCondition{name}
}

// Matches 成功返回 true，失败返回 false
func (c *missingPropertyCondition) Matches(ctx SpringContext) bool {
	return len(ctx.GetPrefixProperties(c.name)) == 0
}

// propertyValueCondition 基于属性值匹配的 Condition 实现
type propertyValueCondition struct {
	name        string
	havingValue interface{}
}

// NewPropertyValueCondition propertyValueCondition 的构造函数
func NewPropertyValueCondition(name string, havingValue interface{}) *propertyValueCondition {
	return &propertyValueCondition{name, havingValue}
}

// Matches 成功返回 true，失败返回 false
func (c *propertyValueCondition) Matches(ctx SpringContext) bool {
	// 参考 /usr/local/go/src/go/types/eval_test.go 示例

	val, ok := ctx.GetDefaultProperty(c.name, "")
	if !ok { // 不存在直接返回 false
		return false
	}

	// 不是字符串则直接比较
	expectValue, ok := c.havingValue.(string)
	if !ok {
		return val == c.havingValue
	}

	// 字符串不是表达式的话则直接比较
	if ok = strings.Contains(expectValue, "$"); !ok {
		return val == expectValue
	}

	expr := strings.Replace(expectValue, "$", cast.ToString(val), -1)
	if ret, err := types.Eval(token.NewFileSet(), nil, token.NoPos, expr); err == nil {
		return ret.Value.String() == "true"
	} else {
		panic(err)
	}
}

// beanCondition 基于 Bean 存在的 Condition 实现
type beanCondition struct {
	selector interface{}
}

// NewBeanCondition beanCondition 的构造函数
func NewBeanCondition(selector interface{}) *beanCondition {
	return &beanCondition{selector}
}

// Matches 成功返回 true，失败返回 false
func (c *beanCondition) Matches(ctx SpringContext) bool {
	_, ok := ctx.FindBean(c.selector)
	return ok
}

// missingBeanCondition 基于 Bean 不能存在的 Condition 实现
type missingBeanCondition struct {
	selector interface{}
}

// NewMissingBeanCondition missingBeanCondition 的构造函数
func NewMissingBeanCondition(selector interface{}) *missingBeanCondition {
	return &missingBeanCondition{selector}
}

// Matches 成功返回 true，失败返回 false
func (c *missingBeanCondition) Matches(ctx SpringContext) bool {
	_, ok := ctx.FindBean(c.selector)
	return !ok
}

// expressionCondition 基于表达式的 Condition 实现
type expressionCondition struct {
	expression string
}

// NewExpressionCondition expressionCondition 的构造函数
func NewExpressionCondition(expression string) *expressionCondition {
	return &expressionCondition{expression}
}

// Matches 成功返回 true，失败返回 false
func (c *expressionCondition) Matches(ctx SpringContext) bool {
	panic(errors.New(SpringConst.UNIMPLEMENTED_METHOD))
}

// profileCondition 基于运行环境匹配的 Condition 实现
type profileCondition struct {
	profile string
}

// NewProfileCondition profileCondition 的构造函数
func NewProfileCondition(profile string) *profileCondition {
	return &profileCondition{profile}
}

// Matches 成功返回 true，失败返回 false
func (c *profileCondition) Matches(ctx SpringContext) bool {
	if c.profile != "" && c.profile != ctx.GetProfile() {
		return false
	}
	return true
}

// ConditionOp conditionNode 的计算方式
type ConditionOp int

const (
	ConditionOr   = ConditionOp(1) // 至少一个满足
	ConditionAnd  = ConditionOp(2) // 所有都要满足
	ConditionNone = ConditionOp(3) // 没有一个满足
)

// conditions 基于条件组的 Condition 实现
type conditions struct {
	op   ConditionOp
	cond []Condition
}

// NewConditions conditions 的构造函数
func NewConditions(op ConditionOp, cond ...Condition) *conditions {
	return &conditions{
		op:   op,
		cond: cond,
	}
}

// Matches 成功返回 true，失败返回 false
func (c *conditions) Matches(ctx SpringContext) bool {

	if len(c.cond) == 0 {
		panic(errors.New("no condition"))
	}

	switch c.op {
	case ConditionOr:
		for _, c0 := range c.cond {
			if c0.Matches(ctx) {
				return true
			}
		}
		return false
	case ConditionAnd:
		for _, c0 := range c.cond {
			if ok := c0.Matches(ctx); !ok {
				return false
			}
		}
		return true
	case ConditionNone:
		for _, c0 := range c.cond {
			if c0.Matches(ctx) {
				return false
			}
		}
		return true
	}

	panic(errors.New("error condition op mode"))
}

// conditionNode Condition 计算式节点，返回值是 'cond op next'
type conditionNode struct {
	cond Condition      // 条件
	op   ConditionOp    // 计算方式
	next *conditionNode // 下一个计算节点
}

// newConditionNode conditionNode 的构造函数
func newConditionNode() *conditionNode {
	return &conditionNode{}
}

// Matches 成功返回 true，失败返回 false
func (c *conditionNode) Matches(ctx SpringContext) bool {

	if c.cond == nil { // 空节点返回 true
		return true
	}

	if c.next != nil && c.next.cond == nil {
		panic(errors.New("last op need a cond triggered"))
	}

	if r := c.cond.Matches(ctx); c.next != nil {

		switch c.op {
		case ConditionOr: // or
			if r {
				return r
			} else {
				return c.next.Matches(ctx)
			}
		case ConditionAnd: // and
			if r {
				return c.next.Matches(ctx)
			} else {
				return false
			}
		default:
			panic(errors.New("error condition op mode"))
		}

	} else {
		return r
	}
}

// Conditional Condition 计算式
type Conditional struct {
	head *conditionNode
	curr *conditionNode
}

// NewConditional Conditional 的构造函数
func NewConditional() *Conditional {
	node := newConditionNode()
	return &Conditional{
		head: node,
		curr: node,
	}
}

// Empty 返回表达式是否为空
func (c *Conditional) Empty() bool {
	return c.head == c.curr
}

// Matches 成功返回 true，失败返回 false
func (c *Conditional) Matches(ctx SpringContext) bool {
	return c.head.Matches(ctx)
}

// Or c=a||b
func (c *Conditional) Or() *Conditional {
	node := newConditionNode()
	c.curr.op = ConditionOr
	c.curr.next = node
	c.curr = node
	return c
}

// And c=a&&b
func (c *Conditional) And() *Conditional {
	node := newConditionNode()
	c.curr.op = ConditionAnd
	c.curr.next = node
	c.curr = node
	return c
}

// OnCondition 设置一个 Condition
func (c *Conditional) OnCondition(cond Condition) *Conditional {
	if c.curr.cond != nil {
		c.And()
	}
	c.curr.cond = cond
	return c
}

// OnConditionNot 设置一个取反的 Condition
func (c *Conditional) OnConditionNot(cond Condition) *Conditional {
	return c.OnCondition(NewNotCondition(cond))
}

// OnProperty 设置一个 propertyCondition
func (c *Conditional) OnProperty(name string) *Conditional {
	return c.OnCondition(NewPropertyCondition(name))
}

// OnMissingProperty 设置一个 missingPropertyCondition
func (c *Conditional) OnMissingProperty(name string) *Conditional {
	return c.OnCondition(NewMissingPropertyCondition(name))
}

// OnPropertyValue 设置一个 propertyValueCondition
func (c *Conditional) OnPropertyValue(name string, havingValue interface{}) *Conditional {
	return c.OnCondition(NewPropertyValueCondition(name, havingValue))
}

// OnBean 设置一个 beanCondition
func (c *Conditional) OnBean(selector interface{}) *Conditional {
	return c.OnCondition(NewBeanCondition(selector))
}

// OnMissingBean 设置一个 missingBeanCondition
func (c *Conditional) OnMissingBean(selector interface{}) *Conditional {
	return c.OnCondition(NewMissingBeanCondition(selector))
}

// OnExpression 设置一个 expressionCondition
func (c *Conditional) OnExpression(expression string) *Conditional {
	return c.OnCondition(NewExpressionCondition(expression))
}

// OnMatches 设置一个 functionCondition
func (c *Conditional) OnMatches(fn ConditionFunc) *Conditional {
	return c.OnCondition(NewFunctionCondition(fn))
}

// OnProfile 设置一个 profileCondition
func (c *Conditional) OnProfile(profile string) *Conditional {
	return c.OnCondition(NewProfileCondition(profile))
}
