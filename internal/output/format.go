package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"text/tabwriter"
)

// Format specifies how output data should be rendered.
type Format string

const (
	JSON  Format = "json"
	Text  Format = "text"
	Table Format = "table"
)

// ParseFormat converts a string to a Format, returning an error for unrecognized values.
func ParseFormat(s string) (Format, error) {
	switch s {
	case "json":
		return JSON, nil
	case "text":
		return Text, nil
	case "table":
		return Table, nil
	default:
		return "", fmt.Errorf("unknown format %q: must be json, text, or table", s)
	}
}

// Print renders data to the writer in the specified format.
func Print(w io.Writer, data any, format Format) error {
	switch format {
	case JSON:
		return printJSON(w, data)
	case Text:
		return printText(w, data)
	case Table:
		return printTable(w, data)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func printJSON(w io.Writer, data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

func printText(w io.Writer, data any) error {
	v := reflect.ValueOf(data)

	switch v.Kind() {
	case reflect.Map:
		return printTextMap(w, v)
	case reflect.Struct:
		return printTextStruct(w, v)
	case reflect.Slice:
		return printTextSlice(w, v)
	default:
		_, err := fmt.Fprintln(w, data)
		return err
	}
}

func printTextMap(w io.Writer, v reflect.Value) error {
	keys := make([]string, 0, v.Len())
	for _, k := range v.MapKeys() {
		keys = append(keys, k.String())
	}
	sort.Strings(keys)

	for _, k := range keys {
		val := v.MapIndex(reflect.ValueOf(k))
		if _, err := fmt.Fprintf(w, "%s: %v\n", k, val.Interface()); err != nil {
			return err
		}
	}
	return nil
}

func printTextStruct(w io.Writer, v reflect.Value) error {
	t := v.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		if _, err := fmt.Fprintf(w, "%s: %v\n", field.Name, v.Field(i).Interface()); err != nil {
			return err
		}
	}
	return nil
}

func printTextSlice(w io.Writer, v reflect.Value) error {
	for i := range v.Len() {
		if _, err := fmt.Fprintln(w, v.Index(i).Interface()); err != nil {
			return err
		}
	}
	return nil
}

func printTable(w io.Writer, data any) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice || v.Len() == 0 {
		return fmt.Errorf("table format requires a non-empty slice")
	}

	first := v.Index(0)
	switch first.Kind() {
	case reflect.Struct:
		return printTableStructs(w, v)
	case reflect.Map:
		return printTableMaps(w, v)
	default:
		return fmt.Errorf("table format requires a slice of structs or maps")
	}
}

func printTableStructs(w io.Writer, v reflect.Value) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	t := v.Index(0).Type()
	headers := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		f := t.Field(i)
		if f.IsExported() {
			headers = append(headers, f.Name)
		}
	}

	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, h)
	}
	fmt.Fprintln(tw)

	for i := range v.Len() {
		elem := v.Index(i)
		col := 0
		for j := range t.NumField() {
			if !t.Field(j).IsExported() {
				continue
			}
			if col > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, elem.Field(j).Interface())
			col++
		}
		fmt.Fprintln(tw)
	}

	return tw.Flush()
}

func printTableMaps(w io.Writer, v reflect.Value) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Collect headers from the first map's keys
	firstMap := v.Index(0)
	headers := make([]string, 0, firstMap.Len())
	for _, k := range firstMap.MapKeys() {
		headers = append(headers, k.String())
	}
	sort.Strings(headers)

	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, h)
	}
	fmt.Fprintln(tw)

	for i := range v.Len() {
		elem := v.Index(i)
		for j, h := range headers {
			if j > 0 {
				fmt.Fprint(tw, "\t")
			}
			val := elem.MapIndex(reflect.ValueOf(h))
			if val.IsValid() {
				fmt.Fprint(tw, val.Interface())
			}
		}
		fmt.Fprintln(tw)
	}

	return tw.Flush()
}
