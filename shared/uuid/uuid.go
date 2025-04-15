package uuid

import (
	"database/sql/driver"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type UUIDArray []uuid.UUID

// Scan implements the sql.Scanner interface
func (ua *UUIDArray) Scan(value interface{}) error {
	if value == nil {
		*ua = make(UUIDArray, 0)
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return errors.New("unsupported Scan, storing driver.Value type " + reflect.TypeOf(value).String() + " into type *[]uuid.UUID")
	}

	// Handle PostgreSQL array representation with curly braces
	if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
		str = str[1 : len(str)-1]
	}

	// Handle empty string
	if str == "" {
		*ua = make(UUIDArray, 0)
		return nil
	}

	// Count the number of UUIDs in the input string
	count := strings.Count(str, ",") + 1

	// Pre-allocate the slice
	uuids := make([]uuid.UUID, 0, count)

	// Parse the PostgreSQL array representation into a slice of uuid.UUID
	if str != "" {
		for _, v := range strings.Split(str, ",") {
			v = strings.TrimSpace(v)
			remUuid, err := uuid.Parse(v)
			if err != nil {
				return err
			}
			uuids = append(uuids, remUuid)
		}
	}

	*ua = uuids
	return nil
}

// Value implements the driver.Valuer interface for converting Golang arrays to PostgreSQL arrays.
func (ua UUIDArray) Value() (driver.Value, error) {
	// Handle the case when the array is empty
	if len(ua) == 0 {
		return "{}", nil
	}

	var builder strings.Builder
	builder.WriteString("{")
	for i, u := range ua {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(`"`)
		builder.WriteString(u.String())
		builder.WriteString(`"`)
	}
	builder.WriteString("}")

	return builder.String(), nil
}

// GormDataType defines the common data type in GORM
func (UUIDArray) GormDataType() string {
	return "uuid[]"
}

// GormDBDataType defines the database data type for different databases
func (UUIDArray) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "uuid[]"
	default:
		return ""
	}
}
