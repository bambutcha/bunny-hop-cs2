// To future
package utils

import "golang.org/x/sys/windows"

func SetConsoleTittle(title string) {
	windows.SetConsoleTittle(title)
}