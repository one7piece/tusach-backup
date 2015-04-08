package test

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type MyInt int

type User struct {
	Name       string
	Role       string
	Password   string
	UpdateTime time.Time
}

func TestReflect(t *testing.T) {
	var x int = 3
	var y MyInt = 4
	var z interface{}
	z = y

	var vox = reflect.ValueOf(x)
	var voy = reflect.ValueOf(y)
	var voz = reflect.ValueOf(z)

	fmt.Printf("TypeOf(x):%v, ValueOf(x):%v, Type:%v, Kind:%v, value:%v\n", reflect.TypeOf(x), vox, vox.Type(), vox.Kind(), vox.Int())
	fmt.Printf("TypeOf(y):%v, ValueOf(y):%v, Type:%v, Kind:%v, value:%v\n", reflect.TypeOf(y), voy, voy.Type(), voy.Kind(), voy.Int())
	fmt.Printf("TypeOf(z):%v, ValueOf(z):%v, Type:%v, Kind:%v, value:%v\n", reflect.TypeOf(z), voz, voz.Type(), voz.Kind(), voz.Int())

	// print the value of the reflection object
	fmt.Printf("interface(x): %v\n", vox.Interface())

	// set value of x to 333
	p := reflect.ValueOf(&x)
	fmt.Printf("type of p:%v, settability of p:%v\n", p.Type(), p.CanSet())
	v := p.Elem() // de-reference p
	fmt.Printf("type of v:%v, settability of v:%v\n", v.Type(), v.CanSet())
	v.SetInt(333)
	fmt.Println("new x:", x)

	// create new instance of type int
	p = reflect.New(reflect.TypeOf(x))
	fmt.Printf("p_x2: %v\n", p)
	v = p.Elem()
	fmt.Printf("p_x2.Elem(): %v\n", v)
	v.SetInt(3333)
	fmt.Printf("x2: %v\n", v.Interface())

	// create new instance of type MyInt
	p = reflect.New(reflect.TypeOf(y))
	fmt.Printf("p_y2: %v\n", p)
	v = p.Elem()
	fmt.Printf("p_y2.Elem(): %v\n", v)
	v.SetInt(4444)
	fmt.Printf("y2: %v\n", v.Interface())

	p = reflect.New(reflect.TypeOf(User{}))
	v = p.Elem()
	v.FieldByName("Name").SetString("dvan")
	v.FieldByName("Password").SetString("pwd")

	now := time.Now()
	var t0 interface{}
	t0 = now
	v.FieldByName("UpdateTime").Set(reflect.ValueOf(t0))

	fmt.Printf("user type: %v\n", v.Type())
	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("user field: %s=%v\n", v.Type().Field(i).Name, v.Field(i).Interface())
	}
}

func TestFieldConversion(t *testing.T) {
	user := User{Name: "dvan", Role: "admin", UpdateTime: time.Now()}
	vo := reflect.ValueOf(user)
	field, _ := vo.Type().FieldByName("Name")

	v := field2db(field.Type, vo.FieldByName("Name").Interface())
	fmt.Printf("field2db: %T(%v)\n", v, v)

	field, _ = vo.Type().FieldByName("UpdateTime")
	v = field2db(field.Type, vo.FieldByName("UpdateTime").Interface())
	fmt.Printf("field2db: %T(%v)\n", v, v)

	str, _ := fromDateTime(time.Now())
	v = db2field(field.Type, str)
	fmt.Printf("db2field: %T(%v)\n", v, v)
}

func field2db(fieldType reflect.Type, fieldValue interface{}) interface{} {
	var result interface{}
	if fieldType == reflect.TypeOf(time.Time{}) {
		result, _ = fromDateTime(fieldValue.(time.Time))
	} else if fieldType.Kind() == reflect.Int {
		result = fieldValue.(int)
	} else if fieldType.Kind() == reflect.Bool {
		if fieldValue.(bool) {
			result = 1
		} else {
			result = 0
		}
	} else {
		result = fieldValue.(string)
	}
	return result
}

func db2field(fieldType reflect.Type, dbValue interface{}) interface{} {
	var result interface{}
	if fieldType == reflect.TypeOf(time.Time{}) {
		result, _ = toDateTime(dbValue.(string))
	} else if fieldType.Kind() == reflect.Int {
		result = dbValue.(int)
	} else if fieldType.Kind() == reflect.Bool {
		if dbValue.(int) == 0 {
			result = false
		} else {
			result = true
		}
	} else {
		result = dbValue.(string)
	}
	return result
}

func TestDateTime(t *testing.T) {
	now := time.Now()
	str, _ := fromDateTime(now)
	fmt.Printf("time str: %s\n", str)
	dt, _ := toDateTime(str)
	fmt.Printf("time obj: %v\n", dt)
}

func toDateTime(str string) (time.Time, error) {
	return time.Parse(time.RFC3339, str)
}

func fromDateTime(t time.Time) (string, error) {
	buffer, err := t.MarshalText()
	return string(buffer), err
}
