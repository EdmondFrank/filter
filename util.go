package filter

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm/schema"
	"goyave.dev/goyave/v4/util/sliceutil"
)

// DataType is determined by the `filter_type` struct tag (see `DataType` for available options).
// If not given, uses GORM's general DataType. Raw database data types are not supported so it is
// recommended to always specify a `filter_type` in this scenario.
type DataType string

// Supported DataTypes
const (
	DataTypeText        DataType = "text"
	DataTypeBool        DataType = "bool"
	DataTypeInt         DataType = "int"
	DataTypeUint        DataType = "uint"
	DataTypeFloat       DataType = "float"
	DataTypeTime        DataType = "time"
	DataTypeUnsupported DataType = "unsupported"
)

func cleanColumns(sch *schema.Schema, columns []string, blacklist []string) []*schema.Field {
	fields := make([]*schema.Field, 0, len(columns))
	for _, c := range columns {
		f, ok := sch.FieldsByDBName[c]
		if ok && !sliceutil.ContainsStr(blacklist, c) {
			fields = append(fields, f)
		}
	}

	return fields
}

func addPrimaryKeys(schema *schema.Schema, fields []string) []string {
	for _, k := range schema.PrimaryFieldDBNames {
		if !sliceutil.ContainsStr(fields, k) {
			fields = append(fields, k)
		}
	}
	return fields
}

func addForeignKeys(sch *schema.Schema, fields []string) []string {
	for _, r := range sch.Relationships.Relations {
		if r.Type == schema.HasOne || r.Type == schema.BelongsTo {
			for _, ref := range r.References {
				if !sliceutil.ContainsStr(fields, ref.ForeignKey.DBName) {
					fields = append(fields, ref.ForeignKey.DBName)
				}
			}
		}
	}
	return fields
}

func columnsContain(fields []*schema.Field, field *schema.Field) bool {
	for _, f := range fields {
		if f.DBName == field.DBName {
			return true
		}
	}
	return false
}

func getDataType(field *schema.Field) DataType {
	fromTag := DataType(strings.ToLower(field.Tag.Get("filter_type")))
	switch fromTag {
	case DataTypeText, DataTypeBool, DataTypeFloat, DataTypeInt, DataTypeUint, DataTypeTime:
		return fromTag
	default:
		switch field.DataType {
		case schema.String:
			return DataTypeText
		case schema.Bool:
			return DataTypeBool
		case schema.Float:
			return DataTypeFloat
		case schema.Int:
			return DataTypeInt
		case schema.Uint:
			return DataTypeUint
		case schema.Time:
			return DataTypeTime
		}
	}
	return DataTypeUnsupported
}

// ConvertToSafeType convert the string argument to a safe type that
// matches the column's data type. Returns false if the input could not
// be converted.
func ConvertToSafeType(arg string, dataType DataType) (interface{}, bool) { // TODO test this + test when datatype doesn't match
	switch dataType {
	case DataTypeText:
		return arg, true
	case DataTypeBool:
		switch arg {
		case "1", "on", "true", "yes":
			return true, true
		case "0", "off", "false", "no":
			return false, true
		}
		return nil, false
	case DataTypeFloat:
		i, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return nil, false
		}
		return i, true
	case DataTypeInt:
		i, err := strconv.ParseInt(arg, 10, 64) // TODO check it works on smallint
		if err != nil {
			return nil, false
		}
		return i, true
	case DataTypeUint:
		i, err := strconv.ParseUint(arg, 10, 64)
		if err != nil {
			return nil, false
		}
		return i, true
	case DataTypeTime:
		if validateTime(arg) {
			return arg, true
		}
	}
	return nil, false
}

func validateTime(timeStr string) bool {
	for _, format := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		_, err := time.Parse(format, timeStr)
		if err == nil {
			return true
		}
	}

	return false
}

// ConvertArgsToSafeType converts a slice of string arguments to safe type
// that matches the column's data type in the same way as `ConvertToSafeType`.
// If any of the values in the given slice could not be converted, returns false.
func ConvertArgsToSafeType(args []string, dataType DataType) ([]interface{}, bool) {
	result := make([]interface{}, 0, len(args))
	for _, arg := range args {
		a, ok := ConvertToSafeType(arg, dataType)
		if !ok {
			return nil, false
		}
		result = append(result, a)
	}
	return result, true
}
