/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package dataresponse

import (
	"strings"

	"github.com/raoptimus/data-response.go/v2/response"
)

type (
	TemplateError struct {
		Pointer string `json:"pointer,omitempty"` // Путь до свойства с проблемой
		NodeID  string `json:"nodeId,omitempty"`  // ID узла(uuid) в котором возникла ошибка
		PortID  string `json:"portId,omitempty"`  // ID порта узла(uuid) в котором возникла ошибка
		Detail  string `json:"detail"`            // Человеко-читаемое описание ошибки
	}
	TemplateErrors []TemplateError
	Template       struct {
		Code    response.HTTPCode `json:"code,omitempty"`
		Status  string            `json:"status,omitempty"`
		Title   string            `json:"title,omitempty"`
		Details any               `json:"details,omitempty"`
		Errors  TemplateErrors    `json:"errors,omitempty"`
	}
	InternalError struct {
		Error string `json:"error"`
		StackTrace string `json:"stackTrace,omitempty"`
	}
)

func (ie *InternalError) String() string  {
	var sb strings.Builder
	sb.WriteString("Error:\n")
	sb.WriteString(ie.Error)

	if len(ie.StackTrace) > 0 {
		sb.WriteString("\nStack Trace:")
		sb.WriteString(ie.StackTrace)
		sb.WriteString("\n\n")
	}

	return sb.String()
}
