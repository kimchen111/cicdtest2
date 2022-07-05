package common

import (
	"reflect"
)

func relectFields(entry any) map[string]string {
	t := reflect.TypeOf(entry)
	mapx := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i)
		mapx[key.Name] = key.Name
	}
	return mapx
}

func resetConst(entry any, mapx map[string]string) {
	for k, v := range mapx {
		vo := reflect.ValueOf(entry).Elem()
		if vo.FieldByName(k).CanSet() {
			vo.FieldByName(k).SetString(v)
		}
	}
}
