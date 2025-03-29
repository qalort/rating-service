package validator

import (
        "fmt"
        "strings"

        "github.com/go-playground/validator/v10"
)

// FormatValidationErrors formats validation errors into a readable format
func FormatValidationErrors(err error) string {
        if validationErrs, ok := err.(validator.ValidationErrors); ok {
                var errMessages []string
                
                for _, fieldError := range validationErrs {
                        fieldName := strings.ToLower(fieldError.Field())
                        
                        switch fieldError.Tag() {
                        case "required":
                                errMessages = append(errMessages, fmt.Sprintf("%s is required", fieldName))
                        case "min":
                                errMessages = append(errMessages, fmt.Sprintf("%s must be at least %s", fieldName, fieldError.Param()))
                        case "max":
                                errMessages = append(errMessages, fmt.Sprintf("%s must not exceed %s", fieldName, fieldError.Param()))
                        case "uuid4":
                                errMessages = append(errMessages, fmt.Sprintf("%s must be a valid UUID", fieldName))
                        default:
                                errMessages = append(errMessages, fmt.Sprintf("%s failed validation: %s", fieldName, fieldError.Tag()))
                        }
                }
                
                return strings.Join(errMessages, "; ")
        }
        
        return err.Error()
}
