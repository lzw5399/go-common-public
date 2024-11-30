package fhttp

import (
    "fmt"
    "sync"

    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/locales/en"
    ut "github.com/go-playground/universal-translator"
    "github.com/go-playground/validator/v10"
    enTranslations "github.com/go-playground/validator/v10/translations/en"
)

var defaultTranslator ut.Translator
var _once sync.Once

func InitValidator() {
    // _once.Do(func() {
    //	if err := initTrans(); err != nil {
    //		panic(err)
    //	}
    // })
}

func DefaultTranslator() ut.Translator {
    return defaultTranslator
}

// initTrans 初始化翻译器
func initTrans() (err error) {
    // 修改gin框架中的Validator引擎属性，实现自定制
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {

        uni := ut.New(en.New())

        // locale 通常取决于 http 请求头的 'Accept-Language'
        var ok bool
        // 也可以使用 uni.FindTranslator(...) 传入多个locale进行查找
        defaultTranslator, ok = uni.GetTranslator("en")
        if !ok {
            return fmt.Errorf("uni.GetTranslator(en) failed")
        }

        // 注册翻译器
        err = enTranslations.RegisterDefaultTranslations(v, defaultTranslator)
        return
    }
    return
}
