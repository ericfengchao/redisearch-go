package redisearch

import "fmt"

const (
	negInf = "-inf"
	posInf = "+inf"
)

type Predicate struct {
	Property                   string
	min, max                   interface{}
	minInclusive, maxInclusive bool
}

func (p Predicate) serialize() string {
	min := fmt.Sprintf("%v", p.min)
	if !p.minInclusive {
		min = fmt.Sprintf("(%v", min)
	}
	max := fmt.Sprintf("%v", p.max)
	if !p.maxInclusive {
		max = fmt.Sprintf("(%v", max)
	}
	return fmt.Sprintf("@%s:[%s %s]", p.Property, min, max)
}

func InRange(property string, min, max interface{}, minInclusive, maxInclusive bool) Predicate {
	return Predicate{
		Property:     property,
		min:          min,
		max:          max,
		minInclusive: minInclusive,
		maxInclusive: maxInclusive,
	}
}

func Equals(property string, value interface{}) Predicate {
	return InRange(property, value, value, true, true)

}

func LessThan(property string, value interface{}) Predicate {
	return InRange(property, negInf, value, true, false)
}

func LessThanEquals(property string, value interface{}) Predicate {
	return InRange(property, negInf, value, true, true)
}

func GreaterThan(property string, value interface{}) Predicate {
	return InRange(property, value, posInf, false, true)
}

func GreaterThanEquals(property string, value interface{}) Predicate {
	return InRange(property, value, posInf, true, true)

}
