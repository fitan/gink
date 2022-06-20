package http

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"reflect"
)

func bindGinKey(ctx *gin.Context, iV reflect.Value) error {
	iV = iV.Elem()
	for i := 0; i < iV.Type().NumField(); i++ {
		tf := iV.Type().Field(i)

		if tf.Anonymous {
			var nestValue reflect.Value

			if iV.Field(i).Type().Kind() != reflect.Ptr {
				nestValue = iV.Field(i).Addr()
			} else {
				if iV.Field(i).IsZero() {
					nestValue = reflect.New(iV.Field(i).Type().Elem())
					iV.Field(i).Set(nestValue)
				} else {
					nestValue = iV.Field(i)
				}
				//} else {
				//	nestValue = reflect.NewAt(iV.Field(i).Type(), unsafe.Pointer(iV.Field(i).UnsafeAddr())).Elem()
				//	iV.Field(i).Set(nestValue)
				//}
			}

			err := bindGinKey(ctx, nestValue)
			if err != nil {
				return err
			}
			continue
		}

		tagV, ok := tf.Tag.Lookup("ginkey")
		if !ok {
			continue
		}

		value, ok := ctx.Get(tagV)

		if ok {
			valueV := reflect.ValueOf(value)
			if iV.Field(i).Type() != valueV.Type() {
				return fmt.Errorf("ctxkey field %v type not equal to %v", iV.Field(i).Type().Name(), valueV.Type().String())
			}
			iV.Field(i).Set(valueV)
		}
	}
	return nil
}

func BindGinKey(ctx *gin.Context, i interface{}) error {
	iV := reflect.ValueOf(i)
	if iV.Type().Kind() != reflect.Ptr {
		return errors.New("binding value not ptr")
	}

	return bindGinKey(ctx, iV)
}
