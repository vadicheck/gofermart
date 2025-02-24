package luhn

const l2 = 2
const l9 = 9
const l10 = 10

func CalculateLuhn(number int) int {
	checkNumber := checksum(number)

	if checkNumber == 0 {
		return 0
	}

	return l10 - checkNumber
}

func Valid(number int) bool {
	return (number%l10+checksum(number/l10))%l10 == 0
}

func checksum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % l10

		if i%l2 == 0 { // even
			cur *= l2
			if cur > l9 {
				cur = cur%l10 + cur/l10
			}
		}

		luhn += cur
		number /= l10
	}

	return luhn % l10
}
