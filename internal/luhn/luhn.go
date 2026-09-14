// Package luhn реализует проверку номеров по алгоритму Луна.
package luhn

// Valid сообщает, состоит ли number только из цифр и проходит ли проверку Луна.
func Valid(number string) bool {
	if number == "" {
		return false
	}

	sum := 0
	alt := false
	for i := len(number) - 1; i >= 0; i-- {
		d := number[i] - '0'
		if d > 9 {
			return false
		}
		n := int(d)
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}
	return sum%10 == 0
}
