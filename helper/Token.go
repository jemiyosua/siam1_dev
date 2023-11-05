package helper

import (
	"encoding/base64"
	"strconv"
	"time"
)

func Token() string {
	t := time.Now()
	tUnixMicro := strconv.FormatInt(int64(time.Nanosecond)*t.UnixNano()/int64(time.Microsecond), 10)
	token00 := base64.StdEncoding.EncodeToString([]byte(tUnixMicro))

	token01, err := GenerateRandomString(10)
	if err != nil {
		panic(err)
	}

	token02, err := GenerateRandomString(10)
	if err != nil {
		panic(err)
	}

	token03, err := GenerateRandomString(10)
	if err != nil {
		panic(err)
	}

	token := token00 + "" + token01 + "" + token02 + "" + token03
	return token
}
