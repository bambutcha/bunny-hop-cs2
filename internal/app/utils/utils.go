// To future
package utils

import (
	"github.com/0xrawsec/golang-win32/win32"
)

func SetConsoleTittle(title string) {
	win32.SetConsoleTittle(title)
}