package common

import (
	"fmt"
	"io"
)

// CloseBody logs any error when closing an io.Closer (e.g., http.Response.Body).
func CloseBody(c io.Closer) {
	err := c.Close()
	if err != nil {
		fmt.Println("error closing response body:", err)
	}
}
