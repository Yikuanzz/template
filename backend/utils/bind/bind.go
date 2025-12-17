package bind

import (
	"errors"
	"fmt"
	"strings"

	"backend/utils/errorx"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// FieldErrorConfig 字段错误配置
// 用于配置字段名到错误码的映射
type FieldErrorConfig struct {
	// InvalidParamCode 通用参数无效错误码
	InvalidParamCode int32
	// RequiredCode 参数必填错误码
	RequiredCode int32
	// FieldErrorCodes 字段名到错误码的映射
	// key: 字段名, value: 该字段的格式错误码
	FieldErrorCodes map[string]int32
	// FieldLabels 字段名到中文标签的映射
	// key: 字段名, value: 中文标签
	FieldLabels map[string]string
}

// HandleBindingError 处理 gin binding 验证错误，转换为 errorx 错误
// config: 错误码配置
// err: gin binding 返回的错误
func HandleBindingError(config FieldErrorConfig, err error) error {
	if err == nil {
		return nil
	}

	// 检查是否是 validator.ValidationErrors 类型
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		// 如果不是验证错误，返回通用参数错误
		if config.InvalidParamCode > 0 {
			return errorx.New(config.InvalidParamCode, errorx.K("reason", err.Error()))
		}
		return errorx.New(0, err.Error())
	}

	// 处理第一个验证错误
	if len(validationErrors) == 0 {
		if config.InvalidParamCode > 0 {
			return errorx.New(config.InvalidParamCode, errorx.K("reason", "参数验证失败"))
		}
		return errorx.New(0, "参数验证失败")
	}

	firstErr := validationErrors[0]
	fieldName := firstErr.Field()
	tag := firstErr.Tag()

	// 获取字段的中文标签
	fieldLabel := getFieldLabel(config, fieldName)

	// 如果是 required 错误
	if tag == "required" {
		if config.RequiredCode > 0 {
			return errorx.New(config.RequiredCode, errorx.K("param", fieldLabel))
		}
		if config.InvalidParamCode > 0 {
			return errorx.New(config.InvalidParamCode, errorx.K("reason", fmt.Sprintf("%s不能为空", fieldLabel)))
		}
		return errorx.New(0, fmt.Sprintf("%s不能为空", fieldLabel))
	}

	// 查找字段对应的错误码
	if errorCode, ok := config.FieldErrorCodes[fieldName]; ok && errorCode > 0 {
		fieldValue := getFieldValue(firstErr)
		// 将字段名转换为小写，以匹配错误消息模板中的占位符格式（如 {email}, {mobile}）
		paramKey := strings.ToLower(fieldName)
		return errorx.New(errorCode, errorx.K(paramKey, fieldValue))
	}

	// 通用错误处理
	reason := fmt.Sprintf("%s字段验证失败: %s", fieldLabel, getValidationErrorMessage(firstErr))
	if config.InvalidParamCode > 0 {
		return errorx.New(config.InvalidParamCode, errorx.K("reason", reason))
	}
	return errorx.New(0, reason)
}

// getFieldLabel 获取字段的中文标签
func getFieldLabel(config FieldErrorConfig, fieldName string) string {
	if config.FieldLabels != nil {
		if label, ok := config.FieldLabels[fieldName]; ok {
			return label
		}
	}
	return fieldName
}

// getFieldValue 获取字段的值（用于错误消息）
func getFieldValue(fe validator.FieldError) string {
	if fe.Value() != nil {
		return fmt.Sprintf("%v", fe.Value())
	}
	return ""
}

// getValidationErrorMessage 获取验证错误消息
func getValidationErrorMessage(fe validator.FieldError) string {
	fieldName := fe.Field()
	tag := fe.Tag()

	// 根据标签返回友好的错误消息
	switch tag {
	case "required":
		return fmt.Sprintf("%s不能为空", fieldName)
	case "email":
		return "邮箱格式不正确"
	case "len":
		return fmt.Sprintf("%s长度必须为%s", fieldName, fe.Param())
	case "min":
		return fmt.Sprintf("%s长度不能少于%s", fieldName, fe.Param())
	case "max":
		return fmt.Sprintf("%s长度不能超过%s", fieldName, fe.Param())
	case "gte":
		return fmt.Sprintf("%s必须大于等于%s", fieldName, fe.Param())
	case "lte":
		return fmt.Sprintf("%s必须小于等于%s", fieldName, fe.Param())
	case "gt":
		return fmt.Sprintf("%s必须大于%s", fieldName, fe.Param())
	case "lt":
		return fmt.Sprintf("%s必须小于%s", fieldName, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s必须是以下值之一: %s", fieldName, fe.Param())
	case "regexp":
		return fmt.Sprintf("%s格式不正确", fieldName)
	default:
		return fmt.Sprintf("%s验证失败: %s", fieldName, tag)
	}
}

// ShouldBindJSON 绑定并验证 JSON 请求体
// 如果验证失败，返回 errorx 错误
func ShouldBindJSON(c *gin.Context, obj interface{}, config FieldErrorConfig) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return HandleBindingError(config, err)
	}
	return nil
}

// ShouldBindQuery 绑定并验证 Query 参数
// 如果验证失败，返回 errorx 错误
func ShouldBindQuery(c *gin.Context, obj interface{}, config FieldErrorConfig) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return HandleBindingError(config, err)
	}
	return nil
}

// ShouldBindURI 绑定并验证 URI 参数
// 如果验证失败，返回 errorx 错误
func ShouldBindURI(c *gin.Context, obj interface{}, config FieldErrorConfig) error {
	if err := c.ShouldBindUri(obj); err != nil {
		return HandleBindingError(config, err)
	}
	return nil
}

// ShouldBind 绑定并验证请求（自动识别 Content-Type）
// 如果验证失败，返回 errorx 错误
func ShouldBind(c *gin.Context, obj interface{}, config FieldErrorConfig) error {
	if err := c.ShouldBind(obj); err != nil {
		return HandleBindingError(config, err)
	}
	return nil
}
