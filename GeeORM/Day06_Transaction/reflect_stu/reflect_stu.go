package benchmark

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Config struct {
	Name    string `json:"server-name"` // CONFIG_SERVER_NAME
	IP      string `json:"server-ip"`   // CONFIG_SERVER_IP
	URL     string `json:"server-url"`  // CONFIG_SERVER_URL
	Timeout string `json:"timeout"`     // CONFIG_TIMEOUT
}

func readConfig() *Config {
	// read from xxx.json
	config := Config{}
	typ := reflect.TypeOf(config)
	// Indirect 返回持有v持有的指针指向的值的Value。如果v持有nil指针，会返回Value零值；如果v不持有指针，会返回v。
	value := reflect.Indirect(reflect.ValueOf(&config))
	// 返回v持有的结构体类型值的字段数，如果v的Kind不是Struct会panic
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if val, ok := field.Tag.Lookup("json"); ok {
			key := fmt.Sprintf("CONFIG_%s", strings.ReplaceAll(strings.ToUpper(val), "-", "_"))
			if env, exist := os.LookupEnv(key); exist {
				value.FieldByName(field.Name).Set(reflect.ValueOf(env))
			}
		}
	}
	return &config
}

func main() {
	os.Setenv("CONFIG_SERVER_NAME", "global_server")
	os.Setenv("CONFIG_SERVER_IP", "10.0.0.1")
	os.Setenv("CONFIG_SERVER_URL", "geektutu.com")

	config := readConfig()
	fmt.Printf("config: %v", config)
}

// Main function
func Indirect_stu() {
	val01:= []int  {1, 2, 3, 4}

	var val02 reflect.Value = reflect.ValueOf(&val01)
	fmt.Println("&val2 type:", val02.Kind())

	// using the function
	indirectI1:= reflect.Indirect(val02)
	fmt.Println("indirectI  type:", indirectI1.Kind())
	fmt.Println("indirectI  value:", indirectI1)

	var val1 []int
	var val2 [3]string
	var val3 = make(map[int]string)

	var val4 reflect.Value = reflect.ValueOf(&val1)
	indirectI := reflect.Indirect(val4)
	fmt.Println("val1:", indirectI.Kind())


	var val5 reflect.Value = reflect.ValueOf(&val2)
	indirectStr:= reflect.Indirect(val5)
	fmt.Println("val2 type:", indirectStr.Kind())

	var val6 reflect.Value = reflect.ValueOf(&val3)
	fmt.Println("&val3 type:", val6.Kind())

	indirectM:= reflect.Indirect(val6)
	fmt.Println("val3 type:", indirectM.Kind())

	config := Config{}
	value := reflect.Indirect(reflect.ValueOf(&config))
	fmt.Println("config type:", value.Kind())
}