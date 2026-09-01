package validation

import (
	"database/sql"
	"errors"
	"math/big"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
var decimalPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]{1,18})?$`)
var evmAddressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
var evmTxHashPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`)

func ParseDecimal(value string) (*big.Rat, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || !decimalPattern.MatchString(trimmed) {
		return nil, false
	}

	integerPart := strings.SplitN(trimmed, ".", 2)[0]
	if significantIntegerDigits(integerPart) > 18 {
		return nil, false
	}

	rat, ok := new(big.Rat).SetString(trimmed)
	return rat, ok
}

func DecimalString(decimal *big.Rat) (string, bool) {
	numerator := new(big.Int).Set(decimal.Num())
	denominator := new(big.Int).Set(decimal.Denom())
	integer := new(big.Int)
	remainder := new(big.Int)
	integer.QuoRem(numerator, denominator, remainder)

	if remainder.Sign() == 0 {
		if significantIntegerDigits(integer.String()) > 18 {
			return "", false
		}

		return integer.String(), true
	}

	digits := strings.Builder{}
	ten := big.NewInt(10)
	for i := 0; i < 18 && remainder.Sign() != 0; i++ {
		remainder.Mul(remainder, ten)
		digit := new(big.Int)
		digit.QuoRem(remainder, denominator, remainder)
		digits.WriteString(digit.String())
	}
	if remainder.Sign() != 0 {
		return "", false
	}

	if significantIntegerDigits(integer.String()) > 18 {
		return "", false
	}

	fractional := strings.TrimRight(digits.String(), "0")
	if fractional == "" {
		return integer.String(), true
	}

	return integer.String() + "." + fractional, true
}

func significantIntegerDigits(value string) int {
	trimmed := strings.TrimLeft(value, "0")
	if trimmed == "" {
		return 0
	}

	return len(trimmed)
}

func IsUUIDShape(value string) bool {
	return uuidPattern.MatchString(value)
}

func IsEVMAddressShape(value string) bool {
	return evmAddressPattern.MatchString(value)
}

func IsEVMTxHashShape(value string) bool {
	return evmTxHashPattern.MatchString(value)
}

func OptionalString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}

	trimmedValue := strings.TrimSpace(*value)
	if trimmedValue == "" {
		return sql.NullString{}
	}

	return sql.NullString{String: trimmedValue, Valid: true}
}

func DefaultString(value *string, fallback string) string {
	if value == nil {
		return fallback
	}

	trimmedValue := strings.TrimSpace(*value)
	if trimmedValue == "" {
		return fallback
	}

	return trimmedValue
}

func IsForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
